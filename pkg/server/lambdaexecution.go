package server

import (
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/pkg/errors"
)

type lambdaExecutionManager struct {
	regions   []string
	frequency time.Duration
	publicIP  string
	sshPort   string
	sshKey    string
	sshUser   string
}

type lambdaPayload struct {
	UUID               string
	ConnectBackAddress string
	SSHPort            string
	SSHKey             string
	SSHUser            string
}

func (l *lambdaExecutionManager) run() {
	log.Println("Using public IP", l.publicIP)
	log.Println("Lambda execution frequency", l.frequency)
	count := 0
	setInvokeConfig := true
	for {
		if count > 0 {
			setInvokeConfig = false
		}
		for region := range l.regions {
			err := l.executeFunction(region, setInvokeConfig)
			if err != nil {
				log.Println(err)
			}
			time.Sleep(l.frequency)
		}
		count++
	}
}

func (l *lambdaExecutionManager) executeFunction(region int, setInvokeConfig bool) error {
	log.Println("Executing Lambda function in region", l.regions[region])
	sess, err := GetSessionAWS()
	if err != nil {
		return err
	}

	svc := lambda.New(sess, &aws.Config{Region: aws.String(l.regions[region])})

	if setInvokeConfig {
		maximumEventAgeInSeconds := int64(1800)
		maximumRetryAttempts := int64(0)
		log.Printf("Setting invoke configuration maximumRetryAttempts=%v maximumEventAgeInSeconds=%v\n",
			maximumRetryAttempts, maximumEventAgeInSeconds)
		_, err = svc.PutFunctionEventInvokeConfig(&lambda.PutFunctionEventInvokeConfigInput{
			FunctionName:             aws.String(lambdaFunctionName),
			MaximumEventAgeInSeconds: aws.Int64(maximumEventAgeInSeconds),
			MaximumRetryAttempts:     aws.Int64(maximumRetryAttempts),
		})
		if err != nil {
			return err
		}
	}

	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	lambdaPayload := lambdaPayload{
		UUID:               id.String(),
		ConnectBackAddress: l.publicIP,
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

	log.Printf("Invoking Lambda function with UUID=%v\n", lambdaPayload.UUID)
	_, err = svc.Invoke(params)
	if err != nil {
		return errors.Wrap(err, "Failed to execute Lambda function")
	}
	return nil
}

func newLambdaExecutionManager(publicIP string, regions []string, frequency time.Duration, sshUser string, sshPort string,
	privateKey []byte, onDemandExecution chan bool) (*lambdaExecutionManager, error) {
	executionManager := &lambdaExecutionManager{
		regions:   regions,
		frequency: frequency,
		publicIP:  publicIP,
		sshPort:   sshPort,
		sshKey:    string(privateKey[:]),
		sshUser:   sshUser,
	}
	go executionManager.run()

	go func() {
		for {
			<-onDemandExecution
			log.Println("Starting new tunnel as existing tunnel failed")
			err := executionManager.executeFunction(0, false)
			if err != nil {
				log.Println(err)
			}
			time.Sleep(time.Second * 5)
		}
	}()

	return executionManager, nil
}
