package cli_client

import (
	"github.com/spf13/cobra"
	"log"
)

var tmpEventID int64

// usbCmd represents the usb command
var evtCmd = &cobra.Command{
	Use:   "EVT",
	Short: "Receive P4wnP1 service events",
	Run: func(cmd *cobra.Command, args []string) {
		err := receiveEvent(tmpEventID)
		if err != nil { log.Fatal(err)}
	},
}

func receiveEvent(eType int64) (err error) {
	return ClientRegisterEvent(StrRemoteHost, StrRemotePort, eType)
}



func init() {
	rootCmd.AddCommand(evtCmd)
	evtCmd.Flags().Int64VarP(&tmpEventID,"event-id", "i", 0,"Listen to events of given ID (0 = Any)")
}
