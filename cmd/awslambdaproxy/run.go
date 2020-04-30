package awslambdaproxy

import (
	"fmt"
	"log"
	"os/user"
	"strings"
	"time"

	"github.com/dan-v/awslambdaproxy/pkg/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	frequency                            time.Duration
	memory                               int
	debug, debugProxy                    bool
	sshUser, sshPort, regions, listeners string
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run awslambdaproxy",
	Long: `This will execute awslambdaproxy in regions specified. Examples:

# Example 1 - Execute proxy in four different regions with rotation happening every 60 seconds
./awslambdaproxy run -r us-west-2,us-west-1,us-east-1,us-east-2 -f 60s

# Example 2 - Choose a different port and username/password for proxy and add another listener on localhost with no auth
./awslambdaproxy run -r us-west-2 -l "admin:admin@:8888,localhost:9090"

# Example 3 - Increase function memory size for better network performance
./awslambdaproxy run -r us-west-2 -m 512
`,
	Run: func(cmd *cobra.Command, args []string) {
		aDebug := viper.GetBool("debug")
		aDebugProxy := viper.GetBool("debug-proxy")
		aSSHUser := viper.GetString("ssh-user")
		aSSHPort := viper.GetString("ssh-port")
		aRegions := strings.Split(viper.GetString("regions"), ",")
		aMemory := viper.GetInt("memory")
		aFrequency := viper.GetDuration("frequency")
		aListeners := strings.Split(viper.GetString("listeners"), ",")

		if _, err := server.GetSessionAWS(); err != nil {
			log.Fatal("unable to find valid aws credentials")
		}

		s, err := server.New(server.Config{
			LambdaRegions:            aRegions,
			LambdaMemory:             aMemory,
			LambdaExecutionFrequency: aFrequency,
			ProxyListeners:           aListeners,
			ProxyDebug:               aDebugProxy,
			ReverseTunnelSSHUser:     aSSHUser,
			ReverseTunnelSSHPort:     aSSHPort,
			Debug:                    aDebug,
		})
		if err != nil {
			log.Fatal(err)
		}
		s.Run()
	},
}

func getCurrentUserName() string {
	u, _ := user.Current()
	if u != nil {
		return u.Username
	}
	return ""
}

func init() {
	RootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&regions, "regions", "r", "us-west-2",
		fmt.Sprintf("regions to run proxy. valid regions include %v", server.GetValidLambdaRegions()))
	runCmd.Flags().DurationVarP(&frequency, "frequency", "f", server.LambdaMaxExecutionFrequency,
		fmt.Sprintf("frequency to execute Lambda function. minimum is %v and maximum is %v. "+
			"if multiple regions are specified, this will cause traffic to rotate round robin at the interval "+
			"specified here", server.LambdaMinExecutionFrequency.String(), server.LambdaMaxExecutionFrequency.String()))
	runCmd.Flags().IntVarP(&memory, "memory", "m", 128,
		fmt.Sprintf("memory size in MB for lambda function. minimum is %v and maximum is %v. "+
			"higher memory size may allow for faster network throughput.",
			server.LambdaMinMemorySize, server.LambdaMaxMemorySize))
	runCmd.Flags().StringVarP(&listeners, "listeners", "l", "admin:awslambdaproxy@:8080",
		"defines the listening port and authentication details in form [scheme://][user:pass@host]:port. "+
			"add as many listeners as you'd like. see documentation for gost for more details "+
			"https://github.com/ginuerzh/gost/blob/master/README_en.md#getting-started.")
	runCmd.Flags().StringVarP(&sshUser, "ssh-user", "", getCurrentUserName(),
		"ssh user for tunnel connections from lambda.")
	runCmd.Flags().StringVarP(&sshPort, "ssh-port", "", "22",
		"ssh port for tunnel connections from lambda.")
	runCmd.Flags().BoolVar(&debugProxy, "debug-proxy", false,
		"enable debug logging for proxy (note: this will log your visited domains)")
	runCmd.Flags().BoolVar(&debug, "debug", false,
		"enable general debug logging")

	viper.BindPFlag("regions", runCmd.Flags().Lookup("regions"))
	viper.BindPFlag("frequency", runCmd.Flags().Lookup("frequency"))
	viper.BindPFlag("memory", runCmd.Flags().Lookup("memory"))
	viper.BindPFlag("ssh-user", runCmd.Flags().Lookup("ssh-user"))
	viper.BindPFlag("ssh-port", runCmd.Flags().Lookup("ssh-port"))
	viper.BindPFlag("listeners", runCmd.Flags().Lookup("listeners"))
	viper.BindPFlag("debug-proxy", runCmd.Flags().Lookup("debug-proxy"))
	viper.BindPFlag("debug", runCmd.Flags().Lookup("debug"))
}
