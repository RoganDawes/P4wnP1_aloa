package cli_client

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"

	"github.com/davecgh/go-spew/spew"
	"google.golang.org/grpc/status"
)

//Empty settings used to store cobra flags
var (
	//tmpGadgetSettings = pb.GadgetSettings{CdcEcmSettings:&pb.GadgetSettingsEthernet{},RndisSettings:&pb.GadgetSettingsEthernet{}}
	tmpNoAutoDeploy          = false
	tmpDisableGadget  bool   = false
	tmpUseHIDKeyboard uint8  = 0
	tmpUseHIDMouse    uint8  = 0
	tmpUseHIDRaw      uint8  = 0
	tmpUseRNDIS       uint8  = 0
	tmpUseECM         uint8  = 0
	tmpUseSerial      uint8  = 0
	tmpUseUMS         uint8  = 0
	tmpUMSFile        string = ""
	tmpUMSCdromMode   bool   = false

)

func init(){
	//Configure spew for struct deep printing (disable using printer interface for gRPC structs)
	spew.Config.Indent="\t"
	spew.Config.DisableMethods = true
	spew.Config.DisablePointerAddresses = true
}

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

var usbGetDeployedCmd = &cobra.Command{
	Use:   "deployed",
	Short: "Get deployed USB Gadget settings (the currently running configuration for the kernel module)",
	Long: ``,
	Run: cobraUsbGetDeployed,
}

var usbSetDeployeCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deployed the USB Gadget settings (as running configuration for the kernel module)",
	Long: ``,
	Run: cobraUsbDeploySettings,
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

	fmt.Printf("Old USB Gadget Settings:\n%s", spew.Sdump(gs))

	if tmpDisableGadget {
		fmt.Println("Setting gadget to disabled (won't get bound to UDC after deployment)")
		gs.Enabled = false
	} else {
		fmt.Println("Setting gadget to enabled (will be usable after deployment)")
		gs.Enabled = true
	}

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
		if tmpUseECM == 0 {
			fmt.Println("Disabeling CDC ECM")
			gs.Use_CDC_ECM = false
		} else {
			fmt.Println("Enabeling CDC ECM")
			gs.Use_CDC_ECM = true
		}
	}

	if (cmd.Flags().Lookup("serial").Changed) {
		if tmpUseSerial == 0 {
			fmt.Println("Disabeling Serial")
			gs.Use_SERIAL = false
		} else {
			fmt.Println("Enabeling Serial")
			gs.Use_SERIAL = true
		}
	}

	if (cmd.Flags().Lookup("hid-keyboard").Changed) {
		if tmpUseHIDKeyboard == 0 {
			fmt.Println("Disabeling HID keyboard")
			gs.Use_HID_KEYBOARD = false
		} else {
			fmt.Println("Enabeling HID keyboard")
			gs.Use_HID_KEYBOARD = true
		}
	}

	if (cmd.Flags().Lookup("hid-mouse").Changed) {
		if tmpUseHIDMouse == 0 {
			fmt.Println("Disabeling HID mouse")
			gs.Use_HID_MOUSE = false
		} else {
			fmt.Println("Enabeling HID mouse")
			gs.Use_HID_MOUSE = true
		}
	}

	if (cmd.Flags().Lookup("hid-raw").Changed) {
		if tmpUseHIDRaw == 0 {
			fmt.Println("Disabeling HID raw device")
			gs.Use_HID_RAW = false
		} else {
			fmt.Println("Enabeling HID raw device")
			gs.Use_HID_RAW = true
		}
	}

	if (cmd.Flags().Lookup("ums").Changed) {
		if tmpUseUMS == 0 {
			fmt.Println("Disabeling USB Mass Storage")
			gs.Use_UMS = false
		} else {
			fmt.Println("Enabeling USB Mass Storage")
			gs.Use_UMS = true

			gs.UmsSettings.Cdrom = tmpUMSCdromMode
			if tmpUMSCdromMode {
				fmt.Println("Setting USB Mass Storage to CD-Rom mode")
			} else {
				fmt.Println("Setting USB Mass Storage to flash drive mode")
			}

			if cmd.Flags().Lookup("ums-file").Changed {
				fmt.Printf("Serving USB Mass Storage from '%s'\n", tmpUMSFile)
				gs.UmsSettings.File = tmpUMSFile
			}
		}
	}




	//Try to set the change config
	err = ClientSetGadgetSettings(StrRemoteHost, StrRemotePort, *gs)
	if err != nil {
		//ToDo: Adopt parsing of Error Message to other gRPC calls
		log.Printf("Error setting new gadget settings: %v\n", status.Convert(err).Message())
		return
	}

	gs, err = ClientGetGadgetSettings(StrRemoteHost, StrRemotePort)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("New USB Gadget Settings:\n%s", spew.Sdump(gs))

	//if "auto deploy" isn't disabled, we deploy the gadget immediately to ConfigFS
	if !tmpNoAutoDeploy {
		fmt.Println("Auto-deploy the new settings...")
		if gs, err := ClientDeployGadgetSettings(StrRemoteHost, StrRemotePort); err != nil {
			fmt.Printf("Error deploying Gadget Settings: %v\nReverted to:\n%s", err, spew.Sdump(gs))
		} else {
			fmt.Printf("Successfully deployed:\n%s", spew.Sdump(gs))
		}
	}

	return
}

