package awslambdaproxy

import "strings"

const Version = "0.0.1"
func LambdaVersion() string {
	return "v" + strings.Replace(Version, ".", "-", -1)
}