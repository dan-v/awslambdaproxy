package main

import (
	"log"
	"os"
	"flag"
)

const (
	proxyUnixSocket = "/tmp/lambda-proxy.socket"
)

func LambdaInit(tunnelHost string, proxyType string) {
	log.Println("Starting LambdaProxyServer")
	lambdaProxyServer := startLambdaProxyServer(proxyType)

	log.Println("Establishing tunnel connection to", tunnelHost)
	lambdaTunnelConnection := setupLambdaTunnelConnection(tunnelHost)

	log.Println("Starting LambdaDataCopyManager")
	dataCopyManager := newLambdaDataCopyManager(lambdaProxyServer, lambdaTunnelConnection)
	dataCopyManager.run()
}

func main() {
	addressPtr := flag.String("address", "localhost:8081", "IP and port of server to connect to")
	proxyTypePtr := flag.String("proxy-type", "http", "Proxy type to setup: 'http' or 'socks'")

	flag.Parse()

	if *addressPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *proxyTypePtr != "http" && *proxyTypePtr != "socks" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	LambdaInit(*addressPtr, *proxyTypePtr)

}

