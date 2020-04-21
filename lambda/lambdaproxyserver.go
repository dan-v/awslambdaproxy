package main

import (
	"log"

	"github.com/ginuerzh/gost"
)

type lambdaProxyServer struct {
	port string
}

func (l *lambdaProxyServer) run() {
	ln, err := gost.TCPListener(l.port)
	if err != nil {
		log.Fatal(err)
	}
	h := gost.AutoHandler()
	s := &gost.Server{Listener: ln}
	log.Fatal(s.Serve(h))
}

func startLambdaProxyServer() *lambdaProxyServer {
	port := ":8080"
	server := &lambdaProxyServer{
		port: port,
	}
	go server.run()
	return server
}
