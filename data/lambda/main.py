import logging
from subprocess import Popen, PIPE

logger = logging.getLogger()
logger.setLevel(logging.INFO)

# Handler that will be called by Lambda
def handler(event, context):
    logger.info("Event: {}".format(event))
    logger.info("Context: {}".format(context))

    command = "./awslambdaproxy-lambda -address {}".format(event)
    logger.info("Running: {}".format(command))
    try:
        proc = Popen(command, shell=True, stdout=PIPE, stderr=PIPE)
        out, err = proc.communicate()
        print out, err, proc.returncode
    except Exception as e:
        logger.error("Error: {}".format(e))
