package main

import (
	"github.com/dan-v/awslambdaproxy"
	"flag"
	"os"
	"strings"
	"fmt"
	"time"
)

const (
	maxFrequency = "255"
)

func main() {
	regionsPtr := flag.String("regions", "us-west-2", "Regions to run proxy (e.g. us-west-2) (can be comma separated list)")
	frequencyPtr := flag.Int( "frequency", 180, "Frequency in seconds to execute Lambda function. If multiple regions are specified, this will cause traffic to rotate round robin at the interval specified here")
	proxyPortPtr := flag.String("proxy-port", "8080", "Port to listen for proxy connections")
	tunnelPortPtr := flag.String("tunnel-port", "8081", "Port to listen for reverse connection from Lambda")
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

	// handle frequency
	if *frequencyPtr > 255 {
		fmt.Println("Maximum freqency is " + maxFrequency + " seconds")
		os.Exit(1)
	}
	frequencySeconds := time.Second * time.Duration(*frequencyPtr)

	// handle aws env variables
	access := os.Getenv("AWS_ACCESS_KEY_ID")
	if access == "" {
		fmt.Println("Must specify environment variable AWS_ACCESS_KEY_ID")
		os.Exit(1)
	}
	secret := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if secret == "" {
		fmt.Println("Must specify environment variable AWS_SECRET_ACCESS_KEY")
		os.Exit(1)
	}

	regions := strings.Split(*regionsPtr, ",")
	awslambdaproxy.ServerInit(*proxyPortPtr, *tunnelPortPtr, regions, frequencySeconds)
}

