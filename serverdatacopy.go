package awslambdaproxy

import (
	"net"
	"log"
)

type DataCopyManager struct {
	userConnectionManager *UserConnectionManager
	tunnelConnectionManager *TunnelConnectionManager
}

func (d *DataCopyManager) run() {
	for {
		userConn := d.userConnectionManager.nextConnection()
		go d.handleClient(userConn)
	}
}

func (d *DataCopyManager) handleClient(userConn net.Conn) {
	d.tunnelConnectionManager.mutex.RLock()
	tunnelStream, err := d.tunnelConnectionManager.lastTunnel.sess.Open()
	d.tunnelConnectionManager.mutex.RUnlock()
	if err != nil {
		log.Println("Failed to open stream to remote")
		return
	}

	log.Println("Opened stream to remote " + tunnelStream.RemoteAddr().String())
	bidirectionalCopy(tunnelStream, userConn)
}

func newDataCopyManager(user *UserConnectionManager, tunnel *TunnelConnectionManager) *DataCopyManager {
	return &DataCopyManager{
		userConnectionManager: user,
		tunnelConnectionManager: tunnel,
	}
}