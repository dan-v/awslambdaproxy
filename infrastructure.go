package awslambdaproxy

// majority of this is borrowed from https://github.com/goadapp/goad/blob/master/infrastructure/infrastructure.go

import (
	"time"
	"log"

	"github.com/pkg/errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/lambda"
)

const (
	lambdaFunctionName = "awslambdaproxy"
	lambdaFunctionHandler = "main.handler"
	lambdaFunctionRuntime = "python2.7"
	lambdaFunctionIamRole = "awslambdaproxy-role"
	lambdaFunctionIamRolePolicyName = "awslambdaproxy-role-policy"
	lambdaFunctionZipLocation = "data/lambda.zip"
)

type LambdaInfrastructure struct {
	config           *aws.Config
	regions          []string
	lambdaTimeout    int64
	lambdaMemorySize int64
}

func SetupLambdaInfrastructure() (error) {
	svc := iam.New(session.New(), &aws.Config{})

	_, err := svc.GetRole(&iam.GetRoleInput{
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
					PolicyDocument: aws.String(`
					{
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

func (infra *LambdaInfrastructure) setup() error {
	svc := iam.New(session.New(), infra.config)
	resp, err := svc.GetRole(&iam.GetRoleInput{
		RoleName: aws.String(lambdaFunctionIamRole),
	})
	if err != nil {
		return errors.Wrap(err, "Could not find IAM role " + lambdaFunctionIamRole + ". Probably need to run setup.")
	}
	roleArn := *resp.Role.Arn
	zip, err := Asset(lambdaFunctionZipLocation)
	if err != nil {
		return errors.Wrap(err, "Could not read ZIP file: " + lambdaFunctionZipLocation)
	}
	for _, region := range infra.regions {
		log.Println("Setting up Lambda function in region: " + region)
		err = infra.createOrUpdateLambdaFunction(region, roleArn, zip)
		if err != nil {
			return errors.Wrap(err, "Could not create Lambda function in region " + region)
		}
	}
	return nil
}

func setupLambdaInfrastructure(regions []string, memorySize int64, timeout int64) (error) {
	infra := LambdaInfrastructure{
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

func (infra *LambdaInfrastructure) createOrUpdateLambdaFunction(region, roleArn string, payload []byte) error {
	config := infra.config.WithRegion(region)
	svc := lambda.New(session.New(), config)

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

func (infra *LambdaInfrastructure) deleteLambdaFunction(svc *lambda.Lambda) error {
	_, err := svc.DeleteFunction(&lambda.DeleteFunctionInput{
		FunctionName: aws.String(lambdaFunctionName),
	})
	if err != nil {
		return err
	}
	return nil
}

func (infra *LambdaInfrastructure) createLambdaFunction(svc *lambda.Lambda, roleArn string, payload []byte) error {
	function, err := svc.CreateFunction(&lambda.CreateFunctionInput{
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
	return createLambdaAlias(svc, function.Version)
}

func (infra *LambdaInfrastructure) updateLambdaFunction(svc *lambda.Lambda, roleArn string, payload []byte) error {
	function, err := svc.UpdateFunctionCode(&lambda.UpdateFunctionCodeInput{
		ZipFile:      payload,
		FunctionName: aws.String(lambdaFunctionName),
		Publish:      aws.Bool(true),
	})
	if err != nil {
		return err
	}
	return createLambdaAlias(svc, function.Version)
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

func createLambdaAlias(svc *lambda.Lambda, functionVersion *string) error {
	_, err := svc.CreateAlias(&lambda.CreateAliasInput{
		FunctionName:    aws.String(lambdaFunctionName),
		FunctionVersion: functionVersion,
		Name:            aws.String(LambdaVersion()),
	})
	return err
}

func (infra *LambdaInfrastructure) createIAMLambdaRole(roleName string) (arn string, err error) {
	svc := iam.New(session.New(), infra.config)

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
				if err := infra.createIAMLambdaRolePolicy(*res.Role.RoleName); err != nil {
					return "", err
				}
				return *res.Role.Arn, nil
			}
		}
		return "", err
	}

	return *resp.Role.Arn, nil
}

func (infra *LambdaInfrastructure) createIAMLambdaRolePolicy(roleName string) error {
	svc := iam.New(session.New(), infra.config)

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