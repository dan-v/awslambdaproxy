package awslambdaproxy

import (
	"github.com/ginuerzh/gost"
	"log"
)

const (
	forwardProxy = "socks5://localhost:8082"
)

// LocalProxy is proxy listener and where to forward
type LocalProxy struct {
	listeners    []string
	forwardProxy string
}

func (l *LocalProxy) run() {
	chain := gost.NewProxyChain()
	if err := chain.AddProxyNodeString(l.forwardProxy); err != nil {
		log.Fatal(err)
	}
	chain.Init()

	for _, ns := range l.listeners {
		serverNode, err := gost.ParseProxyNode(ns)
		if err != nil {
			log.Fatal(err)
		}

		go func(node gost.ProxyNode) {
			server := gost.NewProxyServer(node, chain)
			log.Fatal(server.Serve())
		}(serverNode)
	}
}

// NewLocalProxy starts a local proxy that will forward to proxy running in Lambda
func NewLocalProxy(listeners []string) (*LocalProxy, error) {
	l := &LocalProxy{
		listeners:    listeners,
		forwardProxy: forwardProxy,
	}
	go l.run()
	return l, nil
}
