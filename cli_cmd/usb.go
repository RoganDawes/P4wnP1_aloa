package cli_cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// usbCmd represents the usb command
var usbCmd = &cobra.Command{
	Use:   "usb",
	Short: "Set or get USB Gadget settings",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("usb called")
	},
}

var usbGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get USB Gadget settings",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("usb get called %v", args)
	},
}

func init() {
	rootCmd.AddCommand(usbCmd)
	usbCmd.AddCommand(usbGetCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// usbCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// usbCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
