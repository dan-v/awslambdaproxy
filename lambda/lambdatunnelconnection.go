package main

import (
	"io/ioutil"
	"log"
	"net"

	"github.com/hashicorp/yamux"
	"golang.org/x/crypto/ssh"
)

const (
	tunnelPortOnRemoteServer = "localhost:8081"
)

type lambdaTunnelConnection struct {
	tunnelHost  string
	sshUsername string
	sshSigner   ssh.Signer
	conn        net.Conn
	sess        *yamux.Session
}

func (l *lambdaTunnelConnection) publicKeyFile() ssh.AuthMethod {
	return ssh.PublicKeys(l.sshSigner)
}

func (l *lambdaTunnelConnection) setup() {
	sshConfig := &ssh.ClientConfig{
		User:            l.sshUsername,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			l.publicKeyFile(),
		},
	}

	tunnelConn, err := ssh.Dial("tcp", l.tunnelHost, sshConfig)
	if err != nil {
		log.Fatalf("Failed to start SSH tunnel to %v: %v\n", l.tunnelHost, err)
	}
	log.Printf("Setup SSH tunnel to tunnelHost=%v\n", l.tunnelHost)

	localConn, err := tunnelConn.Dial("tcp", tunnelPortOnRemoteServer)
	if err != nil {
		log.Fatalf("Failed to create connection to tunnelPortOnRemoteServer=%v: %v\n", tunnelPortOnRemoteServer, err)
	}
	l.conn = localConn
	log.Printf("Setup connection to tunnelPortOnRemoteServer=%v\n", tunnelPortOnRemoteServer)

	tunnelSession, err := yamux.Server(localConn, nil)
	if err != nil {
		log.Fatalf("Failed to start session inside tunnel: %v\n", err)
	}
	log.Println("Started yamux session inside tunnel")
	l.sess = tunnelSession
}

func (l *lambdaTunnelConnection) close() {
	log.Printf("Closing session")
	err := l.sess.Close()
	if err != nil {
		log.Printf("Error closing session: %v", err)
	}
	log.Printf("Closing connection")
	err = l.conn.Close()
	if err != nil {
		log.Printf("Error closing connection: %v", err)
	}
}

func setupLambdaTunnelConnection(tunnelHost string, sshPort string, sshUsername string,
	sshPrivateKeyFile string) (*lambdaTunnelConnection, error) {
	data, err := ioutil.ReadFile(sshPrivateKeyFile)
	if err != nil {
		log.Println("Failed to read private key file", sshPrivateKeyFile)
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(data)
	if err != nil {
		log.Println("Failed to parse private key", sshPrivateKeyFile)
		return nil, err
	}

	tunnel := &lambdaTunnelConnection{
		tunnelHost:  tunnelHost + ":" + sshPort,
		sshUsername: sshUsername,
		sshSigner:   signer,
	}

	tunnel.setup()
	return tunnel, nil
}
