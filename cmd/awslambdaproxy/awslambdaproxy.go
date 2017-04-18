package main

import (
	"fmt"
	"os"
	"github.com/dan-v/awslambdaproxy/cmd/awslambdaproxy/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}