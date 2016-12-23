# AWS Lambda Proxy

awslambdaproxy is an [AWS Lambda](https://aws.amazon.com/lambda/) powered HTTP proxy. This provides a constantly rotating IP address for your web traffic from any AWS region where AWS Lambda is available.

![](http://i.giphy.com/2dKZ7DEBg1NCM.gif)

## Status of code
The code here is a proof of concept and a way for me to start learning Go. Beware that currently there is absolutely no security on any of the exposed endpoints.

## How it works
awslambdaproxy is executed on a host (e.g. EC2 instance) and it handles creating Lambda resources, in the regions specified, that run a simple HTTP proxy ([elazarl/goproxy](https://github.com/elazarl/goproxy)). Since Lambda does not allow you to bind ports in your executing functions, the proxy is bound to a unix socket and a reverse tunnel is established from the Lambda function to port 8081 on awslambdaproxy. Once a tunnel connection is established, all user traffic is forwarded from port 8080 through this proxy. Since Lambda functions have a max execution time of 5 minutes, there is a goroutine that continuously executes Lambda functions at a constant interval to ensure there is always a live tunnel in place.

## Usage
Currently you have to compile the project yourself..

1. Fetch the project with `go get`:

  ```sh
  go get github.com/dan-v/awslambdaproxy
  ```

2. Install dependencies

  ```sh
  go get github.com/jteeuwen/go-bindata/...
  go get github.com/hashicorp/yamux
  go get github.com/pkg/errors
  ```

3. Run make to build

  ```sh
  make
  ```

4. You'll find your `awslambdaproxy` binary in the `build` folder…

5. Copy `awslambdaproxy` binary to a publicly accessible host. You will need to open ports `8080` and `8081` on this host to the world.

6. On publicly accessible host, run `awslambdaproxy`. You'll need to ensure AWS access key and secret key environment variables are defined.

    ```sh
    AWS_ACCESS_KEY_ID=XXXXXXXXXXXXXXXXX AWS_SECRET_ACCESS_KEY=YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY+ ./awslambdaproxy -regions us-west-2,us-east-1,us-east-2
    ```
    
7. Configure your web browser to point at the host running `awslambdaproxy` on port 8080.

8. Profit

## $$ AWS Costs $$
awslambdaproxy should be able to run mostly on the free tier minus bandiwdth costs. It can run on a tier t2.micro instance and the default configuration 128MB Lambda functions that are created with a constantly running Lambda function should also fall in the free tier usage. The bandwidth is what will cost you money.

## Future work
* Rewrite code to be testable
* Write tests
* Fix connections dropping each time a new tunnel is established