func cobraUsbGet(cmd *cobra.Command, args []string) {
	if gs, err := ClientGetGadgetSettings(StrRemoteHost, StrRemotePort); err == nil {
		fmt.Printf("USB Gadget Settings:\n%s", spew.Sdump(gs))
	} else {
		log.Println(err)
	}
}

func cobraUsbDeploySettings(cmd *cobra.Command, args []string) {

	if gs, err := ClientDeployGadgetSettings(StrRemoteHost, StrRemotePort); err != nil {
		fmt.Printf("Error deploying Gadget Settings: %v\nReverted to:\n%s", err, spew.Sdump(gs))
	} else {
		fmt.Printf("Successfully deployed:\n%s", spew.Sdump(gs))
	}

}


func cobraUsbGetDeployed(cmd *cobra.Command, args []string) {
	if gs, err := ClientGetDeployedGadgetSettings(StrRemoteHost, StrRemotePort); err == nil {
		fmt.Printf("Deployed USB Gadget Settings:\n%s", spew.Sdump(gs))
	} else {
		log.Println(err)
	}
}

func init() {
//	rootCmd.AddCommand(usbCmd)
	usbCmd.AddCommand(usbGetCmd)
	usbCmd.AddCommand(usbSetCmd)
	usbGetCmd.AddCommand(usbGetDeployedCmd)
	usbSetCmd.AddCommand(usbSetDeployeCmd)

	usbSetCmd.Flags().BoolVarP(&tmpNoAutoDeploy, "no-deploy","n", false, "If this flag is set, the gadget isn't deployed automatically (allows further changes before deployment)")
	usbSetCmd.Flags().BoolVarP(&tmpDisableGadget, "disabled","d", false, "If this flag is set, the gadget stays inactive after deployment (not bound to UDC)")
	usbSetCmd.Flags().Uint8VarP(&tmpUseRNDIS, "rndis", "r",0,"Use the RNDIS gadget function (0: disable, 1..n: enable)")
	usbSetCmd.Flags().Uint8VarP(&tmpUseECM, "cdc-ecm", "e",0,"Use the CDC ECM gadget function (0: disable, 1..n: enable)")
	usbSetCmd.Flags().Uint8VarP(&tmpUseSerial, "serial", "s",0,"Use the SERIAL gadget function (0: disable, 1..n: enable)")

	usbSetCmd.Flags().Uint8VarP(&tmpUseHIDKeyboard, "hid-keyboard", "k",0,"Use the HID KEYBOARD gadget function (0: disable, 1..n: enable)")
	usbSetCmd.Flags().Uint8VarP(&tmpUseHIDMouse, "hid-mouse", "m",0,"Use the HID MOUSE gadget function (0: disable, 1..n: enable)")
	usbSetCmd.Flags().Uint8VarP(&tmpUseHIDRaw, "hid-raw", "g",0,"Use the HID RAW gadget function (0: disable, 1..n: enable)")

	usbSetCmd.Flags().Uint8VarP(&tmpUseUMS, "ums", "u",0,"Use the USB Mass Storage gadget function (0: disable, 1..n: enable)")

	usbSetCmd.Flags().BoolVar(&tmpUMSCdromMode, "ums-cdrom", false, "If this flag is set, UMS emulates a CD-Rom instead of a flashdrive (ignored, if UMS disabled)")
	usbSetCmd.Flags().StringVar(&tmpUMSFile, "ums-file", "", "Path to the image or block device backing UMS (ignored, if UMS disabled)")
}
