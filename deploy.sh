#!/bin/bash

cd terraform
terraform init
terraform plan -out terraform.plan
terraform apply terraform.plan
IP=$(terraform output lambdaproxybox-ec2-ip)
EC2_USER=$(terraform output ec2_user)
cat <<EOF


###################################### To start proxy: ########################################
      ssh -o "StrictHostKeyChecking=no" $EC2_USER@$IP /home/$EC2_USER/awslambdaproxy run &
###############################################################################################
EOF
cd -
