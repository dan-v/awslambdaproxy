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
	maxTunnels  = 10
	forwardPort = "8082"
	tunnelPort  = "8081"
)

type TunnelConnection struct {
	conn net.Conn
	sess *yamux.Session
	streams map[uint32]*yamux.Stream
	time time.Time
}

type ConnectionManager struct {
	forwardListener       net.Listener
	tunnelListener        net.Listener
	tunnelConnections     map[string]TunnelConnection
	tunnelMutex           sync.RWMutex
	tunnelExpectedRuntime float64
	tunnelRedeployNeeded  chan bool
	activeTunnel          string
	localProxy	      *LocalProxy
}

func (t *ConnectionManager) runForwarder() {
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

func (t *ConnectionManager) handleForwardConnection(localProxyConn net.Conn) {
	tunnelStream, err := t.openNewStreamInActiveTunnel()
	if err != nil {
		log.Println("Failed to open new stream in active tunnel", err)
		return
	}

	bidirectionalCopy(localProxyConn, tunnelStream)
}

func (t *ConnectionManager) runTunnel() {
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
		t.tunnelConnections[t.activeTunnel] = TunnelConnection{
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

func (t *ConnectionManager) removeTunnelConnection(connectionId string){
	t.tunnelConnections[connectionId].sess.Close()
	t.tunnelConnections[connectionId].conn.Close()
	t.tunnelMutex.Lock()
	delete(t.tunnelConnections, connectionId)
	t.tunnelMutex.Unlock()
}

func (t *ConnectionManager) monitorTunnelSessionHealth(connectionId string) {
	for {
		_, err := t.tunnelConnections[connectionId].sess.Ping()
		if err != nil {
			if time.Since(t.tunnelConnections[connectionId].time).Seconds() < t.tunnelExpectedRuntime {
				log.Println("Signaling for emergency tunnel due to tunnel ending early: ", time.Since(t.tunnelConnections[connectionId].time).Seconds())
				t.tunnelRedeployNeeded <- true
			}
			t.removeTunnelConnection(connectionId)
			break
		}
		if time.Since(t.tunnelConnections[connectionId].time).Seconds() > t.tunnelExpectedRuntime {
			numStreams := t.tunnelConnections[connectionId].sess.NumStreams()
			if numStreams > 0 {
				log.Println("Tunnel " + connectionId + " that is being closed still has open streams: " + strconv.Itoa(numStreams) + ". Delaying cleanup.")
				time.Sleep(20 * time.Second)
				log.Println("Delayed cleanup now running for ", connectionId)
			} else {
				log.Println("Tunnel " + connectionId + " is safe to close")
			}
			log.Println("Removing tunnel", connectionId)
			t.removeTunnelConnection(connectionId)
			break
		}
		time.Sleep(time.Millisecond * 50)
	}
}

func (t *ConnectionManager) openNewStreamInActiveTunnel() (*yamux.Stream, error) {
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

func (t *ConnectionManager) waitUntilTunnelIsAvailable() error {
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
			} else {
				log.Println("Waiting for tunnel to be established..")
			}
		}
	}
}

func (t *ConnectionManager) isReady() bool {
	if t.activeTunnel == "" {
		return false
	} else {
		return true
	}
}

func newTunnelConnectionManager(frequency time.Duration, localProxy *LocalProxy) (*ConnectionManager, error) {
	forwardListener, err := startForwardListener()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to start UserListener")
	}

	tunnelListener, err := startTunnelListener()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to start TunnelListener")
	}

	connectionManager := &ConnectionManager{
		forwardListener:       forwardListener,
		tunnelListener:        tunnelListener,
		tunnelConnections:     make(map[string]TunnelConnection),
		tunnelRedeployNeeded:  make(chan bool),
		tunnelExpectedRuntime: frequency.Seconds(),
		localProxy: localProxy,
	}

	go connectionManager.runTunnel()
	go connectionManager.runForwarder()

	return connectionManager, nil
}

func startTunnelListener() (net.Listener, error) {
	tunnelAddress := "localhost:" + tunnelPort
	tunnelListener, err := net.Listen("tcp", tunnelAddress)

	if err != nil {
		return nil, errors.Wrap(err, "Failed to start TCP tunnel listener on port " + tunnelPort)
	}
	log.Println("Started tunnel listener on port " + tunnelPort)
	return tunnelListener, nil
}

func startForwardListener() (net.Listener, error) {
	forwardAddress := "localhost:" + forwardPort
	forwardListener, err := net.Listen("tcp", forwardAddress)

	if err != nil {
		return nil, errors.Wrap(err, "Failed to start TCP user listener on port " + forwardPort)
	}
	log.Println("Started user listener on port " + forwardPort)
	return forwardListener, nil
}