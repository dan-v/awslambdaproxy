package awslambdaproxy

import (
	"time"
	"os"
	"log"
	"runtime"
)

func ServerInit(proxyPort string, sshUser string, sshPort string, proxyUsername string, proxyPassword string,
			lambdaRegions []string, lambdaMemorySize int64, lambdaExecutionFrequency time.Duration,
			lambdaExecutionTimeout int64) {
	publicIp, err := getPublicIp()
	if err != nil {
		log.Println("Error getting public IP address", err.Error())
		os.Exit(1)
	}

	log.Println("Setting up Lambda infrastructure")
	err = setupLambdaInfrastructure(lambdaRegions, lambdaMemorySize, lambdaExecutionTimeout)
	if err != nil {
		log.Println("Failed to setup Lambda infrastructure", err.Error())
		os.Exit(1)
	}

	log.Println("Starting SSHManager")
	privateKey, err := NewSSHManager()
	if err != nil {
		log.Println("Failed to setup SSHManager", err.Error())
		os.Exit(1)
	}

	log.Println("Starting TunnelConnectionManager")
	tunnelConnectionManager, err := newTunnelConnectionManager(lambdaExecutionFrequency)
	if err != nil {
		log.Println("Failed to setup TunnelConnectionManager", err.Error())
		os.Exit(1)
	}

	log.Println("Starting LambdaExecutionManager")
	_, err = newLambdaExecutionManager(publicIp, lambdaRegions, lambdaExecutionFrequency,
		sshUser, sshPort, privateKey, proxyUsername, proxyPassword, tunnelConnectionManager.emergencyTunnel)
	if err != nil {
		log.Println("Failed to setup LambdaExecutionManager", err.Error())
		os.Exit(1)
	}

	log.Println("Starting UserConnectionManager")
	userConnectionManager, err := newUserConnectionManager(proxyPort)
	if err != nil {
		log.Println("Failed to setup UserConnectionManager", err.Error())
		os.Exit(1)
	}

	tunnelConnectionManager.waitUntilReady(lambdaExecutionFrequency)

	log.Println("#######################################")
	log.Println("SOCKS Proxy IP: ", publicIp)
	log.Println("SOCKS Username: ", proxyUsername)
	log.Println("SOCKS Password: ", proxyPassword)
	log.Println("#######################################")

	newDataCopyManager(userConnectionManager, tunnelConnectionManager)

	runtime.Goexit()
}