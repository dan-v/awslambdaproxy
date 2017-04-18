package awslambdaproxy

import (
	"time"
	"log"
	"runtime"
)

func ServerInit(sshUser string, sshPort string, regions []string, memory int64, frequency time.Duration,
		listeners []string, timeout int64) {
	publicIp, err := getPublicIp()
	if err != nil {
		log.Fatal("Error getting public IP address", err.Error())
	}

	log.Println("Setting up Lambda infrastructure")
	err = setupLambdaInfrastructure(regions, memory, timeout)
	if err != nil {
		log.Fatal("Failed to setup Lambda infrastructure", err.Error())
	}

	log.Println("Starting SSHManager")
	privateKey, err := NewSSHManager()
	if err != nil {
		log.Fatal("Failed to setup SSHManager", err.Error())
	}

	localProxy, err := NewLocalProxy(listeners)
	if err != nil {
		log.Fatal("Failed to setup LocalProxy", err.Error())
	}

	log.Println("Starting ConnectionManager")
	tunnelConnectionManager, err := newTunnelConnectionManager(frequency, localProxy)
	if err != nil {
		log.Fatal("Failed to setup ConnectionManager", err.Error())
	}

	log.Println("Starting LambdaExecutionManager")
	_, err = newLambdaExecutionManager(publicIp, regions, frequency,
		sshUser, sshPort, privateKey, tunnelConnectionManager.tunnelRedeployNeeded)
	if err != nil {
		log.Fatal("Failed to setup LambdaExecutionManager", err.Error())
	}

	log.Println("#######################################")
	log.Println("Proxy IP: ", publicIp)
	log.Println("Listeners: ", listeners)
	log.Println("#######################################")

	runtime.Goexit()
}