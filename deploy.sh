#!/bin/bash

terraform init
terraform plan -out terraform.plan
terraform apply terraform.plan
IP=$(terraform output lambdaproxybox-ec2-ip)
EC2_USER=$(terraform output ec2_user)
cat <<EOF


########################## To start proxy: ##########################
      ssh ubuntu@$IP /home/$EC2_USER/awslambdaproxy run &
#####################################################################
EOF
