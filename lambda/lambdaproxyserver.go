package main

import (
	"fmt"
	"log"
	"net"

	"github.com/ginuerzh/gost"
)

type lambdaProxyServer struct {
	port string
	ln gost.Listener
	server *gost.Server
}

func (l *lambdaProxyServer) run() {
	ln, err := gost.TCPListener(":0")
	if err != nil {
		log.Fatal(err)
	}
	l.ln = ln
	l.port = fmt.Sprintf(":%v", ln.Addr().(*net.TCPAddr).Port)

	h := gost.AutoHandler()
	s := &gost.Server{Listener: ln}
	l.server = s

	err = s.Serve(h)
	if err != nil {
		log.Printf("Server is now exiting: %v\n", err)
	}
}

func (l *lambdaProxyServer) close() {
	log.Println("Closing down server")
	err := l.server.Close()
	if err != nil {
		log.Printf("closing server error: %v\n", err)
	}
	log.Println("Closing down listener")
	err = l.ln.Close()
	if err != nil {
		log.Printf("closing listener error: %v\n", err)
	}
}

func startLambdaProxyServer() *lambdaProxyServer {
	server := &lambdaProxyServer{}
	go server.run()
	return server
}
