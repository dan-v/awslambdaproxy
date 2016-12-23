package awslambdaproxy

import (
	"time"
	"encoding/json"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/pkg/errors"
)

type LambdaExecutionManager struct {
	port string
	regions []string
	frequency time.Duration
	publicIp string
}

func (l *LambdaExecutionManager) run() {
	log.Println("Using public IP", l.publicIp)
	log.Println("Lambda execution frequency", l.frequency)
	for {
		sess := session.New(&aws.Config{})
		for region := range l.regions {
			svc := lambda.New(sess, &aws.Config{Region: aws.String(l.regions[region])})
			payload, _ := json.Marshal(l.publicIp + ":" + l.port)
			params := &lambda.InvokeInput{
				FunctionName:   aws.String(lambdaFunctionName),
				InvocationType: aws.String(lambda.InvocationTypeEvent),
				Payload:        payload,
			}
			log.Println("Executing Lambda function in region", l.regions[region])
			_, err := svc.Invoke(params)
			if err != nil {
				log.Println("Failed to execute Lambda function.", err.Error())
			}
			time.Sleep(l.frequency)
		}
	}
}

func newLambdaExecutionManager(port string, regions []string, frequency time.Duration) (*LambdaExecutionManager, error) {
	publicIp, err := getPublicIp()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting IP address")
	}
	return &LambdaExecutionManager{
		port: port,
		regions: regions,
		frequency: frequency,
		publicIp: publicIp,
	}, nil
}