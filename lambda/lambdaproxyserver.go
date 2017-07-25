package main

import (
	"crypto/tls"
	"log"

	"github.com/dan-v/gost"
	"github.com/golang/glog"
)

type lambdaProxyServer struct {
	port string
}

func (l *lambdaProxyServer) run() {
	chain := gost.NewProxyChain()
	if err := chain.AddProxyNodeString(); err != nil {
		log.Fatal(err)
	}
	chain.Init()

	serverNode, err := gost.ParseProxyNode(l.port)
	if err != nil {
		log.Fatal(err)
	}

	certFile := gost.DefaultCertFile
	keyFile := gost.DefaultKeyFile
	cert, err := gost.LoadCertificate(certFile, keyFile)
	if err != nil {
		glog.Fatal(err)
	}

	server := gost.NewProxyServer(serverNode, chain, &tls.Config{Certificates: []tls.Certificate{cert}})
	log.Fatal(server.Serve())
}

func startLambdaProxyServer() *lambdaProxyServer {
	port := ":8080"
	server := &lambdaProxyServer{
		port: port,
	}
	go server.run()
	return server
}
