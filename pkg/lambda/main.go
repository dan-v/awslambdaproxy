package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
)

const privateKeyFile = "/tmp/privatekey"

type Request struct {
	UUID    string `json:"UUID"`
	Address string `json:"ConnectBackAddress"`
	SSHPort string `json:"SSHPort"`
	SSHKey  string `json:"SSHKey"`
	SSHUser string `json:"SSHUser"`
}

type Response struct {
	Message string `json:"message"`
	Ok      bool   `json:"ok"`
}

func Handler(request Request) (Response, error) {
	log.Printf("Processing request UUID=%v\n", request.UUID)
	sshKeyData := []byte(request.SSHKey)
	err := ioutil.WriteFile(privateKeyFile, sshKeyData, 0600)
	if err != nil {
		log.Fatal("Failed to write SSH key to disk. ", err)
	}

	log.Println("Starting proxy server")
	lambdaProxyServer := startLambdaProxyServer()
	defer lambdaProxyServer.close()

	log.Printf("Establishing ssh tunnel connection to %v\n", request.Address)
	lambdaTunnelConnection, err := setupLambdaTunnelConnection(request.Address, request.SSHPort, request.SSHUser, privateKeyFile)
	if err != nil {
		log.Fatalf("Failed to establish connection to %v: %v\n", request.Address, err)
	}
	defer lambdaTunnelConnection.close()

	log.Println("Starting lambdaDataCopyManager")
	dataCopyManager := newLambdaDataCopyManager(lambdaProxyServer, lambdaTunnelConnection)
	dataCopyManager.run()
	return Response{
		Message: fmt.Sprintf("Finished processing request"),
		Ok:      true,
	}, nil
}

func main() {
	lambda.Start(Handler)
}
