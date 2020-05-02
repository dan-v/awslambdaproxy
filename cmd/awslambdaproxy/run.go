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
	frequency                                                               time.Duration
	memory                                                                  int
	debug, debugProxy                                                       bool
	lambdaName, lambdaIamRole, sshUser, sshPort, regions, listeners, bypass string
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run awslambdaproxy",
	Long: `this will execute awslambdaproxy in the specified regions. examples:

# execute proxy in four different regions with rotation happening every 60 seconds
./awslambdaproxy run -r us-west-2,us-west-1,us-east-1,us-east-2 -f 60s

# choose a different port and username/password for proxy and add another listener on localhost with no auth
./awslambdaproxy run -l "admin:admin@:8888,localhost:9090"

# bypass certain domains from using lambda proxy
./awslambdaproxy run -b "*.websocket.org,*.youtube.com"

# specify a dns server for the proxy server to use for dns lookups
./awslambdaproxy run -l "admin:awslambdaproxy@:8080?dns=1.1.1.1"

# increase function memory size for better network performance
./awslambdaproxy run -m 512
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
		aBypass := viper.GetString("bypass")
		aLambdaName := viper.GetString("lambda-name")
		aLambdaIamRoleName := viper.GetString("lambda-iam-role-name")

		if _, err := server.GetSessionAWS(); err != nil {
			log.Fatal("unable to find valid aws credentials")
		}

		s, err := server.New(server.Config{
			LambdaName:               aLambdaName,
			LambdaIamRoleName:        aLambdaIamRoleName,
			LambdaRegions:            aRegions,
			LambdaMemory:             aMemory,
			LambdaExecutionFrequency: aFrequency,
			ProxyListeners:           aListeners,
			ProxyDebug:               aDebugProxy,
			ReverseTunnelSSHUser:     aSSHUser,
			ReverseTunnelSSHPort:     aSSHPort,
			Bypass:                   aBypass,
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

	runCmd.Flags().StringVarP(&lambdaName, "lambda-name", "n", "awslambdaproxy",
		fmt.Sprintf("name of lambda function"))
	runCmd.Flags().StringVarP(&lambdaIamRole, "lambda-iam-role-name", "i", "awslambdaproxy-role",
		fmt.Sprintf("name of lambda function"))
	runCmd.Flags().StringVarP(&regions, "regions", "r", "us-west-2",
		fmt.Sprintf("comma separted list of regions to run proxy (e.g. us-west-2,us-west-1,us-east-1). "+
			"valid regions include %v", server.GetValidLambdaRegions()))
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
	runCmd.Flags().StringVarP(&bypass, "bypass", "b", "",
		"comma separated list of domains/ips to bypass lambda proxy (e.g. *.websocket.org,*.youtube.com). "+
			"note that when using sock5 proxy mode you'll need to be remotely resolving dns for this to work.")

	viper.BindPFlag("lambda-name", runCmd.Flags().Lookup("lambda-name"))
	viper.BindPFlag("lambda-iam-role-name", runCmd.Flags().Lookup("lambda-iam-role-name"))
	viper.BindPFlag("regions", runCmd.Flags().Lookup("regions"))
	viper.BindPFlag("frequency", runCmd.Flags().Lookup("frequency"))
	viper.BindPFlag("memory", runCmd.Flags().Lookup("memory"))
	viper.BindPFlag("ssh-user", runCmd.Flags().Lookup("ssh-user"))
	viper.BindPFlag("ssh-port", runCmd.Flags().Lookup("ssh-port"))
	viper.BindPFlag("listeners", runCmd.Flags().Lookup("listeners"))
	viper.BindPFlag("debug-proxy", runCmd.Flags().Lookup("debug-proxy"))
	viper.BindPFlag("debug", runCmd.Flags().Lookup("debug"))
	viper.BindPFlag("bypass", runCmd.Flags().Lookup("bypass"))
}
