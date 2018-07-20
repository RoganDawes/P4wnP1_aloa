package cli_client

import (
	"github.com/spf13/cobra"
"../common"
	"log"
)

// usbCmd represents the usb command
var evtCmd = &cobra.Command{
	Use:   "EVT",
	Short: "Receive P4wnP1 service events",
	Run: func(cmd *cobra.Command, args []string) {
		err := receiveEvent(common.EVT_LOG)
		if err != nil { log.Fatal(err)}
	},
}

func receiveEvent(eType int64) (err error) {
	return ClientRegisterEvent(StrRemoteHost, StrRemotePort, eType)
}



func init() {
	rootCmd.AddCommand(evtCmd)
}
