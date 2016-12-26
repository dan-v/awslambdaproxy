package awslambdaproxy

import (
	"time"
	"os"
	"log"
)

func ServerInit(proxyPort string, tunnelPort string, regions []string, lambdaExecutionFrequency time.Duration) {
	lambdaExecutionTimeout := int64(lambdaExecutionFrequency.Seconds()) + int64(5)

	log.Println("Setting up Lambda infrastructure")
	err := setupLambdaInfrastructure(regions, lambdaExecutionTimeout)
	if err != nil {
		log.Println("Failed to setup Lambda infrastructure", err.Error())
		os.Exit(1)
	}

	log.Println("Starting TunnelConnectionManager")
	tunnelConnectionManager, err := newTunnelConnectionManager(tunnelPort)
	if err != nil {
		log.Println("Failed to setup TunnelConnectionManager", err.Error())
		os.Exit(1)
	}
	go tunnelConnectionManager.run()

	log.Println("Starting LambdaExecutionManager")
	lambdaExecutionManager, err := newLambdaExecutionManager(tunnelPort, regions, lambdaExecutionFrequency)
	if err != nil {
		log.Println("Failed to setup LambdaExecutionManager", err.Error())
		os.Exit(1)
	}
	go lambdaExecutionManager.run()

	tunnelConnectionManager.waitUntilReady()

	log.Println("Starting UserConnectionManager")
	userConnectionManager, err := newUserConnectionManager(proxyPort)
	if err != nil {
		log.Println("Failed to setup UserConnectionManager", err.Error())
		os.Exit(1)
	}
	go userConnectionManager.run()

	log.Println("Starting DataCopyManager")
	dataCopyManager := newDataCopyManager(userConnectionManager, tunnelConnectionManager)
	dataCopyManager.run()
}