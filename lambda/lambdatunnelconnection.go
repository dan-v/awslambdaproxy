package main

import (
	"net"
	"os"
	"log"

	"github.com/hashicorp/yamux"
)

type LambdaTunnelConnection struct {
	tunnelHost string
	conn net.Conn
	sess *yamux.Session
}

func (l *LambdaTunnelConnection) setup() {
	tunnelConn, err := net.Dial("tcp", l.tunnelHost)
	if err != nil {
		log.Println("Failed to start tunnel to: ", l.tunnelHost)
		os.Exit(1)
	}
	log.Println("Created tunnel to: " + l.tunnelHost)
	l.conn = tunnelConn

	tunnelSession, err := yamux.Server(tunnelConn, nil)
	if err != nil {
		log.Println("Failed to start session inside tunnel")
		os.Exit(1)
	}
	log.Println("Started yamux session inside tunnel")
	l.sess = tunnelSession
}

func setupLambdaTunnelConnection(tunnelHost string) *LambdaTunnelConnection {
	ltc := &LambdaTunnelConnection{
		tunnelHost: tunnelHost,
	}
	ltc.setup()
	return ltc
}