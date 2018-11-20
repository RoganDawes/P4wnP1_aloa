package cli_client

import (
	"encoding/json"
	"fmt"
	pb "github.com/mame82/P4wnP1_go/proto"
	"github.com/spf13/cobra"
	"log"
	"os"
)

type devPath int

const (
	dev_path_hid_keyboard devPath = iota
	dev_path_hid_mouse
	dev_path_hid_raw
	dev_path_hid_all
)

var (
	tmpUsbDisableGadget  = false
	tmpUsbUseHIDKeyboard = false
	tmpUsbUseHIDMouse    = false
	tmpUsbUseHIDRaw      = false
	tmpUsbUseRNDIS       = false
	tmpUsbUseECM         = false
	tmpUsbUseSerial      = false
	tmpUsbUseUMS         = false
	tmpUsbUMSFile        = ""
	tmpUsbUMSCdromMode   = false

	tmpUsbSerialnumber   = "deadbeef1337"
	tmpUsbVid            = "0x1d6b"
	tmpUsbPid            = "0x1347"
	tmpUsbManufacturer   = "MaMe82"
	tmpUsbProduct        = "P4wnP1 by MaMe82"


)


func PrintGadgetSettings(gs *pb.GadgetSettings, useJson bool) {
	res := ""
	if useJson {
		b, err := json.Marshal(gs)
		if err == nil {
			res = string(b)
		}
	} else {
//			res = "Composite Gadget\n"
			res += fmt.Sprintf("Enabled:      %v\n", gs.Enabled)
			res += fmt.Sprintf("Product:      %s\n", gs.Product)
			res += fmt.Sprintf("Manufacturer: %s\n", gs.Manufacturer)
			res += fmt.Sprintf("Serialnumber: %s\n", gs.Serial)
			res += fmt.Sprintf("PID:          %s\n", gs.Pid)
			res += fmt.Sprintf("VID:          %s\n", gs.Vid)
			res += "\n"
			res += fmt.Sprintf("Functions:\n")
			res += fmt.Sprintf("    RNDIS:        %v\n", gs.Use_RNDIS)
			res += fmt.Sprintf("    CDC ECM:      %v\n", gs.Use_CDC_ECM)
			res += fmt.Sprintf("    Serial:       %v\n", gs.Use_SERIAL)
			res += fmt.Sprintf("    HID Mouse:    %v\n", gs.Use_HID_MOUSE)
			res += fmt.Sprintf("    HID Keyboard: %v\n", gs.Use_HID_KEYBOARD)
			res += fmt.Sprintf("    HID Generic:  %v\n", gs.Use_HID_RAW)
			res += fmt.Sprintf("    Mass Storage: %v\n", gs.Use_UMS)

			if gs.Use_UMS {
				if gs.UmsSettings.Cdrom {
					res += fmt.Sprintf("    ---- Storage Mode: CD-Rom\n")
				} else {
					res += fmt.Sprintf("    ---- Storage Mode: Flashdrive\n")
				}
				res += fmt.Sprintf("    ---- Storage File: %s\n", gs.UmsSettings.File)
			}

	}
	fmt.Println(res)
}

func usbSet(cmd *cobra.Command, args []string) {
	newGs, err := ClientGetDeployedGadgetSettings(StrRemoteHost, StrRemotePort)
	if err != nil {
		log.Println(err)
		return
	}

	fPid := cmd.Flags().Lookup("pid")
	fVid := cmd.Flags().Lookup("vid")
	fProduct := cmd.Flags().Lookup("product")
	fManufacturer := cmd.Flags().Lookup("manufacturer")
	fSerialno := cmd.Flags().Lookup("sn")

	if fVid.Changed {
		newGs.Vid = tmpUsbVid
	}
	if fPid.Changed {
		newGs.Pid = tmpUsbPid
	}
	if fManufacturer.Changed {
		newGs.Manufacturer = tmpUsbManufacturer
	}
	if fProduct.Changed {
		newGs.Product = tmpUsbProduct
	}
	if fSerialno.Changed {
		newGs.Serial = tmpUsbSerialnumber
	}


	newGs.Enabled = !tmpUsbDisableGadget
	newGs.Use_RNDIS = tmpUsbUseRNDIS
	newGs.Use_CDC_ECM = tmpUsbUseECM
	newGs.Use_SERIAL = tmpUsbUseSerial
	newGs.Use_HID_KEYBOARD = tmpUsbUseHIDKeyboard
	newGs.Use_HID_MOUSE = tmpUsbUseHIDMouse
	newGs.Use_HID_RAW = tmpUsbUseHIDRaw
	newGs.Use_UMS = tmpUsbUseUMS


	if tmpUsbUseUMS {
		newGs.UmsSettings.Cdrom = tmpUsbUMSCdromMode
		if cmd.Flags().Lookup("ums-file").Changed {
			fmt.Printf("Serving USB Mass Storage from '%s'\n", tmpUsbUMSFile)
			newGs.UmsSettings.File = tmpUsbUMSFile
		}
	}

	//Update service settings
	deployedGs,err := ClientDeployGadgetSettings(StrRemoteHost, StrRemotePort, newGs)
	if err != nil {
		fmt.Printf("Error deploying Gadget Settings: %v\nReverted to:\n%+v", err, deployedGs)
		os.Exit(-1)
		return
	}


	if BoolJson {
		PrintGadgetSettings(deployedGs,true)
	} else {
		fmt.Println("Successfully deployed USB gadget settings")
		PrintGadgetSettings(deployedGs,false)
	}

	return
}

