SHELL := /bin/bash
TARGET := awslambdaproxy
VERSION := $(shell cat VERSION)
OS := linux
ARCH := amd64
PACKAGE := github.com/dan-v/$(TARGET)

.PHONY: \
	help \
	clean \
	clean-artifacts \
	tools \
	test \
	coverage \
	vet \
	lint \
	fmt \
	build \
	build-lambda \
	build-server \
	doc \
	version \
	release

all: tools fmt lint vet build release

help:
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@echo '    help               Show this help screen.'
	@echo '    clean              Remove binaries, artifacts and releases.'
	@echo '    tools              Install tools needed by the project.'
	@echo '    test               Run unit tests.'
	@echo '    coverage           Report code tests coverage.'
	@echo '    vet                Run go vet.'
	@echo '    lint               Run golint.'
	@echo '    fmt                Run go fmt.'
	@echo '    build              Build all.'
	@echo '    build-lambda       Build lambda function.'
	@echo '    build-server       Build server.'
	@echo '    release            Zip up final artifact'
	@echo '    doc                Start Go documentation server on port 8080.'
	@echo '    version            Display Go version.'
	@echo ''

print-%:
	@echo $* = $($*)

clean:
	rm -Rf artifacts
	rm -vf $(CURDIR)/coverage.*

tools:
	go get golang.org/x/lint/golint
	go get github.com/axw/gocov/gocov
	go get github.com/matm/gocov-html

test:
	go test -v ./...

coverage:
	gocov test ./... > $(CURDIR)/coverage.out 2>/dev/null
	gocov report $(CURDIR)/coverage.out
	if test -z "$$CI"; then \
	  gocov-html $(CURDIR)/coverage.out > $(CURDIR)/coverage.html; \
	  if which open &>/dev/null; then \
	    open $(CURDIR)/coverage.html; \
	  fi; \
	fi

vet:
	go vet -v ./...

lint:
	golint $(go list ./... | grep -v /vendor/)

fmt:
	go fmt ./...

build-lambda:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o artifacts/lambda/main ./pkg/lambda
	zip -jr artifacts/lambda artifacts/lambda
	go-bindata -nocompress -pkg server -o pkg/server/bindata.go artifacts/lambda.zip

build-server:
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build -ldflags \
	    "-X $(PACKAGE)/cmd/awslambdaproxy.version=$(VERSION)" \
	    -v -o $(CURDIR)/artifacts/server/$(OS)/$(TARGET) ./cmd/main.go

build-server-osx:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags \
	    "-X $(PACKAGE)/cmd/awslambdaproxy.version=$(VERSION)" \
	    -v -o $(CURDIR)/artifacts/server/darwin/$(TARGET) ./cmd/main.go

build: build-lambda build-server

build-osx: build-lambda build-server-osx

doc:
	godoc -http=:8080 -index

version:
	@go version

release:
	mkdir -p ./artifacts
	zip -jr ./artifacts/$(TARGET)-$(OS)-$(VERSION).zip ./artifacts/server/$(OS)/$(TARGET)