package awslambdaproxy

import (
	"log"

	"github.com/dan-v/awslambdaproxy/pkg/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "setup awslambdaproxy aws infrastructure",
	Long:  `this will setup all required aws infrastructure to run awslambdaproxy.`,
	Run: func(cmd *cobra.Command, args []string) {

		aLambdaIamRoleName := viper.GetString("lambda-iam-role-name")

		if _, err := server.GetSessionAWS(); err != nil {
			log.Fatal("unable to find valid aws credentials")
		}

		err := server.SetupLambdaInfrastructure(aLambdaIamRoleName)
		if err != nil {
			log.Fatal("failed to run setup for awslambdaproxy: ", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(setupCmd)
}