func usbMount(cmd *cobra.Command, args []string) {
	gs, err := ClientGetDeployedGadgetSettings(StrRemoteHost, StrRemotePort)
	if err != nil {
		log.Println(err)
		return
	}


	if gs.Use_UMS {
		//gs.UmsSettings.Cdrom = tmpUsbUMSCdromMode
		if cmd.Flags().Lookup("ums-file").Changed {
			fmt.Printf("Serving USB Mass Storage from '%s'\n", tmpUsbUMSFile)
			//gs.UmsSettings.File = tmpUMSFile
			ClientMountUMSImage(StrRemoteHost, StrRemotePort, tmpUsbUMSFile, tmpUsbUMSCdromMode)
		}
	} else {
		fmt.Println("UMS disabled")
		os.Exit(-1)
	}


	return
}

func usbGet(cmd *cobra.Command, args []string) {
	if gs, err := ClientGetDeployedGadgetSettings(StrRemoteHost, StrRemotePort); err == nil {
		if BoolJson {
			PrintGadgetSettings(gs,true)
		} else {
			PrintGadgetSettings(gs,false)
		}
	} else {
		log.Println(err)
	}
}

func usbGetDevicePath(dev devPath) {
	gs, err := ClientGetDeployedGadgetSettings(StrRemoteHost, StrRemotePort)
	if err != nil {
		fmt.Println("%+v\n", err)
		os.Exit(-1)
	}

	res := struct {
		DevPathKeyboard string
		DevPathMouse    string
		DevPathRaw      string
	}{}
	res.DevPathKeyboard = gs.DevPathHidKeyboard
	res.DevPathMouse = gs.DevPathHidMouse
	res.DevPathRaw = gs.DevPathHidRaw

	if dev == dev_path_hid_raw && len(res.DevPathRaw) == 0 {
		fmt.Println("Error: raw HID device disabled")
		os.Exit(-1)
	}
	if dev == dev_path_hid_keyboard && len(res.DevPathKeyboard) == 0 {
		fmt.Println("Error: HID keyboard device disabled")
		os.Exit(-1)
	}
	if dev == dev_path_hid_mouse && len(res.DevPathMouse) == 0 {
		fmt.Println("Error: HID mouse device disabled")
		os.Exit(-1)
	}

	if BoolJson {
		var bytes []byte
		switch dev {
		case dev_path_hid_keyboard:
			bytes, err = json.Marshal(res.DevPathKeyboard)
		case dev_path_hid_mouse:
			bytes, err = json.Marshal(res.DevPathMouse)
		case dev_path_hid_raw:
			bytes, err = json.Marshal(res.DevPathRaw)
		case dev_path_hid_all:
			bytes, err = json.Marshal(res)
		default:
			bytes, err = json.Marshal(res)
		}
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(-1)
		}
		fmt.Println(string(bytes))
	} else {
		switch dev {
		case dev_path_hid_keyboard:
			fmt.Println(res.DevPathKeyboard)
		case dev_path_hid_mouse:
			fmt.Println(res.DevPathMouse)
		case dev_path_hid_raw:
			fmt.Println(res.DevPathRaw)
		case dev_path_hid_all:
			fmt.Printf("%+v\n", res)
		default:
			fmt.Printf("%+v\n", res)
		}

	}

}

