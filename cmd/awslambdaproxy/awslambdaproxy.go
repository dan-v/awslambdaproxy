package main

import (
	"github.com/dan-v/awslambdaproxy"
	"flag"
	"os"
	"strings"
)

func main() {
	proxyPortPtr := flag.String("proxy-port", "8080", "Port to listen for proxy connections")
	tunnelPortPtr := flag.String("tunnel-port", "8081", "Port to listen for reverse connection from Lambda")
	regionsPtr := flag.String("regions", "us-west-2", "Regions to run proxy (e.g. us-west-2) (can be comma separated list)")
	flag.Parse()

	if *proxyPortPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *tunnelPortPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *regionsPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	regions := strings.Split(*regionsPtr, ",")
	awslambdaproxy.ServerInit(*proxyPortPtr, *tunnelPortPtr, regions)
}

