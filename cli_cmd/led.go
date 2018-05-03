package cli_cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// usbCmd represents the usb command
var ledCmd = &cobra.Command{
	Use:   "led",
	Short: "Set LED of P4wnP1",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("usb called")
	},
}


func init() {
	rootCmd.AddCommand(ledCmd)


	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// usbCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// usbCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
