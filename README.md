# AWS Lambda Proxy

awslambdaproxy is an [AWS Lambda](https://aws.amazon.com/lambda/) powered HTTP proxy. This provides a constantly rotating IP address for your web traffic from any AWS region where AWS Lambda is available.

![](http://i.giphy.com/2dKZ7DEBg1NCM.gif)

## How it works
awslambdaproxy is executed on publicly accessible host (e.g. EC2 instance) and it handles creating Lambda resources, in the regions specified, that run a simple HTTP proxy ([elazarl/goproxy](https://github.com/elazarl/goproxy)). Since Lambda does not allow you to bind ports in your executing functions, the proxy is bound to a unix socket and a reverse tunnel is established from the Lambda function to port 8081 on the host running awslambdaproxy. Once a tunnel connection is established, all user web traffic is forwarded from port 8080 through this proxy. Since Lambda functions have a max execution time of 5 minutes, there is a goroutine that continuously executes Lambda functions at a constant interval to ensure there is always a live tunnel in place.

## Usage
Currently you have to compile the project yourself.

1. Fetch the project with `go get`:

  ```sh
  go get github.com/dan-v/awslambdaproxy
  ```

2. Install dependencies

  ```sh
  go get github.com/jteeuwen/go-bindata/...
  go get github.com/hashicorp/yamux
  go get github.com/pkg/errors
  go get github.com/aws/aws-sdk-go/
  go get github.com/elazarl/goproxy
  ```

3. Run make to build awslambdaproxy

  ```sh
  make
  ```

4. You'll find your `awslambdaproxy` binary in the `build` folder.

5. Copy `build/linux/x86-64/awslambdaproxy` binary to a publicly accessible linux host. You will need to open the following ports:

    * Port 8080 - this port listens for user proxy connections and needs to only be opened to whatever your external IP address is where you plan to browse the web.
    * Port 8081 - this port listens for tunnel connections from executing Lambda functions and needs to be opened to the world. This is a security concern and ideally will be locked down in the future.

6. On publicly accessible host, run `awslambdaproxy`. You'll need to ensure AWS access key and secret key environment variables are defined. For now, this access key should have AdministratorAccess.

    ```sh
    export AWS_ACCESS_KEY_ID=XXXXXXXXXXXXXXXXX
    export AWS_SECRET_ACCESS_KEY=YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY
    ./awslambdaproxy -regions us-west-2,us-east-1,us-east-2
    ```
    
7. Configure your web browser (or system) to use an HTTP proxy at the host running `awslambdaproxy` on port 8080.

## AWS costs
awslambdaproxy should be able to run mostly on the free tier minus bandwidth costs. It can run on a t2.micro instance and the default configuration 128MB Lambda functions that are created with a constantly running Lambda function should also fall in the free tier usage. The bandwidth is what will cost you money.

## Future work
* Fix connections dropping each time a new tunnel is established
* Rewrite code to be testable
* Write tests
* Add security