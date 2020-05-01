package awslambdaproxy

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version string
	cfgFile string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:     "awslambdaproxy",
	Short:   "An AWS Lambda powered HTTP/SOCKS web proxy",
	Version: version,
	Long: `awslambdaproxy is an AWS Lambda powered HTTP/SOCKS web proxy.
It provides a constantly rotating IP address for your network traffic
from all regions where AWS Lambda is available. The goal is to obfuscate
your connections and make it harder to track you as a user.`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	var replacer = strings.NewReplacer("-", "_", ".", "_")
	viper.SetConfigName(".awslambdaproxy") // name of config file (without extension)
	viper.AddConfigPath("$HOME")           // adding home directory as first search path
	viper.AutomaticEnv()                   // read in environment variables that match
	viper.SetEnvKeyReplacer(replacer)

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
