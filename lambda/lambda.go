package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
)

const privateKeyFile = "/tmp/privatekey"

type Request struct {
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
	log.Println("Starting lambdaProxyServer")
	lambdaProxyServer := startLambdaProxyServer()

	log.Println("Establishing tunnel connection to", request.Address)

	sshKeyData := []byte(request.SSHKey)
	err := ioutil.WriteFile(privateKeyFile, sshKeyData, 0600)
	if err != nil {
		log.Fatal("Failed to write SSH key to disk. ", err)
	}

	lambdaTunnelConnection, err := setupLambdaTunnelConnection(request.Address, request.SSHPort, request.SSHUser, privateKeyFile)
	if err != nil {
		log.Fatal("Failed to establish connection to "+request.Address, err)
	}

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
