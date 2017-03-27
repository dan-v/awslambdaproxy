package main

import (
	"log"
	"os"
	"flag"
)

const (
	proxyUnixSocket = "/tmp/lambda-proxy.socket"
)

func LambdaInit(tunnelHost string, sshPort string, sshPrivateKeyFile string, sshUsername, proxyUsername string, proxyPassword string) {
	log.Println("Starting LambdaProxyServer")
	lambdaProxyServer := startLambdaProxyServer(proxyUsername, proxyPassword)

	log.Println("Establishing tunnel connection to", tunnelHost)
	lambdaTunnelConnection, err := setupLambdaTunnelConnection(tunnelHost, sshPort, sshUsername, sshPrivateKeyFile)
	if err != nil {
		log.Fatal("Failed to establish connection to " + tunnelHost + ". Error: ", err)
	}

	log.Println("Starting LambdaDataCopyManager")
	dataCopyManager := newLambdaDataCopyManager(lambdaProxyServer, lambdaTunnelConnection)
	dataCopyManager.run()
}

func main() {
	addressPtr := flag.String("address", "localhost", "IP of server to connect to")
	sshPortPtr := flag.String("ssh-port", "22", "SSH port")
	sshUsernamePtr := flag.String("ssh-user", "ubuntu", "SSH username")
	sshPrivateKeyFilePtr := flag.String("ssh-private-key", "/tmp/privatekey", "SSH private key file")
	proxyUsernamePtr := flag.String("proxy-username", "admin", "Username for authentication")
	proxyPasswordPtr := flag.String("proxy-password", "admin", "Password for authentication")

	flag.Parse()

	if *addressPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *sshPortPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *sshPrivateKeyFilePtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *sshUsernamePtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *proxyUsernamePtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *proxyPasswordPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	LambdaInit(*addressPtr, *sshPortPtr, *sshPrivateKeyFilePtr, *sshUsernamePtr, *proxyUsernamePtr, *proxyPasswordPtr)
}

