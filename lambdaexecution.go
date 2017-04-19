package awslambdaproxy

import (
	"encoding/json"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/pkg/errors"
)

type LambdaExecutionManager struct {
	regions   []string
	frequency time.Duration
	publicIp  string
	sshPort   string
	sshKey    string
	sshUser   string
}

type LambdaPayload struct {
	ConnectBackAddress string
	SSHPort            string
	SSHKey             string
	SSHUser            string
}

func (l *LambdaExecutionManager) run() {
	log.Println("Using public IP", l.publicIp)
	log.Println("Lambda execution frequency", l.frequency)
	for {
		for region := range l.regions {
			l.executeFunction(region)
			time.Sleep(l.frequency)
		}
	}
}

func (l *LambdaExecutionManager) executeFunction(region int) error {
	log.Println("Executing Lambda function in region", l.regions[region])
	sess := session.New(&aws.Config{})
	svc := lambda.New(sess, &aws.Config{Region: aws.String(l.regions[region])})
	lambdaPayload := LambdaPayload{
		ConnectBackAddress: l.publicIp,
		SSHPort:            l.sshPort,
		SSHKey:             l.sshKey,
		SSHUser:            l.sshUser,
	}
	payload, _ := json.Marshal(lambdaPayload)
	params := &lambda.InvokeInput{
		FunctionName:   aws.String(lambdaFunctionName),
		InvocationType: aws.String(lambda.InvocationTypeEvent),
		Payload:        payload,
	}
	_, err := svc.Invoke(params)
	if err != nil {
		return errors.Wrap(err, "Failed to execute Lambda function")
	}
	return nil
}

func newLambdaExecutionManager(publicIp string, regions []string, frequency time.Duration, sshUser string, sshPort string,
	privateKey []byte, onDemandExecution chan bool) (*LambdaExecutionManager, error) {
	executionManager := &LambdaExecutionManager{
		regions:   regions,
		frequency: frequency,
		publicIp:  publicIp,
		sshPort:   sshPort,
		sshKey:    string(privateKey[:]),
		sshUser:   sshUser,
	}
	go executionManager.run()

	go func() {
		for {
			<-onDemandExecution
			log.Println("Starting new tunnel as existing tunnel failed")
			executionManager.executeFunction(0)
			time.Sleep(time.Second * 5)
		}
	}()

	return executionManager, nil
}
