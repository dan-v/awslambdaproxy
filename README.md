<b>awslambdaproxy</b> is an [AWS Lambda](https://aws.amazon.com/lambda/) powered HTTP(s)/SOCKS web proxy. It provides a constantly rotating IP address for your web traffic from all regions where AWS Lambda is available. The goal is to obfuscate your web traffic and make it harder to track you as a user.

![](/images/overview.gif?raw=true)

## Features
* HTTP(s) and SOCKS5 proxy support.
* No special software required. Just configure your system to use an HTTP or SOCKS proxy.
* Each AWS Lambda region provides 1 outgoing IP address that gets rotated roughly every 4 hours. That means if you use 10 AWS regions, you'll get 60 unique IPs per day.
* Configurable IP rotation frequency between multiple regions. By default IP will rotate to new region every 3 minutes.
* Personal proxy server not shared with anyone else.
* Mostly [AWS free tier](https://aws.amazon.com/free/) compatible (see FAQ below).

## Project status
Current code status: <b>proof of concept</b>. This is the first Go application that I've ever written. It has no tests. Listening endpoints have no security yet. It may not work. It may blow up. Use at your own risk.

## How it works
At a high level, awslambdaproxy proxies HTTP(s) traffic through AWS Lambda regional endpoints. To do this, awslambdaproxy is setup on a publicly accessible host (e.g. EC2 instance) and it handles creating Lambda resources that run a simple HTTP ([elazarl/goproxy](https://github.com/elazarl/goproxy)) or SOCKS ([armon/go-socks5](https://github.com/armon/go-socks5)) proxy. Since Lambda does not allow you to bind ports in your executing functions, the proxy server is bound to a unix socket and a reverse tunnel is established from the Lambda function to port 8081 on the host running awslambdaproxy. Once a tunnel connection is established, all user web traffic is forwarded from port 8080 through this proxy. Lambda functions have a max execution time of 5 minutes, so there is a goroutine that continuously executes Lambda functions to ensure there is always a live tunnel in place. If multiple regions are specified, user traffic will be routed in a round robin fashion across these regions.

![](/images/how-it-works.png?raw=true)

## Installation

### Binary
The easiest way is to download a pre-built binary from the [GitHub Releases](https://github.com/dan-v/awslambdaproxy/releases) page.

### Source
1. Fetch the project with `go get`:

  ```sh
  go get github.com/dan-v/awslambdaproxy
  cd $GOPATH/src/github.com/dan-v/awslambdaproxy
  ```

2. Install dependencies (using [Glide](https://github.com/Masterminds/glide) for dependency management)

  ```sh
  glide install
  ```

3. Run make to build awslambdaproxy. You'll find your `awslambdaproxy` binary in the `build` folder.

  ```sh
  make
  ```

## Usage

1. Copy `awslambdaproxy` binary to a publicly accessible linux host (e.g. EC2 instance). You will need to open the following ports on this host:

    * Port 8080 - this port listens for user proxy connections and needs to only be opened to whatever your external IP address is where you plan to browse the web.
    * Port 8081 - this port listens for tunnel connections from executing Lambda functions and needs to be opened to the world. <b>This is a security concern and will be locked down in the future.</b>

2. On publicly accessible host, run `awslambdaproxy`. You'll need to ensure AWS access key and secret key environment variables are defined. For now, this access key should have AdministratorAccess.

    ```sh
    export AWS_ACCESS_KEY_ID=XXXXXXXXXX
    export AWS_SECRET_ACCESS_KEY=YYYYYYYYYYYYYYYYYYYYYY
    ./awslambdaproxy -regions us-west-2,us-west-1,us-east-1,us-east-2 -proxy-type socks
    ```
    
3. Configure your web browser (or OS) to use the HTTP/HTTPS or SOCKS5 proxy (depending on -proxy-type selection) on the publicly accessible host running `awslambdaproxy` on port 8080.

## FAQ
1. <b>Should I use awslambdaproxy?</b> That's up to you. Use at your own risk.
2. <b>Why did you use AWS Lambda for this?</b> The primary reason for using AWS Lambda in this project is the vast pool of IP addresses available that automatically rotate.
3. <b>How big is the pool of available IP addresses?</b> This I don't know, but I do know I did not have a duplicate IP while running the proxy for a week.
4. <b>Will this make me completely anonymous?</b> No, absolutely not. The goal of this project is just to obfuscate your web traffic by rotating your IP address. All of your traffic is going through AWS which could be traced back to your account. You can also be tracked still with [browser fingerprinting](https://panopticlick.eff.org/), etc. Your [IP address may still leak](https://ipleak.net/) due to WebRTC, Flash, etc.
5. <b>How often will my external IP address change?</b> For each region specified, the IP address will change roughly every 4 hours. This of course is subject to change at any moment as this is not something that is documented by AWS Lambda.
6. <b>How much does this cost?</b> awslambdaproxy should be able to run mostly on the [AWS free tier](https://aws.amazon.com/free/) minus bandwidth costs. It can run on a t2.micro instance and the default 128MB Lambda function that is constantly running should also fall in the free tier usage. The bandwidth is what will cost you money; you will pay for bandwidth usage for both EC2 and Lambda.
7. <b>Why does my connection drop periodically?</b> AWS Lambda functions can currently only execute for a maximum of 5 minutes. In order to maintain an ongoing HTTP proxy a new function is executed and all new traffic is cut over to it. Any ongoing connections to previous Lambda function will hard stop after a timeout period. You generally won't see any issues for normal web browsing as connections are very short lived, but for any long lived connections you may see issues.

# Powered by
* [goproxy](https://github.com/elazarl/goproxy) - An HTTP proxy server written in Go.
* [go-socks5](https://github.com/armon/go-socks5) - A SOCKS5 proxy written in Go.
* [yamux](https://github.com/hashicorp/yamux) - Golang connection multiplexing library.
* [goad](https://github.com/goadapp/goad) - Code was borrowed from this project to handle AWS Lambda zip creation and function upload.

# Future work
* Add security to proxy and tunnel connections
* Fix connections dropping each time a new tunnel is established
* Create minimal IAM policy
* Rewrite code to be testable and write tests