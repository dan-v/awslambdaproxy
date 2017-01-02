package main

import (
	"os"
	"net"
	"log"
	"net/http"
	"time"

	"github.com/armon/go-socks5"
	"github.com/elazarl/goproxy"
)

type LambdaProxyServer struct {
	unixSocket string
	proxyType string
	listener net.Listener
}

func (l *LambdaProxyServer) runHttp() {
	log.Println("Starting HTTP proxy server on socket", l.unixSocket)
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

func (l *LambdaProxyServer) runSocks() {
	log.Println("Starting Socks proxy server on socket", l.unixSocket)
	conf := &socks5.Config{}
	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}
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
		log.Fatal(server.Serve(unixListener))
	}
}

func (l *LambdaProxyServer) isReady() bool {
	if l.listener != nil {
		return true
	} else {
		return false
	}
}

func startLambdaProxyServer(proxyType string) *LambdaProxyServer {
	ret := &LambdaProxyServer{
		unixSocket: proxyUnixSocket,
		proxyType: proxyType,
	}
	if ret.proxyType == "http" {
		go ret.runHttp()
	} else {
		go ret.runSocks()
	}
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