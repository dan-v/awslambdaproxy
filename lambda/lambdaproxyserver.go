package main

import (
	"os"
	"net"
	"net/http"
	"log"
	"time"

	"github.com/elazarl/goproxy"
)

type LambdaProxyServer struct {
	unixSocket string
	listener net.Listener
}

func (l *LambdaProxyServer) run() {
	log.Println("Starting proxy server on socket", l.unixSocket)
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true
	for {
		os.Remove(l.unixSocket)
		socketAddress, err := net.ResolveUnixAddr("unix", l.unixSocket)
		if err != nil {
			log.Println("Failed to resolve unix socket " + l.unixSocket)
			os.Exit(1)
		}
		unixListener, err := net.ListenUnix("unix", socketAddress)
		if err != nil {
			log.Println("Created listener on unix socket " + l.unixSocket)
			os.Exit(1)
		}
		l.listener = unixListener
		os.Chmod(l.unixSocket, 0777)
		log.Fatal(http.Serve(unixListener, proxy))
	}
}

func (l *LambdaProxyServer) isReady() bool {
	if l.listener != nil {
		return true
	} else {
		return false
	}
}

func startLambdaProxyServer() *LambdaProxyServer {
	ret := &LambdaProxyServer{
		unixSocket: proxyUnixSocket,
	}
	go ret.run()
	for {
		if ret.isReady() == true {
			break
		} else {
			log.Println("Proxy server not ready yet..")
			time.Sleep(time.Second * 1)
		}
	}
	return ret
}