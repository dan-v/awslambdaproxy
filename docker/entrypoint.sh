#!/bin/bash

if [ "$1" == "setup" ]; then
  # ask for credentials to setup as this should be a different key with elevated permissions
  read -p 'Enter AWS_ACCESS_KEY_ID: ' AWS_ACCESS_KEY_ID
  read -sp 'Enter AWS_SECRET_ACCESS_KEY: ' AWS_SECRET_ACCESS_KEY
  export AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID
  export AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY
  /app/awslambdaproxy setup
  exit 0
fi

# if docker secret has been provided for AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY use it
if [[ -f /run/secrets/AWS_ACCESS_KEY_ID && -f /run/secrets/AWS_SECRET_ACCESS_KEY ]];
then
  export AWS_ACCESS_KEY_ID=$(cat /run/secrets/AWS_ACCESS_KEY_ID)
  export AWS_SECRET_ACCESS_KEY=$(cat /run/secrets/AWS_SECRET_ACCESS_KEY)
fi

# if still don't have keys, exit with error
if [ -z "${AWS_ACCESS_KEY_ID}" ]; then
  echo "Need to provide AWS_ACCESS_KEY_ID as secret or environment variable"
  exit 1
fi
if [ -z "${AWS_SECRET_ACCESS_KEY}" ]; then
  echo "Need to provide AWS_SECRET_ACCESS_KEY as secret or environment variable"
  exit 1
fi

# setup ssh
mkdir -p /tmp/etc/ssh
ssh-keygen -A -f /tmp
/usr/sbin/sshd

# run by default and pass any supplied arguments
/app/awslambdaproxy run $@