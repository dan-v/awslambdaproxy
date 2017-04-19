package awslambdaproxy

import (
	"log"
	"runtime"
	"time"
)

// ServerInit is the main entrypoint for the server portion of awslambdaproxy
func ServerInit(sshUser string, sshPort string, regions []string, memory int64, frequency time.Duration,
	listeners []string, timeout int64) {
	publicIP, err := getPublicIP()
	if err != nil {
		log.Fatal("Error getting public IP address", err.Error())
	}

	log.Println("Setting up Lambda infrastructure")
	err = setupLambdaInfrastructure(regions, memory, timeout)
	if err != nil {
		log.Fatal("Failed to setup Lambda infrastructure", err.Error())
	}

	log.Println("Starting sshManager")
	privateKey, err := NewSSHManager()
	if err != nil {
		log.Fatal("Failed to setup sshManager", err.Error())
	}

	localProxy, err := NewLocalProxy(listeners)
	if err != nil {
		log.Fatal("Failed to setup LocalProxy", err.Error())
	}

	log.Println("Starting connectionManager")
	tunnelConnectionManager, err := newTunnelConnectionManager(frequency, localProxy)
	if err != nil {
		log.Fatal("Failed to setup connectionManager", err.Error())
	}

	log.Println("Starting lambdaExecutionManager")
	_, err = newLambdaExecutionManager(publicIP, regions, frequency,
		sshUser, sshPort, privateKey, tunnelConnectionManager.tunnelRedeployNeeded)
	if err != nil {
		log.Fatal("Failed to setup lambdaExecutionManager", err.Error())
	}

	log.Println("#######################################")
	log.Println("Proxy IP: ", publicIP)
	log.Println("Listeners: ", listeners)
	log.Println("#######################################")

	runtime.Goexit()
}
