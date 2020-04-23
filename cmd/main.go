package main

import (
	"fmt"
	"github.com/dan-v/awslambdaproxy/cmd/awslambdaproxy"
	"os"
)

func main() {
	if err := awslambdaproxy.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
