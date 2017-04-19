package cmd

import (
	"fmt"
	"github.com/dan-v/awslambdaproxy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"
)

var (
	frequency                            time.Duration
	memory                               int64
	sshUser, sshPort, regions, listeners string
	// Max execution time on lambda is 300 seconds currently
	lambdaMaxFrequency  = time.Duration(290 * time.Second) // leave 10 seconds of leeway
	lambdaMinMemorySize = 128
	lambdaMaxMemorySize = 1536
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run awslambdaproxy",
	Long: `This will execute awslambdaproxy in regions specified. Examples:

# Make sure credentials are exported
export AWS_ACCESS_KEY_ID=XXXXXXXXXX
export AWS_SECRET_ACCESS_KEY=YYYYYYYYYYYYYYYYYYYYYY

# Example 1 - Execute proxy in four different regions with rotation happening every 60 seconds
./awslambdaproxy run -r us-west-2,us-west-1,us-east-1,us-east-2 -f 60s

# Example 2 - Choose a different port and username/password for proxy and add another listener on localhost with no auth
./awslambdaproxy run -r us-west-2 -l "admin:admin@:8888,localhost:9090"

# Example 3 - Increase function memory size for better network performance
./awslambdaproxy run -r us-west-2 -m 512
`,
	Run: func(cmd *cobra.Command, args []string) {
		aSshUser := viper.GetString("ssh-user")
		aSshPort := viper.GetString("ssh-port")
		aRegions := strings.Split(viper.GetString("regions"), ",")
		aMemory := viper.GetInt64("memory")
		aFrequency := viper.GetDuration("frequency")
		aListeners := strings.Split(viper.GetString("listeners"), ",")
		aTimeout := int64(viper.GetDuration("frequency").Seconds()) + int64(30)

		// check memory
		if aMemory > int64(lambdaMaxMemorySize) {
			fmt.Println("Maximum lambda memory size is " + strconv.Itoa(lambdaMaxMemorySize) + " MB")
			os.Exit(1)
		}
		if aMemory < int64(lambdaMinMemorySize) {
			fmt.Println("Minimum lambda memory size is " + strconv.Itoa(lambdaMinMemorySize) + " MB")
			os.Exit(1)
		}

		// check frequency
		if aFrequency > lambdaMaxFrequency {
			fmt.Println("Maximum lambda frequency is " + lambdaMaxFrequency.String() + " seconds")
			os.Exit(1)
		}

		// check for required aws keys
		access := os.Getenv("AWS_ACCESS_KEY_ID")
		if access == "" {
			fmt.Println("Must specify environment variable AWS_ACCESS_KEY_ID")
			os.Exit(1)
		}
		secret := os.Getenv("AWS_SECRET_ACCESS_KEY")
		if secret == "" {
			fmt.Println("Must specify environment variable AWS_SECRET_ACCESS_KEY")
			os.Exit(1)
		}

		awslambdaproxy.ServerInit(aSshUser, aSshPort, aRegions, aMemory, aFrequency, aListeners, aTimeout)
	},
}

func getCurrentUserName() string {
	user, _ := user.Current()
	if user != nil {
		return user.Username
	}
	return ""
}

func init() {
	RootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&regions, "regions", "r", "us-west-2", "Regions to "+
		"run proxy.")
	runCmd.Flags().DurationVarP(&frequency, "frequency", "f", time.Duration(time.Second*260), "Frequency "+
		"to execute Lambda function.  Maximum is "+lambdaMaxFrequency.String()+". If multiple "+
		"lambda-regions are specified, this will cause traffic to rotate round robin at the interval "+
		"specified here.")
	runCmd.Flags().Int64VarP(&memory, "memory", "m", 128, "Memory size in MB for Lambda function. "+
		"Higher memory may allow for faster network throughput.")
	runCmd.Flags().StringVarP(&listeners, "listeners", "l", "admin:awslambdaproxy@:8080", "Add as many listeners"+
		"as you'd like.")
	runCmd.Flags().StringVarP(&sshUser, "ssh-user", "", getCurrentUserName(), "SSH user for tunnel "+
		"connections from Lambda.")
	runCmd.Flags().StringVarP(&sshPort, "ssh-port", "", "22", "SSH port for tunnel "+
		"connections from Lambda.")

	viper.BindPFlag("regions", runCmd.Flags().Lookup("regions"))
	viper.BindPFlag("frequency", runCmd.Flags().Lookup("frequency"))
	viper.BindPFlag("memory", runCmd.Flags().Lookup("memory"))
	viper.BindPFlag("ssh-user", runCmd.Flags().Lookup("ssh-user"))
	viper.BindPFlag("ssh-port", runCmd.Flags().Lookup("ssh-port"))
	viper.BindPFlag("listeners", runCmd.Flags().Lookup("listeners"))
}
