package awslambdaproxy

import (
	"net"
	"log"
	"github.com/pkg/errors"
)

type UserConnectionManager struct {
	port string
	listener net.Listener
	connections chan net.Conn
}

func (u *UserConnectionManager) run() {
	for {
		c, err := u.listener.Accept()
		if err != nil {
			log.Println("Failed to accept user connection")
			return
		}
		u.connections <- c
	}
}

func (u *UserConnectionManager) nextConnection() net.Conn {
	return <-u.connections
}

func newUserConnectionManager(port string) (*UserConnectionManager, error) {
	listener, err := startUserListener(port)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to start UserConnectionManager")
	}
	connectionManager := &UserConnectionManager{
		listener: listener,
		connections: make(chan net.Conn),
	}
	go connectionManager.run()
	return connectionManager, nil
}

func startUserListener(proxyPort string) (net.Listener, error) {
	proxyAddress := "0.0.0.0:" + proxyPort
	proxyListener, err := net.Listen("tcp", proxyAddress)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to start TCP user listener on port " + proxyPort)
	}
	log.Println("Started TCP user listener on port " + proxyPort)
	return proxyListener, nil
}