<b>awslambdaproxy</b> is an [AWS Lambda](https://aws.amazon.com/lambda/) powered HTTP(s) proxy. It provides a constantly rotating IP address for your web traffic from any region where AWS Lambda is available.

![](/images/overview.gif?raw=true)

## How it works
At a high level, awslambdaproxy proxies HTTP(s) traffic through AWS Lambda regional endpoints. To do this, awslambdaproxy is setup on a publicly accessible host (e.g. EC2 instance) and it handles creating Lambda resources, in the regions specified, that run a simple HTTP proxy ([elazarl/goproxy](https://github.com/elazarl/goproxy)). Since Lambda does not allow you to bind ports in your executing functions, the HTTP proxy is bound to a unix socket and a reverse tunnel is established from the Lambda function to port 8081 on the host running awslambdaproxy. Once a tunnel connection is established, all user web traffic is forwarded from port 8080 through this HTTP proxy. Lambda functions have a max execution time of 5 minutes, so there is a goroutine that continuously executes Lambda functions at a constant frequency to ensure there is always a live tunnel in place. If multiple regions are specified, user HTTP traffic will be routed in a round robin fashion across these regions.

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
  brew install glide
  glide install
  ```

3. Run make to build awslambdaproxy. You'll find your `awslambdaproxy` binary in the `build` folder.

  ```sh
  make
  ```

## Usage

1. Copy `awslambdaproxy` binary to a publicly accessible linux host. You will need to open the following ports:

    * Port 8080 - this port listens for user proxy connections and needs to only be opened to whatever your external IP address is where you plan to browse the web.
    * Port 8081 - this port listens for tunnel connections from executing Lambda functions and needs to be opened to the world. This is a security concern and ideally will be locked down in the future.

2. On publicly accessible host, run `awslambdaproxy`. You'll need to ensure AWS access key and secret key environment variables are defined. For now, this access key should have AdministratorAccess.

    ```sh
    export AWS_ACCESS_KEY_ID=XXXXXXXXXX
    export AWS_SECRET_ACCESS_KEY=YYYYYYYYYYYYYYYYYYYYYY
    ./awslambdaproxy -regions us-west-2,us-west-1,us-east-1,us-east-2 -frequency 10
    ```
    
3. Configure your web browser (or OS) to use an HTTP proxy at the publicly accessible host running `awslambdaproxy` on port 8080.

## FAQ
1. <b>How often will the external IP address change?</b> For each region specified, the IP address will change roughly every 4 hours. This of course is subject to change at any moment as this is not something that is documented for the AWS Lambda service.
2. <b>How much does this cost?</b> awslambdaproxy should be able to run mostly on the AWS free tier minus bandwidth costs. It can run on a t2.micro instance and the default configuration 128MB Lambda functions that are created with a constantly running Lambda function should also fall in the free tier usage. The bandwidth is what will cost you money. Use at your own risk.

## Future work
* Fix connections dropping each time a new tunnel is established
* Rewrite code to be testable
* Write tests
* Add security
