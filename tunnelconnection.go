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
	time time.Time
}

type TunnelConnectionManager struct {
	listener      net.Listener
	currentTunnel string
	tunnels       map[string]TunnelConnection
	mutex         sync.RWMutex
	expectedTunnelSeconds float64
	emergencyTunnel chan bool
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
		t.currentTunnel = c.RemoteAddr().String()
		t.tunnels[t.currentTunnel] = TunnelConnection{c, tunnelSession, time.Now()}
		t.mutex.Unlock()
		go t.monitorConnectionHealth(c.RemoteAddr().String())
		log.Println("Active tunnel count: ", len(t.tunnels))
	}
}

func (t *TunnelConnectionManager) monitorConnectionHealth(connectionId string) {
	for {
		_, err := t.tunnels[connectionId].sess.Ping()
		if err != nil {
			if time.Since(t.tunnels[connectionId].time).Seconds() < t.expectedTunnelSeconds {
				log.Println("Signaling for emergency tunnel due to tunnel ending early: ", time.Since(t.tunnels[connectionId].time).Seconds())
				t.emergencyTunnel <- true
			}
			t.tunnels[connectionId].sess.Close()
			t.tunnels[connectionId].conn.Close()
			t.mutex.Lock()
			delete(t.tunnels, connectionId)
			t.mutex.Unlock()
			break
		}
		time.Sleep(time.Millisecond * 250)
	}
}

func (t *TunnelConnectionManager) getActiveSession() (net.Conn, error) {
	for {
		t.mutex.RLock()
		tunnel, ok := t.tunnels[t.currentTunnel]
		t.mutex.RUnlock()
		if ok {
			sess, err := tunnel.sess.Open()
			return sess, err
		}
		log.Println("TunnelConnectionManager.getActiveSession failed. Retrying..")
		time.Sleep(time.Second * 1)
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
	if t.currentTunnel == "" {
		return false
	} else {
		return true
	}
}

func newTunnelConnectionManager(port string, lambdaExecutionFrequency time.Duration) (*TunnelConnectionManager, error) {
	listener, err := startTunnelListener(port)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to start TunnelConnectionManager")
	}
	connectionManager := &TunnelConnectionManager{
		listener: listener,
		tunnels: make(map[string]TunnelConnection),
		emergencyTunnel: make(chan bool),
		expectedTunnelSeconds: lambdaExecutionFrequency.Seconds(),
	}
	go connectionManager.run()
	return connectionManager, nil
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