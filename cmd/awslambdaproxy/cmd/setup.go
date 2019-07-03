package cmd

import (
	"fmt"
	"os"

	"github.com/dan-v/awslambdaproxy"
	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup awslambdaproxy AWS infrastructure",
	Long:  `This will setup all required AWS infrastructure to run awslambdaproxy.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := awslambdaproxy.SetupLambdaInfrastructure()
		if err != nil {
			fmt.Print("Failed to run setup for awslambdaproxy: ", err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(setupCmd)
}
