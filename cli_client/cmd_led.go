package cli_client

import (
	"github.com/spf13/cobra"

	pb "../proto"
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// usbCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// usbCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	ledSetCmd.Flags().Uint32Var(&blink_count,"blink", 0,"Set blink count (0: Off, 1..254: blink n times, >254: On)")
}
