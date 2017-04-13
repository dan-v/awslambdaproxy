package main

import (
	"log"

	"github.com/ginuerzh/gost"
)

type LambdaProxyServer struct {
	port string
}

func (l *LambdaProxyServer) run() {
	chain := gost.NewProxyChain()
	if err := chain.AddProxyNodeString(); err != nil {
		log.Fatal(err)
	}
	chain.Init()

	serverNode, err := gost.ParseProxyNode(l.port)
	if err != nil {
		log.Fatal(err)
	}

	server := gost.NewProxyServer(serverNode, chain)
	log.Fatal(server.Serve())
}

func startLambdaProxyServer() *LambdaProxyServer {
	port := ":8080"
	server := &LambdaProxyServer{
		port: port,
	}
	go server.run()
	return server
}