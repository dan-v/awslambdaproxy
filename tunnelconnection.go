package awslambdaproxy

import (
	"net"
	"sync"
	"log"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/pkg/errors"
)

type TunnelConnection struct {
	conn net.Conn
	sess *yamux.Session
}

type TunnelConnectionManager struct {
	listener net.Listener
	lastTunnel TunnelConnection
	mutex sync.RWMutex
}

func (t *TunnelConnectionManager) run() {
	for {
		c, err := t.listener.Accept()
		if err != nil {
			log.Println("Failed to accept tunnel connection")
			return
		}
		log.Println("Accepted tunnel connection from", c.RemoteAddr())

		tunnelSession, err := yamux.Client(c, nil)
		if err != nil {
			log.Println("Failed to start session inside tunnel")
			return
		}
		log.Println("Established session to", tunnelSession.RemoteAddr())

		t.mutex.Lock()
		t.lastTunnel = TunnelConnection{c, tunnelSession}
		t.mutex.Unlock()
	}
}

func (t *TunnelConnectionManager) waitUntilReady() {
	for {
		if t.isReady() == true {
			break
		} else {
			log.Println("Waiting for tunnel to be established..")
			time.Sleep(time.Second * 1)
		}
	}
}

func (t *TunnelConnectionManager) isReady() bool {
	if t.lastTunnel == (TunnelConnection{}) {
		return false
	} else {
		return true
	}
}

func newTunnelConnectionManager(port string) (*TunnelConnectionManager, error) {
	listener, err := startTunnelListener(port)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to start TunnelConnectionManager")
	}
	return &TunnelConnectionManager{listener: listener}, nil
}

func startTunnelListener(tunnelPort string) (net.Listener, error) {
	tunnelAddress := "0.0.0.0:" + tunnelPort
	tunnelListener, err := net.Listen("tcp", tunnelAddress)
	if err != nil {
		errors.Wrap(err, "Failed to start TCP tunnel listener on port " + tunnelPort)
	}
	log.Println("Started TCP tunnel listener on port " + tunnelPort)
	return tunnelListener, nil
}