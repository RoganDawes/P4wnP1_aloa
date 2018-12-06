package cli_client

import (
	"github.com/spf13/cobra"
	"os"

	pb "github.com/mame82/P4wnP1_aloa/proto"
	"fmt"
	"log"
)

var blink_count uint32

// usbCmd represents the usb command
var ledCmd = &cobra.Command{
	Use:   "led",
	Short: "Set or Get LED state of P4wnP1",
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flags().Lookup("blink").Changed {
			// blink flag has been set
			if err := ClientSetLED(StrRemoteHost, StrRemotePort, pb.LEDSettings{BlinkCount: blink_count}); err == nil {
				fmt.Printf("LED blink count set to %v\n", blink_count)
			} else {
				log.Println(err)
				os.Exit(-1)
			}
		} else {
			// blink flag has not been set, retrieve current blink count
			if ls, err := ClientGetLED(StrRemoteHost, StrRemotePort); err == nil {
				fmt.Printf("LED blink count %v\n", ls.BlinkCount)
			} else {
				log.Println(err)
				os.Exit(-1)
			}
		}
	},
}


func init() {
	rootCmd.AddCommand(ledCmd)

	ledCmd.Flags().Uint32VarP(&blink_count,"blink", "b", 0,"Set blink count (0: Off, 1..254: blink n times, >254: On)")
}