func init() {
	cmdUsb := &cobra.Command{
		Use:   "usb",
		Short: "USB gadget settings",
	}

	cmdUsbDeploy := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy given gadget settings",
	}

	cmdUsbSet := &cobra.Command{
		Use:   "set",
		Short: "set USB Gadget settings",
		Long:  ``,
		Run:   usbSet,
	}

	cmdUsbMount := &cobra.Command{
		Use:   "mount",
		Short: "mount CDRom/block device image, if UMS is enabled",
		Long:  ``,
		Run:   usbMount,
	}

	cmdUsbGet := &cobra.Command{
		Use:   "get",
		Short: "Retrieve information on current USB gadget settings",
		Run: usbGet,
	}
	cmdUsbGetDevice := &cobra.Command{
		Use:   "device",
		Short: "Retrieve information on current USB gadget settings",
		Run: func(cmd *cobra.Command, args []string) {
			usbGetDevicePath(dev_path_hid_all)
		},
	}
	cmdUsbGetDeviceKbd := &cobra.Command{
		Use:   "keyboard",
		Short: "Retrieve path of HID keyboard device",
		Run: func(cmd *cobra.Command, args []string) {
			usbGetDevicePath(dev_path_hid_keyboard)
		},
	}
	cmdUsbGetDeviceMouse := &cobra.Command{
		Use:   "mouse",
		Short: "Retrieve path of HID mouse device",
		Run: func(cmd *cobra.Command, args []string) {
			usbGetDevicePath(dev_path_hid_mouse)
		},
	}
	cmdUsbGetDeviceRaw := &cobra.Command{
		Use:   "raw",
		Short: "Retrieve path of HID raw device",
		Run: func(cmd *cobra.Command, args []string) {
			usbGetDevicePath(dev_path_hid_raw)
		},
	}

	rootCmd.AddCommand(cmdUsb)
	cmdUsb.AddCommand(cmdUsbDeploy, cmdUsbSet, cmdUsbGet, cmdUsbMount)
	cmdUsbGet.AddCommand(cmdUsbGetDevice)
	cmdUsbGetDevice.AddCommand(cmdUsbGetDeviceKbd)
	cmdUsbGetDevice.AddCommand(cmdUsbGetDeviceMouse)
	cmdUsbGetDevice.AddCommand(cmdUsbGetDeviceRaw)

	cmdUsbSet.Flags().BoolVarP(&tmpUsbDisableGadget, "disable", "n", false, "If this flag is set, the gadget stays inactive after deployment (not bound to UDC)")
	cmdUsbSet.Flags().BoolVarP(&tmpUsbUseRNDIS, "rndis", "r", false, "Use the RNDIS gadget function")
	cmdUsbSet.Flags().BoolVarP(&tmpUsbUseECM, "cdc-ecm", "e", false, "Use the CDC ECM gadget function")
	cmdUsbSet.Flags().BoolVarP(&tmpUsbUseSerial, "serial", "s", false, "Use the SERIAL gadget function")
	cmdUsbSet.Flags().BoolVarP(&tmpUsbUseHIDKeyboard, "hid-keyboard", "k", false, "Use the HID KEYBOARD gadget function")
	cmdUsbSet.Flags().BoolVarP(&tmpUsbUseHIDMouse, "hid-mouse", "m", false, "Use the HID MOUSE gadget function")
	cmdUsbSet.Flags().BoolVarP(&tmpUsbUseHIDRaw, "hid-raw", "g", false, "Use the HID RAW gadget function")
	cmdUsbSet.Flags().BoolVarP(&tmpUsbUseUMS, "ums", "u", false, "Use the USB Mass Storage gadget function")

	cmdUsbSet.Flags().BoolVar(&tmpUsbUMSCdromMode, "ums-cdrom", false, "If this flag is set, UMS emulates a CD-Rom instead of a flashdrive (ignored, if UMS disabled)")
	cmdUsbSet.Flags().StringVar(&tmpUsbUMSFile, "ums-file", "", "Path to the image or block device backing UMS (ignored, if UMS disabled)")

	cmdUsbSet.Flags().StringVarP(&tmpUsbSerialnumber, "sn", "x", "deadbeef1337", "Serial number (alpha numeric)")
	cmdUsbSet.Flags().StringVarP(&tmpUsbVid, "vid", "v", "0x1d6b", "Vendor ID (format '0x1d6b')")
	cmdUsbSet.Flags().StringVarP(&tmpUsbPid, "pid", "p", "0x1347", "Product ID (format '0x1347')")
	cmdUsbSet.Flags().StringVarP(&tmpUsbManufacturer, "manufacturer", "f", "MaMe82", "Manufacturer string")
	cmdUsbSet.Flags().StringVarP(&tmpUsbProduct, "product", "o", "P4wnP1 by MaMe82", "Product name string")

	cmdUsbMount.Flags().BoolVar(&tmpUsbUMSCdromMode, "ums-cdrom", false, "If this flag is set, UMS emulates a CD-Rom instead of a flashdrive (ignored, if UMS disabled)")
	cmdUsbMount.Flags().StringVar(&tmpUsbUMSFile, "ums-file", "", "Path to the image or block device backing UMS (ignored, if UMS disabled)")

	cmdUsb.PersistentFlags().BoolVar(&BoolJson, "json", false, "Output results as JSON if applicable")

}
