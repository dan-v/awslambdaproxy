import logging
from subprocess import Popen, PIPE

logger = logging.getLogger()
logger.setLevel(logging.INFO)

# Handler that will be called by Lambda
def handler(event, context):
    logger.info("Event: {}".format(event))
    logger.info("Context: {}".format(context))
    address = event['ConnectBackAddress']
    ssh_port = event['SSHPort']
    ssh_key = event['SSHKey']
    ssh_user = event['SSHUser']
    proxy_username = event['ProxyUsername']
    proxy_password = event['ProxyPassword']

    key_filename = "/tmp/privatekey"
    with open(key_filename, 'w') as key_file:
        key_file.write(ssh_key)

    command = "./awslambdaproxy-lambda -address {} -ssh-port {} -ssh-private-key {} -ssh-user {}  -proxy-username {} -proxy-password {}".format(address, ssh_port, key_filename, ssh_user, proxy_username, proxy_password)
    logger.info("Running: {}".format(command))
    try:
        proc = Popen(command, shell=True, stdout=PIPE, stderr=PIPE)
        out, err = proc.communicate()
        print out, err, proc.returncode
    except Exception as e:
        logger.error("Error: {}".format(e))
