package awslambdaproxy

import (
	"time"
	"os"
	"log"
	"runtime"
)

func ServerInit(proxyPort string, tunnelPort string, regions []string, lambdaExecutionFrequency time.Duration) {
	lambdaExecutionTimeout := int64(lambdaExecutionFrequency.Seconds()) + int64(10)

	log.Println("Setting up Lambda infrastructure")
	err := setupLambdaInfrastructure(regions, lambdaExecutionTimeout)
	if err != nil {
		log.Println("Failed to setup Lambda infrastructure", err.Error())
		os.Exit(1)
	}

	log.Println("Starting TunnelConnectionManager")
	tunnelConnectionManager, err := newTunnelConnectionManager(tunnelPort, lambdaExecutionFrequency)
	if err != nil {
		log.Println("Failed to setup TunnelConnectionManager", err.Error())
		os.Exit(1)
	}

	log.Println("Starting LambdaExecutionManager")
	lambdaExecutionManager, err := newLambdaExecutionManager(tunnelPort, regions, lambdaExecutionFrequency)
	if err != nil {
		log.Println("Failed to setup LambdaExecutionManager", err.Error())
		os.Exit(1)
	}

	go func(){
		for {
			<-tunnelConnectionManager.emergencyTunnel
			log.Println("EMERGENCY TUNNEL STARTED")
			lambdaExecutionManager.executeFunction(0)
			time.Sleep(time.Second * 5)
		}
	}()

	tunnelConnectionManager.waitUntilReady()

	log.Println("Starting UserConnectionManager")
	userConnectionManager, err := newUserConnectionManager(proxyPort)
	if err != nil {
		log.Println("Failed to setup UserConnectionManager", err.Error())
		os.Exit(1)
	}

	log.Println("Starting DataCopyManager")
	newDataCopyManager(userConnectionManager, tunnelConnectionManager)

	runtime.Goexit()
}