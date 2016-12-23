all: osx linux

lambda:
	GOOS=linux GOARCH=amd64 go build -o data/lambda/awslambdaproxy-lambda ./lambda
	zip -jr data/lambda data/lambda

bindata: lambda
	go get github.com/jteeuwen/go-bindata/...
	go-bindata -nocompress -pkg awslambdaproxy -o bindata.go data/lambda.zip

linux: bindata
	GOOS=linux GOARCH=amd64 go build -o ./build/linux/x86-64/awslambdaproxy ./cmd/awslambdaproxy

osx: bindata
	GOOS=darwin GOARCH=amd64 go build -o ./build/osx/x86-64/awslambdaproxy ./cmd/awslambdaproxy

clean:
	rm -rf data/lambda/awslambdaproxy-lambda
	rm -rf data/lambda.zip
	rm -rf build
	rm -rf bindata.go

all-zip: all
	mkdir ./build/zip
	zip -jr ./build/zip/awslambdaproxy-osx-x86-64 ./build/osx/x86-64/awslambdaproxy
	zip -jr ./build/zip/awslambdaproxy-linux-x86-64 ./build/linux/x86-64/awslambdaproxy

.PHONY: lambda bindata