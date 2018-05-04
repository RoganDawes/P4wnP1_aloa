package cli_client

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	StrRemoteHost string
	StrRemotePort string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cli_client",
	Short: "P4wnP1 (remote) CLI configuration",
	Long: `The cli_client tool could be used to configure P4wnP1
from the command line. The tool relies on RPC so it could be used 
remotely.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&StrRemoteHost, "host", "localhost", "The host with the listening P4wnP1 RPC server")
	rootCmd.PersistentFlags().StringVar(&StrRemotePort, "port", "50051", "The port on which the P4wnP1 RPC server is listening")

	/*
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	*/
}
