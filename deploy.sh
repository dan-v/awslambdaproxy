#!/bin/bash

terraform init
terraform plan -out terraform.plan
terraform apply terraform.plan
IP=$(terraform output lambdaproxybox-ec2-ip)
echo -e "\n\n"
echo "####################### To start proxy: ########################"
echo ""
echo "        ssh ubuntu@$IP /home/ubuntu/awslambdaproxy run &        " 
echo ""
echo "################################################################"
