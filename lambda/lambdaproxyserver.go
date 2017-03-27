package main

import (
	"os"
	"net"
	"log"
	"time"

	"github.com/armon/go-socks5"
)

type LambdaProxyServer struct {
	unixSocket string
	unixListener net.Listener
	username string
	password string
}

func (l *LambdaProxyServer) run() {
	log.Println("Starting Socks proxy server on socket", l.unixSocket)
	creds := socks5.StaticCredentials{
		l.username:l.password,
	}
	cator := socks5.UserPassAuthenticator{Credentials: creds}
	conf := &socks5.Config{
		AuthMethods: []socks5.Authenticator{cator},
		Logger:      log.New(os.Stdout, "", log.LstdFlags),
	}
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
		l.unixListener = unixListener
		os.Chmod(l.unixSocket, 0777)
		log.Fatal(server.Serve(unixListener))
	}
}

func (l *LambdaProxyServer) isReady() bool {
	if l.unixListener != nil {
		return true
	} else {
		return false
	}
}

func startLambdaProxyServer(username string, password string) *LambdaProxyServer {
	server := &LambdaProxyServer{
		unixSocket: proxyUnixSocket,
		username: username,
		password: password,
	}
	go server.run()
	for {
		if server.isReady() == true {
			break
		} else {
			log.Println("Proxy server not ready yet..")
			time.Sleep(time.Second * 1)
		}
	}
	return server
}