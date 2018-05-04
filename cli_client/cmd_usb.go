package cli_client

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"

)

//Empty settings used to store cobra flags
var (
	//tmpGadgetSettings = pb.GadgetSettings{CdcEcmSettings:&pb.GadgetSettingsEthernet{},RndisSettings:&pb.GadgetSettingsEthernet{}}
	tmpUseHIDKeyboard uint8 = 0
	tmpUseHIDMouse uint8 = 0
	tmpUseHIDRaw uint8 = 0
	tmpUseRNDIS uint8 = 0
	tmpUseECM uint8 = 0
	tmpUseSerial uint8 = 0
	tmpUseUMS uint8 = 0
)


// usbCmd represents the usb command
var usbCmd = &cobra.Command{
	Use:   "USB",
	Short: "Set, get or deploy USB Gadget settings",
}

var usbGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get USB Gadget settings",
	Long: ``,
	Run: cobraUsbGet,
}

var usbSetCmd = &cobra.Command{
	Use:   "set",
	Short: "set USB Gadget settings",
	Long: ``,
	Run: cobraUsbSet,
}

func cobraUsbSet(cmd *cobra.Command, args []string) {
	gs, err := ClientGetGadgetSettings(StrRemoteHost, StrRemotePort)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("USB Gadget Settings retreived: %+v\n", gs)

	if (cmd.Flags().Lookup("rndis").Changed) {
		if tmpUseRNDIS == 0 {
			fmt.Println("Disabeling RNDIS")
			gs.Use_RNDIS = false
		} else {
			fmt.Println("Enabeling RNDIS")
			gs.Use_RNDIS = true
		}
	}

	if (cmd.Flags().Lookup("cdc-ecm").Changed) {
		if tmpUseRNDIS == 0 {
			fmt.Println("Disabeling CDC ECM")
			gs.Use_CDC_ECM = false
		} else {
			fmt.Println("Enabeling RNDIS")
			gs.Use_CDC_ECM = true
		}
	}

	if (cmd.Flags().Lookup("serial").Changed) {
		if tmpUseRNDIS == 0 {
			fmt.Println("Disabeling Serial")
			gs.Use_SERIAL = false
		} else {
			fmt.Println("Enabeling Serial")
			gs.Use_SERIAL = true
		}
	}

	if (cmd.Flags().Lookup("hid-keyboard").Changed) {
		if tmpUseRNDIS == 0 {
			fmt.Println("Disabeling HID keyboard")
			gs.Use_HID_KEYBOARD = false
		} else {
			fmt.Println("Enabeling HID keyboard")
			gs.Use_HID_KEYBOARD = true
		}
	}


	//ToDo: Implement the rest (HID, UMS etc.)

	//Try to set the change config
	err = ClientSetGadgetSettings(StrRemoteHost, StrRemotePort, *gs)
	if err != nil {
		log.Printf("Error setting new gadget settings: %v\n", err)
	}

	gs, err = ClientGetGadgetSettings(StrRemoteHost, StrRemotePort)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("USB Gadget Settings set: %+v\n", gs)
	return
}

func cobraUsbGet(cmd *cobra.Command, args []string) {
	if gs, err := ClientGetGadgetSettings(StrRemoteHost, StrRemotePort); err == nil {
		fmt.Printf("USB Gadget Settings: %+v\n", gs)
	} else {
		log.Println(err)
	}
}

func init() {
	rootCmd.AddCommand(usbCmd)
	usbCmd.AddCommand(usbGetCmd)
	usbCmd.AddCommand(usbSetCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// usbCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// usbCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")


	usbSetCmd.Flags().Uint8VarP(&tmpUseRNDIS, "rndis", "n",0,"Use the RNDIS gadget function (0: disable, 1..n: enable)")
	usbSetCmd.Flags().Uint8VarP(&tmpUseECM, "cdc-ecm", "e",0,"Use the CDC ECM gadget function (0: disable, 1..n: enable)")
	usbSetCmd.Flags().Uint8VarP(&tmpUseSerial, "serial", "s",0,"Use the SERIAL gadget function (0: disable, 1..n: enable)")

	usbSetCmd.Flags().Uint8VarP(&tmpUseHIDKeyboard, "hid-keyboard", "k",0,"Use the HID KEYBOARD gadget function (0: disable, 1..n: enable)")
	usbSetCmd.Flags().Uint8VarP(&tmpUseHIDMouse, "hid-mouse", "m",0,"Use the HID MOUSE gadget function (0: disable, 1..n: enable)")
	usbSetCmd.Flags().Uint8VarP(&tmpUseHIDRaw, "hid-raw", "r",0,"Use the HID RAW gadget function (0: disable, 1..n: enable)")

	usbSetCmd.Flags().Uint8VarP(&tmpUseUMS, "ums", "u",0,"Use the USB MAss Storage gadget function (0: disable, 1..n: enable)")
}
