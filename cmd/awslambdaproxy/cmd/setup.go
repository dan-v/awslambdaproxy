package cmd

import (
	"github.com/spf13/cobra"
	"github.com/dan-v/awslambdaproxy"
	"fmt"
	"os"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup awslambdaproxy AWS infrastructure",
	Long: `This will setup all required AWS infrastructure to run awslambdaproxy. Example:

export AWS_ACCESS_KEY_ID=XXXXXXXXXX
export AWS_SECRET_ACCESS_KEY=YYYYYYYYYYYYYYYYYYYYYY
./awslambdaproxy setup`,
	Run: func(cmd *cobra.Command, args []string) {
		err := awslambdaproxy.SetupLambdaInfrastructure()
		if err != nil {
			fmt.Print("Failed to run setup for awslambdaproxy", err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(setupCmd)
}
