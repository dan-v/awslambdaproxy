package main

import (
	"log"
	"net"
	"os"

	"github.com/hashicorp/yamux"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
)

const (
	remoteTunnelPort = "localhost:8081"
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
		log.Println("Failed to start SSH tunnel to: "+l.tunnelHost+". Error: ", err)
		os.Exit(1)
	}
	log.Println("Created SSH tunnel to: " + l.tunnelHost)

	localConn, err := tunnelConn.Dial("tcp", remoteTunnelPort)
	if err != nil {
		log.Println("Failed to create local tunnel to "+remoteTunnelPort, err)
		os.Exit(1)
	}
	l.conn = localConn
	log.Println("Created local tunnel to " + remoteTunnelPort)

	tunnelSession, err := yamux.Server(localConn, nil)
	if err != nil {
		log.Println("Failed to start session inside tunnel")
		os.Exit(1)
	}
	log.Println("Started yamux session inside tunnel")
	l.sess = tunnelSession
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
