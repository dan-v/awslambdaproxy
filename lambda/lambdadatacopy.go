package main

import (
	"os"
	"log"
	"net"
	"io"
	"sync"
	"time"
)

type LambdaDataCopyManager struct {
	lambdaTunnelConnection *LambdaTunnelConnection
	lambdaProxyServer *LambdaProxyServer
}

func (l *LambdaDataCopyManager) run() {
	for {
		proxySocketConn, proxySocketErr := net.Dial("tcp", l.lambdaProxyServer.port)
		if proxySocketErr != nil {
			log.Println("Failed to open connection to proxy", proxySocketErr)
			time.Sleep(time.Second)
			continue
		}
		log.Println("Started connection to proxy on port " + l.lambdaProxyServer.port)

		tunnelStream, tunnelErr := l.lambdaTunnelConnection.sess.Accept()
		if tunnelErr != nil {
			log.Println("Failed to start stream inside session", tunnelErr)
			os.Exit(1)
		}
		log.Println("Started stream inside session")

		go bidirectionalCopy(tunnelStream, proxySocketConn)
	}
}

func newLambdaDataCopyManager(p *LambdaProxyServer, t *LambdaTunnelConnection) *LambdaDataCopyManager {
	return &LambdaDataCopyManager{
		lambdaTunnelConnection: t,
		lambdaProxyServer: p,
	}
}

func bidirectionalCopy(src io.ReadWriteCloser, dst io.ReadWriteCloser) {
	defer dst.Close()
	defer src.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		io.Copy(dst, src)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		io.Copy(src, dst)
		wg.Done()
	}()
	wg.Wait()
}