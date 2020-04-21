package awslambdaproxy

import "strings"

const version = "0.0.9"

// LambdaVersion is version of awslambdaproxy
func LambdaVersion() string {
	return "v" + strings.Replace(version, ".", "-", -1)
}
