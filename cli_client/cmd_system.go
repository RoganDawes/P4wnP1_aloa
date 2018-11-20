package cli_client

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)




func init() {
	cmdSystem := &cobra.Command{
		Use:   "system",
		Short: "system commands",
	}

	cmdSystemReboot := &cobra.Command{
		Use:   "reboot",
		Short: "reboot P4wnP1",
		Run: func(cmd *cobra.Command, args []string) {
			err := ClientReboot(StrRemoteHost, StrRemotePort, TIMEOUT_LONG)
			if err != nil {
				fmt.Println(" failed")
				fmt.Println(err.Error())
				os.Exit(-1)
			}
			fmt.Println(" success")
		},

	}

	cmdSystemShutdown := &cobra.Command{
		Use:   "shutdown",
		Short: "shutdown P4wnP1",
		Run: func(cmd *cobra.Command, args []string) {
			err := ClientReboot(StrRemoteHost, StrRemotePort, TIMEOUT_LONG)
			if err != nil {
				fmt.Println(" failed")
				fmt.Println(err.Error())
				os.Exit(-1)
			}
			fmt.Println(" success")
		},

	}


	rootCmd.AddCommand(cmdSystem)
	cmdSystem.AddCommand(cmdSystemReboot, cmdSystemShutdown)
}
