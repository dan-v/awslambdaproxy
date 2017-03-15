package main

import (
	"net"
	"os"
	"log"

	"github.com/hashicorp/yamux"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
)

type LambdaTunnelConnection struct {
	tunnelHost string
	sshUsername string
	sshSigner ssh.Signer
	conn net.Conn
	sess *yamux.Session
}

func (l *LambdaTunnelConnection) publicKeyFile() ssh.AuthMethod {
	return ssh.PublicKeys(l.sshSigner)
}

func (l *LambdaTunnelConnection) setup() {
	sshConfig := &ssh.ClientConfig{
		User: l.sshUsername,
		Auth: []ssh.AuthMethod{
			l.publicKeyFile(),
		},
	}

	tunnelConn, err := ssh.Dial("tcp", l.tunnelHost, sshConfig)
	if err != nil {
		log.Println("Failed to start SSH tunnel to: " + l.tunnelHost + ". Error: ", err)
		os.Exit(1)
	}
	log.Println("Created SSH tunnel to: " + l.tunnelHost)

	localConn, err := tunnelConn.Dial("tcp", "localhost:8081")
	if err != nil {
		log.Println("Failed to create local tunnel to localhost:8081. Error: ", err)
		os.Exit(1)
	}
	l.conn = localConn
	log.Println("Created local tunnel to localhost:8081")

	tunnelSession, err := yamux.Server(localConn, nil)
	if err != nil {
		log.Println("Failed to start session inside tunnel")
		os.Exit(1)
	}
	log.Println("Started yamux session inside tunnel")
	l.sess = tunnelSession
}

func setupLambdaTunnelConnection(tunnelHost string, sshPort string, sshUsername string,
					sshPrivateKeyFile string) (*LambdaTunnelConnection, error) {
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

	tunnel := &LambdaTunnelConnection{
		tunnelHost: tunnelHost + ":" + sshPort,
		sshUsername: sshUsername,
		sshSigner: signer,
	}
	tunnel.setup()
	return tunnel, nil
}