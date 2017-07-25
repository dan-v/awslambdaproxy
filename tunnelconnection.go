package awslambdaproxy

import (
	"log"
	"net"
	"sync"
	"time"

	"strconv"

	"github.com/hashicorp/yamux"
	"github.com/pkg/errors"
)

const (
	maxTunnels  = 10
	forwardPort = "8082"
	tunnelPort  = "8081"
)

type tunnelConnection struct {
	conn    net.Conn
	sess    *yamux.Session
	streams map[uint32]*yamux.Stream
	time    time.Time
}

type connectionManager struct {
	forwardListener       net.Listener
	tunnelListener        net.Listener
	tunnelConnections     map[string]tunnelConnection
	tunnelMutex           sync.RWMutex
	tunnelExpectedRuntime float64
	tunnelRedeployNeeded  chan bool
	activeTunnel          string
	localProxy            *LocalProxy
}

func (t *connectionManager) runForwarder() {
	t.waitUntilTunnelIsAvailable()
	for {
		c, err := t.forwardListener.Accept()
		if err != nil {
			log.Println("Failed to accept user connection")
			return
		}
		go t.handleForwardConnection(c)
	}
}

func (t *connectionManager) handleForwardConnection(localProxyConn net.Conn) {
	tunnelStream, err := t.openNewStreamInActiveTunnel()
	if err != nil {
		log.Println("Failed to open new stream in active tunnel", err)
		return
	}

	bidirectionalCopy(localProxyConn, tunnelStream)
}

func (t *connectionManager) runTunnel() {
	for {
		if len(t.tunnelConnections) > maxTunnels {
			log.Println("Too many active tunnelConnections: " + string(len(t.tunnelConnections)) + ". MAX=" +
				string(maxTunnels) + ". Waiting for cleanup.")
			time.Sleep(time.Second * 5)
			continue
		}
		c, err := t.tunnelListener.Accept()
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

		t.tunnelMutex.Lock()
		t.activeTunnel = c.RemoteAddr().String()
		t.tunnelConnections[t.activeTunnel] = tunnelConnection{
			conn:    c,
			sess:    tunnelSession,
			streams: make(map[uint32]*yamux.Stream),
			time:    time.Now(),
		}

		t.tunnelMutex.Unlock()
		go t.monitorTunnelSessionHealth(t.activeTunnel)
		log.Println("Active tunnel count: ", len(t.tunnelConnections))
		for k, v := range t.tunnelConnections {
			log.Println("---------------")
			log.Println("Connection: " + k)
			log.Println("Start Time: " + v.time.String())
			log.Println("Total Streams: " + strconv.Itoa(v.sess.NumStreams()))
			log.Println("---------------")
		}
	}
}

func (t *connectionManager) removeTunnelConnection(connectionID string) {
	t.tunnelConnections[connectionID].sess.Close()
	t.tunnelConnections[connectionID].conn.Close()
	t.tunnelMutex.Lock()
	delete(t.tunnelConnections, connectionID)
	t.tunnelMutex.Unlock()
}

func (t *connectionManager) monitorTunnelSessionHealth(connectionID string) {
	for {
		_, err := t.tunnelConnections[connectionID].sess.Ping()
		if err != nil {
			if time.Since(t.tunnelConnections[connectionID].time).Seconds() < t.tunnelExpectedRuntime {
				log.Println("Signaling for emergency tunnel due to tunnel ending early: ", time.Since(t.tunnelConnections[connectionID].time).Seconds())
				t.tunnelRedeployNeeded <- true
			}
			t.removeTunnelConnection(connectionID)
			break
		}
		if time.Since(t.tunnelConnections[connectionID].time).Seconds() > t.tunnelExpectedRuntime {
			numStreams := t.tunnelConnections[connectionID].sess.NumStreams()
			if numStreams > 0 {
				log.Println("Tunnel " + connectionID + " that is being closed still has open streams: " + strconv.Itoa(numStreams) + ". Delaying cleanup.")
				time.Sleep(20 * time.Second)
				log.Println("Delayed cleanup now running for ", connectionID)
			} else {
				log.Println("Tunnel " + connectionID + " is safe to close")
			}
			log.Println("Removing tunnel", connectionID)
			t.removeTunnelConnection(connectionID)
			break
		}
		time.Sleep(time.Millisecond * 50)
	}
}

func (t *connectionManager) openNewStreamInActiveTunnel() (*yamux.Stream, error) {
	for {
		t.tunnelMutex.RLock()
		tunnel, ok := t.tunnelConnections[t.activeTunnel]
		t.tunnelMutex.RUnlock()
		if ok {
			stream, err := tunnel.sess.OpenStream()
			tunnel.streams[stream.StreamID()] = stream
			return stream, err
		}
		log.Println("No active tunnel session available. Retrying..")
		time.Sleep(time.Second)
	}
}

func (t *connectionManager) waitUntilTunnelIsAvailable() error {
	timeout := time.After(time.Second * time.Duration(t.tunnelExpectedRuntime))
	tick := time.Tick(time.Second)
	for {
		select {
		case <-timeout:
			return errors.New("Timed out waiting for tunnel to be established. Likely the " +
				"Lambda function is having issues communicating with this host.")
		case <-tick:
			if t.isReady() == true {
				return nil
			}
			log.Println("Waiting for tunnel to be established..")
		}
	}
}

func (t *connectionManager) isReady() bool {
	if t.activeTunnel == "" {
		return false
	}
	return true
}

func newTunnelConnectionManager(frequency time.Duration, localProxy *LocalProxy) (*connectionManager, error) {
	forwardListener, err := startForwardListener()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to start UserListener")
	}

	tunnelListener, err := startTunnelListener()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to start TunnelListener")
	}

	connectionManager := &connectionManager{
		forwardListener:       forwardListener,
		tunnelListener:        tunnelListener,
		tunnelConnections:     make(map[string]tunnelConnection),
		tunnelRedeployNeeded:  make(chan bool),
		tunnelExpectedRuntime: frequency.Seconds(),
		localProxy:            localProxy,
	}

	go connectionManager.runTunnel()
	go connectionManager.runForwarder()

	return connectionManager, nil
}

func startTunnelListener() (net.Listener, error) {
	tunnelAddress := "localhost:" + tunnelPort
	tunnelListener, err := net.Listen("tcp", tunnelAddress)

	if err != nil {
		return nil, errors.Wrap(err, "Failed to start TCP tunnel listener on port "+tunnelPort)
	}
	log.Println("Started tunnel listener on port " + tunnelPort)
	return tunnelListener, nil
}

func startForwardListener() (net.Listener, error) {
	forwardAddress := "localhost:" + forwardPort
	forwardListener, err := net.Listen("tcp", forwardAddress)

	if err != nil {
		return nil, errors.Wrap(err, "Failed to start TCP user listener on port "+forwardPort)
	}
	log.Println("Started user listener on port " + forwardPort)
	return forwardListener, nil
}
