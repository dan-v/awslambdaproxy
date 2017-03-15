package main

import (
	"github.com/dan-v/awslambdaproxy"
	"flag"
	"os"
	"strings"
	"time"
	"os/user"
	"log"
	"strconv"
)

const (
	// Max execution time on lambda is 300 seconds currently
	lambdaMaxFrequencySeconds = 290
	lambdaMinMemorySize       = 128
	lambdaMaxMemorySize       = 1536
	lambdaDefaultMemorySize   = lambdaMinMemorySize
)

func main() {
	user, err := user.Current()
	if err != nil {
		log.Println("Failed to get current username")
		os.Exit(1)
	}

	lambdaRegionsPtr := flag.String("lambda-regions", "us-west-2", "Regions to run proxy " +
		"(e.g. us-west-2) (can be comma separated list)")
	lambdaFrequencyPtr := flag.Int("lambda-frequency", lambdaMaxFrequencySeconds, "Frequency in " +
		"seconds to execute Lambda function.  Max=" + strconv.Itoa(lambdaMaxFrequencySeconds) + ". If multiple " +
		"lambda-regions are specified, this will cause traffic to rotate round robin at the interval " +
		"specified here")
	lambdaMemorySizePtr := flag.Int("lambda-memory", lambdaDefaultMemorySize, "Memory size in MB "+
		"for Lambda function. Higher memory may allow for faster network throughput.")
	sshUserPtr := flag.String("ssh-user", user.Username, "SSH user for tunnel connections from Lambda")
	sshPortPtr := flag.String("ssh-port", "22", "SSH port for tunnel connections from Lambda")
	proxyPortPtr := flag.String("proxy-port", "8080", "Port to listen for client socks proxy "+
		"connections")
	proxyUsernamePtr := flag.String("proxy-username", "admin", "Username for proxy authentication")
	proxyPasswordPtr := flag.String("proxy-password", "admin", "Password for proxy authentication")
	flag.Parse()

	if *proxyPortPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *lambdaRegionsPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *sshUserPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *sshPortPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *proxyUsernamePtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *proxyPasswordPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// check memory
	if *lambdaMemorySizePtr > lambdaMaxMemorySize {
		log.Println("Maximum lambda memory size is " + strconv.Itoa(lambdaMaxMemorySize) + " MB")
		os.Exit(1)
	}
	if *lambdaMemorySizePtr < lambdaMinMemorySize {
		log.Println("Minimum lambda memory size is " + strconv.Itoa(lambdaMinMemorySize) + " MB")
		os.Exit(1)
	}
	lambdaMemorySize := int64(*lambdaMemorySizePtr)

	// check frequency
	if *lambdaFrequencyPtr > lambdaMaxFrequencySeconds {
		log.Println("Maximum lambda frequency is " + strconv.Itoa(lambdaMaxFrequencySeconds) + " seconds")
		os.Exit(1)
	}
	lambdaFrequencySeconds := time.Second * time.Duration(*lambdaFrequencyPtr)
	lambdaExecutionTimeout := int64(lambdaFrequencySeconds.Seconds()) + int64(10)

	// check for required aws keys
	access := os.Getenv("AWS_ACCESS_KEY_ID")
	if access == "" {
		log.Println("Must specify environment variable AWS_ACCESS_KEY_ID")
		os.Exit(1)
	}
	secret := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if secret == "" {
		log.Println("Must specify environment variable AWS_SECRET_ACCESS_KEY")
		os.Exit(1)
	}

	// handle regions
	lambdaRegions := strings.Split(*lambdaRegionsPtr, ",")

	awslambdaproxy.ServerInit(*proxyPortPtr, *sshUserPtr, *sshPortPtr, *proxyUsernamePtr, *proxyPasswordPtr,
		lambdaRegions, lambdaMemorySize, lambdaFrequencySeconds, lambdaExecutionTimeout)
}
