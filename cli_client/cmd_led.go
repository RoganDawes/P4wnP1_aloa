package cli_client

import (
	"github.com/spf13/cobra"

	pb "github.com/mame82/P4wnP1_go/proto"
	"fmt"
	"log"
)

var blink_count uint32

// usbCmd represents the usb command
var ledCmd = &cobra.Command{
	Use:   "LED",
	Short: "Set or Get LED state of P4wnP1",
}

var ledGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get LED blink count",
	Run: func(cmd *cobra.Command, args []string) {
		if ls, err := ClientGetLED(StrRemoteHost, StrRemotePort); err == nil {
			fmt.Printf("LED blink count %v\n", ls.BlinkCount)
		} else {
			log.Println(err)
		}
	},
}

var ledSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set LED blink count",
	Run: func(cmd *cobra.Command, args []string) {
		blink := cmd.Flags().Lookup("blink")
		if blink.Changed {
			if err := ClientSetLED(StrRemoteHost, StrRemotePort, pb.LEDSettings{BlinkCount: blink_count}); err == nil {
				fmt.Printf("LED blink count set to %v\n", blink.Value)
			} else {
				log.Println(err)
			}
		} else {
			cmd.Usage()
		}
	},
}


func init() {
	rootCmd.AddCommand(ledCmd)
	ledCmd.AddCommand(ledGetCmd)
	ledCmd.AddCommand(ledSetCmd)

	ledSetCmd.Flags().Uint32VarP(&blink_count,"blink", "b", 0,"Set blink count (0: Off, 1..254: blink n times, >254: On)")
}
