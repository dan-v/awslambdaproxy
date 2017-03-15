package awslambdaproxy

import (
	"net"
	"sync"
	"log"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/pkg/errors"
	"strconv"
)

const (
	maxTunnels = 10
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
		if len(t.tunnels) > maxTunnels {
			log.Println("Too many active tunnels: " + string(len(t.tunnels)) + ". MAX=" +
				string(maxTunnels) + ". Waiting for cleanup.")
			time.Sleep(time.Second * 5)
			continue
		}
		c, err := t.listener.Accept()
		if err != nil {
			log.Println("Failed to accept tunnel connection")
			time.Sleep(time.Second * 5)
			continue
		}
		log.Println("Accepted tunnel connection from", c.RemoteAddr())

		tunnelSession, err := yamux.Client(c, nil)
		if err != nil {
			log.Println("Failed to start session inside tunnel")
			time.Sleep(time.Second * 5)
			continue
		}
		log.Println("Established session to", tunnelSession.RemoteAddr())

		t.mutex.Lock()
		t.currentTunnel = c.RemoteAddr().String()
		t.tunnels[t.currentTunnel] = TunnelConnection{c, tunnelSession, time.Now()}
		t.mutex.Unlock()
		go t.monitorConnectionHealth(t.currentTunnel)
		log.Println("Active tunnel count: ", len(t.tunnels))
		for k, v := range t.tunnels {
			log.Println("---------------")
			log.Println("Connection: " + k)
			log.Println("Start Time: " + v.time.String())
			log.Println("Total Streams: " + strconv.Itoa(v.sess.NumStreams()))
			log.Println("---------------")
		}
	}
}

func (t *TunnelConnectionManager) removeConnection(connectionId string){
	t.tunnels[connectionId].sess.Close()
	t.tunnels[connectionId].conn.Close()
	t.mutex.Lock()
	delete(t.tunnels, connectionId)
	t.mutex.Unlock()
}

func (t *TunnelConnectionManager) monitorConnectionHealth(connectionId string) {
	for {
		_, err := t.tunnels[connectionId].sess.Ping()
		if err != nil {
			if time.Since(t.tunnels[connectionId].time).Seconds() < t.expectedTunnelSeconds {
				log.Println("Signaling for emergency tunnel due to tunnel ending early: ", time.Since(t.tunnels[connectionId].time).Seconds())
				t.emergencyTunnel <- true
			}
			t.removeConnection(connectionId)
			break
		}
		if time.Since(t.tunnels[connectionId].time).Seconds() > t.expectedTunnelSeconds {
			numStreams := t.tunnels[connectionId].sess.NumStreams()
			if numStreams > 0 {
				log.Println("Tunnel " + connectionId + " that is being closed still has open streams: " + strconv.Itoa(numStreams))
				time.Sleep(5 * time.Second)
			} else {
				log.Println("Tunnel " + connectionId + " is safe to close")
			}
			t.removeConnection(connectionId)
			break
		}
		time.Sleep(time.Millisecond * 50)
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

func (t *TunnelConnectionManager) waitUntilReady(timeoutDuration time.Duration) error {
	timeout := time.After(timeoutDuration)
	tick := time.Tick(time.Second)
	for {
		select {
		case <-timeout:
			return errors.New("Timed out waiting for tunnel to be established. Likely the " +
				"Lambda function is having issues communicating with this host.")
		case <-tick:
			if t.isReady() == true {
				return nil
			} else {
				log.Println("Waiting for tunnel to be established..")
			}
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

func newTunnelConnectionManager(lambdaExecutionFrequency time.Duration) (*TunnelConnectionManager, error) {
	listener, err := startTunnelListener()
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

func startTunnelListener() (net.Listener, error) {
	tunnelPort := "8081"
	tunnelAddress := "localhost:" + tunnelPort
	tunnelListener, err := net.Listen("tcp", tunnelAddress)

	if err != nil {
		return nil, errors.Wrap(err, "Failed to start TCP tunnel listener on port " + tunnelPort)
	}
	log.Println("Started TCP tunnel listener on port " + tunnelPort)
	return tunnelListener, nil
}