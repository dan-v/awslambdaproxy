all: osx linux

lambda:
	CC=${CC} GOOS=linux GOARCH=amd64 go build --ldflags "${GOFLAGS}" -o data/lambda/main ./lambda
	zip -jr data/lambda data/lambda

bindata: lambda
	go get github.com/jteeuwen/go-bindata/...
	go-bindata -nocompress -pkg awslambdaproxy -o bindata.go data/lambda.zip

linux: bindata
	CC=${CC} GOOS=linux GOARCH=amd64 go build --ldflags "${GOFLAGS}" -o ./build/linux/x86-64/awslambdaproxy${LDTAIL} ./cmd/awslambdaproxy${BINTAIL}

linux-static-musl:  CC = musl-gcc
linux-static-musl:  GOFLAGS = -s -w -linkmode external -extldflags '-static'
linux-static-musl:  BINTAIL = -musl
linux-static-musl: bindata linux

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
