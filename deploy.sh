#!/bin/bash

cd terraform

terraform init
terraform plan -out terraform.plan
terraform apply terraform.plan
IP=$(terraform output lambdaproxybox-ec2-ip)
EC2_USER=$(terraform output ec2_user)
ssh -o "StrictHostKeyChecking=no" -L 8080:localhost:8080 $EC2_USER@$IP ./awslambdaproxy run

cd -
