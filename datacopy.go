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
	tunnelStream, err := d.tunnelConnectionManager.getActiveSession()
	if err != nil {
		log.Println("Failed to open stream to remote")
		return
	}

	bidirectionalCopy(tunnelStream, userConn)
}

func newDataCopyManager(user *UserConnectionManager, tunnel *TunnelConnectionManager) *DataCopyManager {
	copyManager :=  &DataCopyManager{
		userConnectionManager: user,
		tunnelConnectionManager: tunnel,
	}
	go copyManager.run()
	return copyManager
}