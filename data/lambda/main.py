import logging
from subprocess import Popen, PIPE

logger = logging.getLogger()
logger.setLevel(logging.INFO)

# Handler that will be called by Lambda
def handler(event, context):
    logger.info("Event: {}".format(event))
    logger.info("Context: {}".format(context))
    address = event['ConnectBackAddress']
    proxy_type = event['ProxyType']

    command = "./awslambdaproxy-lambda -address {} -proxy-type {}".format(address, proxy_type)
    logger.info("Running: {}".format(command))
    try:
        proc = Popen(command, shell=True, stdout=PIPE, stderr=PIPE)
        out, err = proc.communicate()
        print out, err, proc.returncode
    except Exception as e:
        logger.error("Error: {}".format(e))
