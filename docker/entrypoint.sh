#!/bin/bash

if [ "$1" == "setup" ]; then
  read -p 'Enter AWS_ACCESS_KEY_ID: ' AWS_ACCESS_KEY_ID
  read -sp 'Enter AWS_SECRET_ACCESS_KEY: ' AWS_SECRET_ACCESS_KEY
  export AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID
  export AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY
  /app/awslambdaproxy setup
  exit 0
fi

mkdir /tmp/etc
mkdir /tmp/etc/ssh
ssh-keygen -A -f /tmp
/usr/sbin/sshd

if [[ "${DEBUG_PROXY}" == 'true' ]]; then
  DEBUG_PROXY="--debug-proxy"
fi

/app/awslambdaproxy run -r ${AWS_REGIONS} --ssh-port ${SSH_PORT} -l ${PROXY_LISTENERS} \
  -f ${PROXY_FREQUENCY_REFRESH} -m ${AWS_LAMBDA_MEMORY} ${DEBUG_PROXY}