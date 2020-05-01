package main

import (
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type lambdaDataCopyManager struct {
	lambdaTunnelConnection *lambdaTunnelConnection
	lambdaProxyServer      *lambdaProxyServer
}

func (l *lambdaDataCopyManager) run() {
	for {
		proxySocketConn, proxySocketErr := net.Dial("tcp", l.lambdaProxyServer.port)
		if proxySocketErr != nil {
			log.Printf("Failed to open connection to proxy: %v\n", proxySocketErr)
			time.Sleep(time.Second)
			continue
		}
		log.Printf("Opened local connection to proxy on port %v\n", l.lambdaProxyServer.port)

		tunnelStream, tunnelErr := l.lambdaTunnelConnection.sess.Accept()
		if tunnelErr != nil {
			log.Printf("Failed to start new stream: %v. Exiting function.\n", tunnelErr)
			return
		}
		log.Println("Started new stream")

		go bidirectionalCopy(tunnelStream, proxySocketConn)
	}
}

func newLambdaDataCopyManager(p *lambdaProxyServer, t *lambdaTunnelConnection) *lambdaDataCopyManager {
	return &lambdaDataCopyManager{
		lambdaTunnelConnection: t,
		lambdaProxyServer:      p,
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
