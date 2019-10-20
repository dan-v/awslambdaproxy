<b>awslambdaproxy</b> is an [AWS Lambda](https://aws.amazon.com/lambda/) powered HTTP/SOCKS web proxy. It provides a constantly rotating IP address for your network traffic from all regions where AWS Lambda is available. The goal is to obfuscate your traffic and make it harder to track you as a user.

![](/images/overview.gif?raw=true)

## Features
* HTTP/HTTPS/SOCKS5 proxy protocols support (including authentication).
* No special software required. Just configure your system to use a proxy.
* Each AWS Lambda region provides 1 outgoing IP address that gets rotated roughly every 4 hours. That means if you use 10 AWS regions, you'll get 60 unique IPs per day.
* Configurable IP rotation frequency between multiple regions.
* Personal proxy server not shared with anyone else.
* Mostly [AWS free tier](https://aws.amazon.com/free/) compatible (see FAQ below).

## Project status
Current code status: <b>proof of concept</b>. This is the first Go application that I've ever written. It has no tests. It may not work. It may blow up. Use at your own risk.

## How it works
At a high level, awslambdaproxy proxies TCP/UDP traffic through AWS Lambda regional endpoints. To do this, awslambdaproxy is setup on a publicly accessible host (e.g. EC2 instance) and it handles creating Lambda resources that run a proxy server ([ginuerzh/gost](https://github.com/ginuerzh/gost)). Since Lambda does not allow you to connect to bound ports in executing functions, a reverse SSH tunnel is established from the Lambda function to the host running awslambdaproxy. Once a tunnel connection is established, all user traffic is forwarded through this reverse tunnel to the proxy server. Lambda functions have a max execution time of 15 minutes, so there is a goroutine that continuously executes Lambda functions to ensure there is always a live tunnel in place. If multiple regions are specified, user traffic will be routed in a round robin fashion across these regions.

![](/images/how-it-works.png?raw=true)

## Installation

The easiest way is to run the `deploy.sh` script which uses Terraform to provision the infrastructure. You can download Terraform [here](https://www.terraform.io/downloads.html).

Once the deployment script has finished you'll be shown a command to run that will remotely start the proxy:

```bash
########################## To start proxy: ##########################
      ssh ubuntu@54.133.22.0 /home/ubuntu/awslambdaproxy run &
#####################################################################
```

Finally, you need to configure your browser with the IP and credentials that have just been displayed:
```bash
2019/10/20 17:16:32 #######################################
2019/10/20 17:16:32 Proxy IP:  54.133.22.0
2019/10/20 17:16:32 Listeners:  [admin:awslambdaproxy@:8080]
2019/10/20 17:16:32 #######################################
```

Now you're all set. Visit http://ifconfig.co/ to confirm that the proxy is working.

## Usage

1. Copy `awslambdaproxy` binary to a publicly accessible linux host (e.g. EC2 instance, VPS instance, etc). You will need to open the following ports on this host:
    * Port 22 - functions executing in AWS Lambda will open SSH connections back to the host running `awslambdaproxy`, so this port needs to be open to the world. The SSH key used here is dynamically generated at startup and added to the running users authorized_keys file.
    * Port 8080 - the default configuration will start a HTTP/SOCKS proxy listener on this port with default user/password authentication. If you don't want to publicly expose the proxy server, one option is to setup your own VPN server (e.g. [dosxvpn](https://github.com/dan-v/dosxvpn) or [algo](https://github.com/trailofbits/algo)), connect to it, and just run awslambdaproxy with the proxy listener only on localhost (-l localhost:8080).

2. Optional, but I'd highly recommend taking a look at the Minimal IAM Policies section below. This will allow you to setup minimal permissions required to setup and run the project. Otherwise, if you don't care about security you can always use an access key with full administrator privileges. 

3. `awslambdaproxy` will need access to credentials for AWS in some form. This can be either through exporting environment variables (as shown below), shared crendential file, or an IAM role if assigned to the instance you are running it on. See [this](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials) for more details.

    ```sh
    export AWS_ACCESS_KEY_ID=XXXXXXXXXX
    export AWS_SECRET_ACCESS_KEY=YYYYYYYYYYYYYYYYYYYYYY
    ```
4. Run `awslambdaproxy setup`. 

    ```sh
    ./awslambdaproxy setup
    ```

5. Run `awslambdaproxy run`. 

    ```sh
    ./awslambdaproxy run -r us-west-2,us-west-1,us-east-1,us-east-2
    ```
    
6. Configure your web browser (or OS) to use the SOCKS5 proxy on the publicly accessible host running `awslambdaproxy` on port 8080.

## Minimal IAM Policies
* This assumes you have the AWS CLI setup with an admin user
* Create a user with proper permissions needed to run the setup command. This user can be removed after running the setup command.
```
aws iam create-user --user-name awslambdaproxy-setup
aws iam put-user-policy --user-name awslambdaproxy-setup --policy-name awslambdaproxy-setup --policy-document file://iam/setup.json
aws iam create-access-key --user-name awslambdaproxy-setup
{
    "AccessKey": {
        "UserName": "awslambdaproxy-setup",
        "Status": "Active",
        "CreateDate": "2017-04-17T06:15:18.858Z",
        "SecretAccessKey": "xxxxxxxxxxxxxxxxxxxxxxxxxxxx",
        "AccessKeyId": "xxxxxxxxxxxxxx"
    }
}
```
* Create a user with proper permission needed to run the proxy.
```
aws iam create-user --user-name awslambdaproxy-run
aws iam put-user-policy --user-name awslambdaproxy-run --policy-name awslambdaproxy-run --policy-document file://iam/run.json
aws iam create-access-key --user-name awslambdaproxy-run
{
    "AccessKey": {
        "UserName": "awslambdaproxy-run",
        "Status": "Active",
        "CreateDate": "2017-04-17T06:18:27.531Z",
        "SecretAccessKey": "xxxxxxxxxxxxxxxxxxxxxxxxxxxx",
        "AccessKeyId": "xxxxxxxxxxxxxx"
    }
}
```

## FAQ
1. <b>Should I use awslambdaproxy?</b> That's up to you. Use at your own risk.
2. <b>Why did you use AWS Lambda for this?</b> The primary reason for using AWS Lambda in this project is the vast pool of IP addresses available that automatically rotate.
3. <b>How big is the pool of available IP addresses?</b> This I don't know, but I do know I did not have a duplicate IP while running the proxy for a week.
4. <b>Will this make me completely anonymous?</b> No, absolutely not. The goal of this project is just to obfuscate your web traffic by rotating your IP address. All of your traffic is going through AWS which could be traced back to your account. You can also be tracked still with [browser fingerprinting](https://panopticlick.eff.org/), etc. Your [IP address may still leak](https://ipleak.net/) due to WebRTC, Flash, etc.
5. <b>How often will my external IP address change?</b> For each region specified, the IP address will change roughly every 4 hours. This of course is subject to change at any moment as this is not something that is documented by AWS Lambda.
6. <b>How much does this cost?</b> awslambdaproxy should be able to run mostly on the [AWS free tier](https://aws.amazon.com/free/) minus bandwidth costs. It can run on a t2.micro instance and the default 128MB Lambda function that is constantly running should also fall in the free tier usage. The bandwidth is what will cost you money; you will pay for bandwidth usage for both EC2 and Lambda.
7. <b>Why does my connection drop periodically?</b> AWS Lambda functions can currently only execute for a maximum of 15 minutes. In order to maintain an ongoing proxy a new function is executed and all new traffic is cut over to it. Any ongoing connections to the previous Lambda function will hard stop after a timeout period. You generally won't see any issues for normal web browsing as connections are very short lived, but for any long lived connections you may see issues.

j Powered by
* [gost](https://github.com/ginuerzh/gost) - A simple security tunnel written in Golang.
* [yamux](https://github.com/hashicorp/yamux) - Golang connection multiplexing library.
* [goad](https://github.com/goadapp/goad) - Code was borrowed from this project to handle AWS Lambda zip creation and function upload.

## Build From Source
1. Install [Go](https://golang.org/dl/) and [go-bindata](https://github.com/kevinburke/go-bindata)

2. Fetch the project with `git clone`:

  ```sh
  git clone git@github.com:dan-v/awslambdaproxy.git && cd awslambdaproxy
  ```

3. Run make to build awslambdaproxy. You'll find your `awslambdaproxy` binary in the `build` folder.

  ```sh
  make
  ```
