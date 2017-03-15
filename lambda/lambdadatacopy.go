package main

import (
	"os"
	"log"
	"net"
	"io"
	"sync"
)

type LambdaDataCopyManager struct {
	lambdaTunnelConnection *LambdaTunnelConnection
	lambdaProxyServer *LambdaProxyServer
}

func (l *LambdaDataCopyManager) run() {
	for {
		proxySocketConn, proxySocketErr := net.Dial("unix", l.lambdaProxyServer.unixSocket)
		if proxySocketErr != nil {
			log.Println("Failed to open connection to proxy", proxySocketErr)
			os.Exit(1)
		}
		log.Println("Started connection to proxy on socket " + l.lambdaProxyServer.unixSocket)

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

func bidirectionalCopy(dst io.ReadWriteCloser, src io.ReadWriteCloser) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		io.Copy(dst, src)
		dst.Close()
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		io.Copy(src, dst)
		src.Close()
		wg.Done()
	}()
	wg.Wait()
}