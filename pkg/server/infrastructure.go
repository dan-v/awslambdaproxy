package server

// majority of this is borrowed from https://github.com/goadapp/goad/blob/master/infrastructure/infrastructure.go

import (
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/pkg/errors"
)

const (
	lambdaFunctionName              = "awslambdaproxy"
	lambdaFunctionHandler           = "main"
	lambdaFunctionRuntime           = "go1.x"
	lambdaFunctionIamRole           = "awslambdaproxy-role"
	lambdaFunctionIamRolePolicyName = "awslambdaproxy-role-policy"
	lambdaFunctionZipLocation       = "artifacts/lambda.zip"
)

type lambdaInfrastructure struct {
	config           *aws.Config
	regions          []string
	lambdaTimeout    int64
	lambdaMemorySize int64
}

// SetupLambdaInfrastructure sets up IAM role needed to run awslambdaproxy
func SetupLambdaInfrastructure() error {
	sess, err := GetSessionAWS()
	if err != nil {
		return err
	}

	svc := iam.New(sess, &aws.Config{})
	_, err = svc.GetRole(&iam.GetRoleInput{
		RoleName: aws.String(lambdaFunctionIamRole),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NoSuchEntity" {
				_, err := svc.CreateRole(&iam.CreateRoleInput{
					AssumeRolePolicyDocument: aws.String(`{
					  "Version": "2012-10-17",
					  "Statement": {
					    "Effect": "Allow",
					    "Principal": {"Service": "lambda.amazonaws.com"},
					    "Action": "sts:AssumeRole"
					  }
				    	}`),
					RoleName: aws.String(lambdaFunctionIamRole),
					Path:     aws.String("/"),
				})
				if err != nil {
					return err
				}
				_, err = svc.PutRolePolicy(&iam.PutRolePolicyInput{
					PolicyDocument: aws.String(`{
					  "Version": "2012-10-17",
					  "Statement": [
					    {
					      "Action": [
						"logs:CreateLogGroup",
						"logs:CreateLogStream",
						"logs:PutLogEvents"
					      ],
					      "Effect": "Allow",
					      "Resource": "arn:aws:logs:*:*:*"
					    }
					  ]
					}`),
					PolicyName: aws.String(lambdaFunctionIamRolePolicyName),
					RoleName:   aws.String(lambdaFunctionIamRole),
				})
				if err != nil {
					return err
				}

				return nil
			}
		} else {
			return err
		}
	} else {
		log.Println("Setup has already been run successfully")
		return nil
	}

	return nil
}

func (infra *lambdaInfrastructure) setup() error {
	sess, err := GetSessionAWS()
	if err != nil {
		return err
	}

	svc := iam.New(sess, infra.config)
	resp, err := svc.GetRole(&iam.GetRoleInput{
		RoleName: aws.String(lambdaFunctionIamRole),
	})
	if err != nil {
		return errors.Wrap(err, "Could not find IAM role "+lambdaFunctionIamRole+". Probably need to run setup.")
	}
	roleArn := *resp.Role.Arn
	zip, err := Asset(lambdaFunctionZipLocation)
	if err != nil {
		return errors.Wrap(err, "Could not read ZIP file: "+lambdaFunctionZipLocation)
	}
	for _, region := range infra.regions {
		log.Println("Setting up Lambda function in region: " + region)
		err = infra.createOrUpdateLambdaFunction(sess, region, roleArn, zip)
		if err != nil {
			return errors.Wrap(err, "Could not create Lambda function in region "+region)
		}
	}
	return nil
}

func setupLambdaInfrastructure(regions []string, memorySize int64, timeout int64) error {
	infra := lambdaInfrastructure{
		regions:          regions,
		config:           &aws.Config{},
		lambdaTimeout:    timeout,
		lambdaMemorySize: memorySize,
	}
	if err := infra.setup(); err != nil {
		return errors.Wrap(err, "Could not setup Lambda Infrastructure")
	}
	return nil
}

func (infra *lambdaInfrastructure) createOrUpdateLambdaFunction(sess *session.Session, region, roleArn string, payload []byte) error {
	config := infra.config.WithRegion(region)

	svc := lambda.New(sess, config)
	exists, err := lambdaExists(svc)
	if err != nil {
		return err
	}

	if exists {
		err := infra.deleteLambdaFunction(svc)
		if err != nil {
			return err
		}
	}

	return infra.createLambdaFunction(svc, roleArn, payload)
}

func (infra *lambdaInfrastructure) deleteLambdaFunction(svc *lambda.Lambda) error {
	_, err := svc.DeleteFunction(&lambda.DeleteFunctionInput{
		FunctionName: aws.String(lambdaFunctionName),
	})
	if err != nil {
		return err
	}
	return nil
}

func (infra *lambdaInfrastructure) createLambdaFunction(svc *lambda.Lambda, roleArn string, payload []byte) error {
	_, err := svc.CreateFunction(&lambda.CreateFunctionInput{
		Code: &lambda.FunctionCode{
			ZipFile: payload,
		},
		FunctionName: aws.String(lambdaFunctionName),
		Handler:      aws.String(lambdaFunctionHandler),
		Role:         aws.String(roleArn),
		Runtime:      aws.String(lambdaFunctionRuntime),
		MemorySize:   aws.Int64(infra.lambdaMemorySize),
		Publish:      aws.Bool(true),
		Timeout:      aws.Int64(infra.lambdaTimeout),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "InvalidParameterValueException" {
				time.Sleep(time.Second)
				return infra.createLambdaFunction(svc, roleArn, payload)
			}
		}
		return err
	}
	return nil
}

func lambdaExists(svc *lambda.Lambda) (bool, error) {
	_, err := svc.GetFunction(&lambda.GetFunctionInput{
		FunctionName: aws.String(lambdaFunctionName),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "ResourceNotFoundException" {
				return false, nil
			}
		}
		return false, err
	}

	return true, nil
}

func (infra *lambdaInfrastructure) createIAMLambdaRole(sess *session.Session, roleName string) (arn string, err error) {
	svc := iam.New(sess, infra.config)
	resp, err := svc.GetRole(&iam.GetRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NoSuchEntity" {
				res, err := svc.CreateRole(&iam.CreateRoleInput{
					AssumeRolePolicyDocument: aws.String(`{
					  "Version": "2012-10-17",
					  "Statement": {
					    "Effect": "Allow",
					    "Principal": {"Service": "lambda.amazonaws.com"},
					    "Action": "sts:AssumeRole"
					  }
				    	}`),
					RoleName: aws.String(roleName),
					Path:     aws.String("/"),
				})
				if err != nil {
					return "", err
				}
				if err := infra.createIAMLambdaRolePolicy(sess, *res.Role.RoleName); err != nil {
					return "", err
				}
				return *res.Role.Arn, nil
			}
		}
		return "", err
	}

	return *resp.Role.Arn, nil
}

func (infra *lambdaInfrastructure) createIAMLambdaRolePolicy(sess *session.Session, roleName string) error {
	svc := iam.New(sess, infra.config)
	_, err := svc.PutRolePolicy(&iam.PutRolePolicyInput{
		PolicyDocument: aws.String(`{
		  "Version": "2012-10-17",
		  "Statement": [
		    {
		      "Action": [
			"logs:CreateLogGroup",
			"logs:CreateLogStream",
			"logs:PutLogEvents"
		      ],
		      "Effect": "Allow",
		      "Resource": "arn:aws:logs:*:*:*"
		    }
		  ]
		}`),
		PolicyName: aws.String(lambdaFunctionIamRolePolicyName),
		RoleName:   aws.String(roleName),
	})
	return err
}
