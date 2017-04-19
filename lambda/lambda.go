package main

import (
	"flag"
	"log"
	"os"
)

func lambdaInit(tunnelHost string, sshPort string, sshPrivateKeyFile string, sshUsername string) {
	log.Println("Starting lambdaProxyServer")
	lambdaProxyServer := startLambdaProxyServer()

	log.Println("Establishing tunnel connection to", tunnelHost)
	lambdaTunnelConnection, err := setupLambdaTunnelConnection(tunnelHost, sshPort, sshUsername, sshPrivateKeyFile)
	if err != nil {
		log.Fatal("Failed to establish connection to "+tunnelHost, err)
	}

	log.Println("Starting lambdaDataCopyManager")
	dataCopyManager := newLambdaDataCopyManager(lambdaProxyServer, lambdaTunnelConnection)
	dataCopyManager.run()
}

func main() {
	addressPtr := flag.String("address", "localhost", "IP of server to connect to")
	sshPortPtr := flag.String("ssh-port", "22", "SSH port")
	sshUsernamePtr := flag.String("ssh-user", "ubuntu", "SSH username")
	sshPrivateKeyFilePtr := flag.String("ssh-private-key", "/tmp/privatekey", "SSH private key file")

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

	lambdaInit(*addressPtr, *sshPortPtr, *sshPrivateKeyFilePtr, *sshUsernamePtr)
}
