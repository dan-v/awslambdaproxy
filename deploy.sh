#!/bin/bash

#terraform init
#terraform plan -out terraform.plan
#terraform apply terraform.plan
#IP=$(terraform output lambdaproxybox-ec2-ip)
IP=54.123.23.300
cat <<EOF


########################## To start proxy: ##########################
      ssh ubuntu@$IP /home/ubuntu/awslambdaproxy run &
#####################################################################
EOF
