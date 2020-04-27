all: linux

lambda:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o data/lambda/main ./lambda
	zip -jr data/lambda data/lambda

bindata: lambda
	go-bindata -nocompress -pkg awslambdaproxy -o bindata.go data/lambda.zip

linux: bindata
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./build/linux/x86-64/awslambdaproxy ./cmd/awslambdaproxy

osx:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ./build/osx/x86-64/awslambdaproxy ./cmd/awslambdaproxy

clean:
	rm -rf data/lambda/awslambdaproxy-lambda
	rm -rf data/lambda.zip
	rm -rf build
	rm -rf bindata.go

all-zip: all
	mkdir ./build/zip
	zip -jr ./build/zip/awslambdaproxy-linux-x86-64 ./build/linux/x86-64/awslambdaproxy
	cp data/lambda.zip ./build/zip/

.PHONY: lambda bindata
