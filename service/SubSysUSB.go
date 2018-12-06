// +build linux

package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/mame82/P4wnP1_aloa/common"
	"github.com/mame82/P4wnP1_aloa/hid"
	pb "github.com/mame82/P4wnP1_aloa/proto"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	USB_EP_USAGE_HID_RAW = 1
	USB_EP_USAGE_HID_KEYBOARD = 1
	USB_EP_USAGE_HID_MOUSE = 1
	USB_EP_USAGE_RNDIS = 2
	USB_EP_USAGE_CDC_ECM = 2
	USB_EP_USAGE_CDC_SERIAL = 2
	USB_EP_USAGE_UMS = 2
	USB_EP_USAGE_MAX = 7

	USB_GADGET_NAME = "mame82_gadget"
	USB_GADGET_DIR_BASE      = "/sys/kernel/config/usb_gadget"
	USB_GADGET_DIR           = USB_GADGET_DIR_BASE + "/" + USB_GADGET_NAME

	USB_bcdDevice = "0x0100" //Version 1.00
	USB_bcdUSB    = "0x0200" //mode: USB 2.0

	// composite class / subclass / proto (needs single configuration)
	USB_bDeviceClass    = "0xEF"
	USB_bDeviceSubClass = "0x02"
	USB_bDeviceProtocol = "0x01"

	USB_CONFIGURATION_MaxPower     = "250"
	USB_CONFIGURATION_bmAttributes = "0x80" //should be 0x03 for USB_OTG_SRP | USB_OTG_HNP

	//OS descriptors for RNDIS composite function on Windows
	USB_FUNCTION_RNDIS_os_desc_use                         = "1"
	USB_FUNCTION_RNDIS_os_desc_b_vendor_code               = "0xbc"
	USB_FUNCTION_RNDIS_os_desc_qw_sign                     = "MSFT100"
	USB_FUNCTION_RNDIS_os_desc_interface_compatible_id     = "RNDIS"
	USB_FUNCTION_RNDIS_os_desc_interface_sub_compatible_id = "5162001"

	//HID function, keyboard constants
	USB_FUNCTION_HID_KEYBOARD_protocol      = "1"
	USB_FUNCTION_HID_KEYBOARD_subclass      = "1"
	USB_FUNCTION_HID_KEYBOARD_report_length = "8"
	USB_FUNCTION_HID_KEYBOARD_report_desc   = "\x05\x01\t\x06\xa1\x01\x05\x07\x19\xe0)\xe7\x15\x00%\x01u\x01\x95\x08\x81\x02\x95\x01u\x08\x81\x03\x95\x05u\x01\x05\x08\x19\x01)\x05\x91\x02\x95\x01u\x03\x91\x03\x95\x06u\x08\x15\x00%e\x05\x07\x19\x00)e\x81\x00\xc0"
	USB_FUNCTION_HID_KEYBOARD_name          = "hid.keyboard"

	//HID function, mouse constants
	USB_FUNCTION_HID_MOUSE_protocol      = "2"
	USB_FUNCTION_HID_MOUSE_subclass      = "1"
	USB_FUNCTION_HID_MOUSE_report_length = "6"
	USB_FUNCTION_HID_MOUSE_report_desc   = "\x05\x01\t\x02\xa1\x01\t\x01\xa1\x00\x85\x01\x05\t\x19\x01)\x03\x15\x00%\x01\x95\x03u\x01\x81\x02\x95\x01u\x05\x81\x03\x05\x01\t0\t1\x15\x81%\x7fu\x08\x95\x02\x81\x06\x95\x02u\x08\x81\x01\xc0\xc0\x05\x01\t\x02\xa1\x01\t\x01\xa1\x00\x85\x02\x05\t\x19\x01)\x03\x15\x00%\x01\x95\x03u\x01\x81\x02\x95\x01u\x05\x81\x01\x05\x01\t0\t1\x15\x00&\xff\x7f\x95\x02u\x10\x81\x02\xc0\xc0"
	USB_FUNCTION_HID_MOUSE_name          = "hid.mouse"

	//HID function, custom vendor device constants
	USB_FUNCTION_HID_RAW_protocol      = "1"
	USB_FUNCTION_HID_RAW_subclass      = "1"
	USB_FUNCTION_HID_RAW_report_length = "64"
	USB_FUNCTION_HID_RAW_report_desc   = "\x06\x00\xff\t\x01\xa1\x01\t\x01\x15\x00&\xff\x00u\x08\x95@\x81\x02\t\x02\x15\x00&\xff\x00u\x08\x95@\x91\x02\xc0"
	USB_FUNCTION_HID_RAW_name          = "hid.raw"



)

var (
	ErrUsbNotUsable = errors.New("USB subsystem not available")
	ErrHidNotUsable = errors.New("HIDScript not available (mouse and keyboard disabled)")
	rp_usbHidDevName                      = regexp.MustCompile("(?m)DEVNAME=(.*)\n")
)

type UsbManagerState struct {
	//UndeployedGadgetSettings *pb.GadgetSettings
	DevicePath map[string]string
}

type UsbGadgetManager struct {
	RootSvc *Service
	Usable bool

	State *UsbManagerState
	// ToDo: variable, indicating if HIDScript is usable
	hidCtl             *hid.HIDController // Points to an HID controller instance only if keyboard and/or mouse are enabled, nil otherwise
	gadgetSettingsLock *sync.Mutex
}

func (gm *UsbGadgetManager) HandleEvent(event hid.Event) {
	fmt.Printf("GADGET MANAGER HID EVENT: %+v\n", event)
	gm.RootSvc.SubSysEvent.Emit(ConstructEventHID(event))
}

func (gm *UsbGadgetManager) HidScriptUsable() error {
	// Fix: If a TriggerAction fires a HIDScript during Deployment of GadgetSettings, this method is called
	// upfront, to check if mouse or keyboard are up
	// We avoid returning false, if Deployment is still going on (undefined state), so we block
	gm.gadgetSettingsLock.Lock()
	defer gm.gadgetSettingsLock.Unlock()

	if gm.hidCtl == nil {
		return ErrHidNotUsable
	}
	return nil
}

func (gm *UsbGadgetManager) HidScriptRun(ctx context.Context, scriptContent string) (result interface{}, err error) {
	err = gm.HidScriptUsable()
	if err != nil {	return }

	scriptVal,err := gm.hidCtl.RunScript(ctx, scriptContent, true)
	if err != nil { return nil, err}

	return scriptVal.Export()
}

func (gm *UsbGadgetManager) HidScriptStartBackground(ctx context.Context, scriptContent string) (job *hid.AsyncOttoJob, err error) {
	err = gm.HidScriptUsable()
	if err != nil {	return }

	return gm.hidCtl.StartScriptAsBackgroundJob(ctx, scriptContent, true)
}

//WaitBackgroundJobResult(ctx context.Context, job *AsyncOttoJob) (val otto.Value, err error) {
func (gm *UsbGadgetManager) HidScriptWaitBackgroundJobResult(ctx context.Context, job *hid.AsyncOttoJob) (result interface{}, err error) {
	err = gm.HidScriptUsable()
	if err != nil {	return }

	scriptVal,err := gm.hidCtl.WaitBackgroundJobResult(ctx, job)
	if err != nil { return nil, err}

	return scriptVal.Export()
}

func (gm *UsbGadgetManager) HidScriptGetBackgroundJobByID(id int) (job *hid.AsyncOttoJob, err error) {
	err = gm.HidScriptUsable()
	if err != nil {	return }

	return gm.hidCtl.GetBackgroundJobByID(id)
}

func (gm *UsbGadgetManager) HidScriptGetAllRunningBackgroundJobs() (jobs []*hid.AsyncOttoJob, err error) {
	err = gm.HidScriptUsable()
	if err != nil {	return }

	return gm.hidCtl.GetAllBackgroundJobs()
}

func (gm *UsbGadgetManager) HidScriptCancelAllRunningBackgroundJobs() (err error) {
	err = gm.HidScriptUsable()
	if err != nil {	return }

	gm.hidCtl.CancelAllBackgroundJobs()
	return
}



func NewUSBGadgetManager(rooSvc *Service) (newUGM *UsbGadgetManager, err error) {
	newUGM = &UsbGadgetManager{
		RootSvc:            rooSvc,
		Usable:             true,
		gadgetSettingsLock: &sync.Mutex{},
		State: &UsbManagerState{
			DevicePath: map[string]string{},
		},
	}

	if err = CheckLibComposite(); err != nil {
		//return nil, errors.New(fmt.Sprintf("Couldn't load libcomposite: %v", err))
		newUGM.Usable = false
		return newUGM,nil
	}



	newUGM.State.DevicePath[USB_FUNCTION_HID_KEYBOARD_name] = ""
	newUGM.State.DevicePath[USB_FUNCTION_HID_MOUSE_name] = ""
	newUGM.State.DevicePath[USB_FUNCTION_HID_RAW_name] = ""


	defGS := GetDefaultGadgetSettings()
	//newUGM.State.UndeployedGadgetSettings = &defGS //preload state with default settings
	err = newUGM.DeployGadgetSettings(&defGS)
	if err != nil {
		newUGM.Usable = false
		return newUGM,nil
	}
	return
}



func ValidateGadgetSetting(gs *pb.GadgetSettings) error {
	/* ToDo: validations
	- Done: check host_addr/dev_addr of RNDIS + CDC ECM to be valid MAC addresses via regex
	- check host_addr/dev_addr of RNDIS + CDC ECM for duplicates
	- check EP consumption to be not more than 7 (ECM 2 EP, RNDIS 2 EP, HID Mouse 1 EP, HID Keyboard 1 EP, HID Raw 1 EP, Serial 2 EP ??, UMS 2 EP ?)
	- check serial, product, Manufacturer to not be empty
	- check Pid, Vid with regex (Note: we don't check if Vid+Pid have been used for another composite function setup, yet)
	- Done: If the gadget is enabled, at least one function has to be enabled
	 */

	log.Println("Validating gadget settings ...")

	if gs.Use_RNDIS {
		_, err := net.ParseMAC(gs.RndisSettings.DevAddr)
		if err != nil { return errors.New(fmt.Sprintf("Validation Error RNDIS DeviceAddress: %v", err))}

		_, err = net.ParseMAC(gs.RndisSettings.HostAddr)
		if err != nil { return errors.New(fmt.Sprintf("Validation Error RNDIS HostAddress: %v", err))}
	}

	if gs.Use_CDC_ECM {
		_, err := net.ParseMAC(gs.CdcEcmSettings.DevAddr)
		if err != nil { return errors.New(fmt.Sprintf("Validation Error CDC ECM DeviceAddress: %v", err))}

		_, err = net.ParseMAC(gs.CdcEcmSettings.HostAddr)
		if err != nil { return errors.New(fmt.Sprintf("Validation Error CDC ECM HostAddress: %v", err))}
	}

	//check endpoint consumption
	sumEp := 0
	if gs.Use_RNDIS { sumEp += USB_EP_USAGE_RNDIS }
	if gs.Use_CDC_ECM { sumEp += USB_EP_USAGE_CDC_ECM }
	if gs.Use_UMS { sumEp += USB_EP_USAGE_UMS }
	if gs.Use_HID_MOUSE { sumEp += USB_EP_USAGE_HID_MOUSE }
	if gs.Use_HID_RAW { sumEp += USB_EP_USAGE_HID_RAW }
	if gs.Use_HID_KEYBOARD { sumEp += USB_EP_USAGE_HID_KEYBOARD }
	if gs.Use_SERIAL { sumEp += USB_EP_USAGE_CDC_SERIAL }

	strConsumption := fmt.Sprintf("Gadget Settings consume %v out of %v available USB Endpoints\n", sumEp, USB_EP_USAGE_MAX)
	log.Print(strConsumption)
	if sumEp > USB_EP_USAGE_MAX { return errors.New(strConsumption)}

	//check if composite gadget is enabled without functions
	if gs.Enabled &&
		!gs.Use_CDC_ECM &&
		!gs.Use_RNDIS &&
		!gs.Use_HID_KEYBOARD &&
		!gs.Use_HID_MOUSE &&
		!gs.Use_HID_RAW &&
		!gs.Use_UMS &&
		!gs.Use_SERIAL {
			return errors.New("if the composite gadget isn't disabled, as least one function has to be enabled")
	}

	return nil
}

func addUSBEthernetBridge() {
	//Create the bridge
	CreateBridge(USB_ETHERNET_BRIDGE_NAME)
	setInterfaceMac(USB_ETHERNET_BRIDGE_NAME, USB_ETHERNET_BRIDGE_MAC)
	SetBridgeSTP(USB_ETHERNET_BRIDGE_NAME, false) //aboid loosing time by learning interface states, both usb0 and usb1 have to be set to forward
	SetBridgeForwardDelay(USB_ETHERNET_BRIDGE_NAME, 0)

	//add the interfaces
	if err := AddInterfaceToBridgeIfExistent(USB_ETHERNET_BRIDGE_NAME, "usb0"); err != nil {
		log.Println(err)
	}
	if err := AddInterfaceToBridgeIfExistent(USB_ETHERNET_BRIDGE_NAME, "usb1"); err != nil {
		log.Println(err)
	}

	//enable the bridge
	NetworkLinkUp(USB_ETHERNET_BRIDGE_NAME)
}

func deleteUSBEthernetBridge() {
	//we ignore error results
	DeleteBridge(USB_ETHERNET_BRIDGE_NAME)
}

/*
Polls for presence of "usb0" / "usb1" till one of both is active or timeout is reached
 */

func pollForUSBEthernet(timeout time.Duration) error {
	for startTime := time.Now(); time.Since(startTime) < timeout; {
		if present := CheckInterfaceExistence("usb0"); present {
			return nil
		}
		if present := CheckInterfaceExistence("usb1"); present {
			return nil
		}

		//Take a breath
		time.Sleep(100*time.Millisecond)
		fmt.Print(".")
	}
	return errors.New(fmt.Sprintf("timeout %v reached before usb0 or usb1 became ready", timeout))
}




//depends on `lsmod` binary
func CheckLibComposite() error {
	log.Printf("Checking for libcomposite...")
	out, err := exec.Command("lsmod").Output()
	if err != nil {
		log.Fatal(err)
	}

	if strings.Contains(string(out), "libcomposite") {
		log.Printf("... libcomposite loaded")
		return nil
	}

	//if here, libcomposite isn't loaded ... try to load
	log.Printf("Libcomposite not loaded, trying to fix ...")
	err = exec.Command("modprobe", "libcomposite").Run()
	if err == nil {
		log.Printf("... libcomposite loaded")
	}

	return err
}

func getUDCName() (string, error) {
	files, err := ioutil.ReadDir("/sys/class/udc")
	if err != nil {
		return "", errors.New("couldn't find working UDC driver")
	}
	if len(files) < 1 {
		return "", errors.New("couldn't find working UDC driver")
	}
	return files[0].Name(), nil

}

func (gm *UsbGadgetManager) ParseGadgetState(gadgetName string) (result *pb.GadgetSettings, err error) {
	err = nil
	result = &pb.GadgetSettings{}

	fmt.Println("ParseGadgetState before lock ...")
	//Don't parse while deploying
	gm.gadgetSettingsLock.Lock()
	defer gm.gadgetSettingsLock.Unlock()

	fmt.Println("ParseGadgetState beyond lock ...")

	gadgetDir := USB_GADGET_DIR_BASE + "/" + gadgetName

	//check if root exists, return error otherwise
	if _, err = os.Stat(gadgetDir); os.IsNotExist(err) {
		err = errors.New(fmt.Sprintf("gadget %s doesn't exist", gadgetName))
		result = nil
		return
	}

	//ToDo: check if enabled (UDC in functionfs is set to content of /sys/class/udc)

	if res, err := ioutil.ReadFile(gadgetDir + "/idVendor"); err != nil {
		err1 := errors.New(fmt.Sprintf("gadget %s error reading Vid", gadgetName))
		return nil, err1
	} else {
		result.Vid = strings.TrimSuffix(string(res), "\n")
	}

	if res, err := ioutil.ReadFile(gadgetDir + "/idProduct"); err != nil {
		err1 := errors.New(fmt.Sprintf("gadget %s error reading Pid", gadgetName))
		return nil, err1
	} else {
		result.Pid = strings.TrimSuffix(string(res), "\n")
	}

	if res, err := ioutil.ReadFile(gadgetDir + "/strings/0x409/serialnumber"); err != nil {
		err1 := errors.New(fmt.Sprintf("gadget %s error reading Serial", gadgetName))
		return nil, err1
	} else {
		result.Serial = strings.TrimSuffix(string(res), "\n")
	}

	if res, err := ioutil.ReadFile(gadgetDir + "/strings/0x409/manufacturer"); err != nil {
		err1 := errors.New(fmt.Sprintf("gadget %s error reading Manufacturer", gadgetName))
		return nil, err1
	} else {
		result.Manufacturer = strings.TrimSuffix(string(res), "\n")
	}

	if res, err := ioutil.ReadFile(gadgetDir + "/strings/0x409/product"); err != nil {
		err1 := errors.New(fmt.Sprintf("gadget %s error reading Product", gadgetName))
		return nil, err1
	} else {
		result.Product = strings.TrimSuffix(string(res), "\n")
	}

	//Check enabled functions in configuration

	//USB RNDIS
	if _, err1 := os.Stat(gadgetDir +"/configs/c.1/rndis.usb0"); !os.IsNotExist(err1) {
		result.Use_RNDIS = true

		result.RndisSettings = &pb.GadgetSettingsEthernet{}

		if res, err := ioutil.ReadFile(gadgetDir + "/functions/rndis.usb0/host_addr"); err != nil {
			err1 := errors.New(fmt.Sprintf("gadget %s error reading RNDIS host_addr", gadgetName))
			return nil, err1
		} else {
			result.RndisSettings.HostAddr = strings.TrimSuffix(string(res), "\000\n")
		}

		if res, err := ioutil.ReadFile(gadgetDir + "/functions/rndis.usb0/dev_addr"); err != nil {
			err1 := errors.New(fmt.Sprintf("gadget %s error reading RNDIS dev_addr", gadgetName))
			return nil, err1
		} else {
			result.RndisSettings.DevAddr = strings.TrimSuffix(string(res), "\000\n")
		}
	} else {
		// we provide GadgetSettingsEthernet with default MAC adresses anyway, to have defaults in case RNDIS should be enabled
		result.RndisSettings = &pb.GadgetSettingsEthernet{
			HostAddr: DEFAULT_RNDIS_HOST_ADDR,
			DevAddr: DEFAULT_RNDIS_DEV_ADDR,
		}
	}

	//USB CDC ECM
	if _, err1 := os.Stat(gadgetDir +"/configs/c.1/ecm.usb1"); !os.IsNotExist(err1) {
		result.Use_CDC_ECM = true

		result.CdcEcmSettings = &pb.GadgetSettingsEthernet{}

		if res, err := ioutil.ReadFile(gadgetDir + "/functions/ecm.usb1/host_addr"); err != nil {
			err1 := errors.New(fmt.Sprintf("gadget %s error reading CDC ECM host_addr", gadgetName))
			return nil, err1
		} else {
			result.CdcEcmSettings.HostAddr = strings.TrimSuffix(string(res), "\000\n")
		}

		if res, err := ioutil.ReadFile(gadgetDir + "/functions/ecm.usb1/dev_addr"); err != nil {
			err1 := errors.New(fmt.Sprintf("gadget %s error reading CDC ECM dev_addr", gadgetName))
			return nil, err1
		} else {
			result.CdcEcmSettings.DevAddr = strings.TrimSuffix(string(res), "\000\n")
		}

	} else {
		// we provide GadgetSettingsEthernet with default MAC adresses anyway, to have defaults in case CDC ECM should be enabled
		result.CdcEcmSettings = &pb.GadgetSettingsEthernet{
			HostAddr: DEFAULT_CDC_ECM_HOST_ADDR,
			DevAddr: DEFAULT_CDC_ECM_DEV_ADDR,
		}
	}

	//USB serial
	if _, err1 := os.Stat(gadgetDir +"/configs/c.1/acm.GS0"); !os.IsNotExist(err1) {
		result.Use_SERIAL = true
	}

	//USB HID Keyboard
	if _, err1 := os.Stat(gadgetDir +"/configs/c.1/"+USB_FUNCTION_HID_KEYBOARD_name); !os.IsNotExist(err1) {
		result.Use_HID_KEYBOARD = true
	}

	//USB HID Mouse
	if _, err1 := os.Stat(gadgetDir +"/configs/c.1/"+USB_FUNCTION_HID_MOUSE_name); !os.IsNotExist(err1) {
		result.Use_HID_MOUSE = true
	}

	//USB HID Raw
	if _, err1 := os.Stat(gadgetDir +"/configs/c.1/"+USB_FUNCTION_HID_RAW_name); !os.IsNotExist(err1) {
		result.Use_HID_RAW = true
	}

	//USB Mass Storage
	if _, err1 := os.Stat(gadgetDir +"/configs/c.1/mass_storage.ms1"); !os.IsNotExist(err1) {
		result.Use_UMS = true
		result.UmsSettings = &pb.GadgetSettingsUMS{}

		//Check if running as CD-Rom
		if res, err := ioutil.ReadFile(gadgetDir + "/functions/mass_storage.ms1/lun.0/cdrom"); err != nil {
			err1 := errors.New(fmt.Sprintf("gadget %s error reading USB Mass Storage cdrom emulation state", gadgetName))
			return nil, err1
		} else {
			if strings.HasPrefix(string(res), "1") {
				result.UmsSettings.Cdrom = true
			} //else branche unneeded, as false is default
		}

		//Check name of backing file
		if res, err := ioutil.ReadFile(gadgetDir + "/functions/mass_storage.ms1/lun.0/file"); err != nil {
			err1 := errors.New(fmt.Sprintf("gadget %s error reading USB Mass Storage image file setting", gadgetName))
			return nil, err1
		} else {
			result.UmsSettings.File = strings.TrimSuffix(string(res), "\000\n")

			// remove path
			result.UmsSettings.File = filepath.Base(result.UmsSettings.File)
		}
	} else {
		result.UmsSettings = &pb.GadgetSettingsUMS{}
	}

	//check if UDC is set (Gadget enabled)
	udcName, _ := getUDCName()

	if res, err := ioutil.ReadFile(gadgetDir + "/UDC"); err != nil {
		err1 := errors.New(fmt.Sprintf("gadget %s error reading UDC", gadgetName))
		return nil, err1
	} else {
		udcNameSet := strings.TrimSuffix(string(res), "\n")
		//log.Printf("UDC test: udcNameSet %s, udcName %s", udcNameSet, udcName)
		if udcName == udcNameSet {
			result.Enabled = true
		}
	}

	return

}

// This command is working on the active gadget directly, so changes aren't reflected back
// to the GadgetSettingsState
func MountUMSFile(filename string) error {
	funcdir := USB_GADGET_DIR + "/functions/mass_storage.ms1"
	err := ioutil.WriteFile(funcdir+"/lun.0/file", []byte(filename), os.ModePerm)
	if err != nil {
		return errors.New(fmt.Sprintf("settings backing file for USB Mass Storage failed: %v", err))
	}
	return nil
}

func (gm *UsbGadgetManager) DeployGadgetSettings(settings *pb.GadgetSettings) (err error) {
	if !gm.Usable {
		return ErrUsbNotUsable
	}

	//fmt.Println("DeployGadgetSettings before lock ...")

	//Lock, only one change at a time
	gm.gadgetSettingsLock.Lock()
	defer gm.gadgetSettingsLock.Unlock()
	//fmt.Println("DeployGadgetSettings beyond lock ...")

	err = ValidateGadgetSetting(settings)
	if err != nil {
		return
	}

	//Crash fix: abort HID controller before destorying gadget, to avoid crashes by attempting to close filedescriptors of non existent files
	if gm.hidCtl != nil {
		gm.hidCtl.Abort()
	}

	fmt.Println("... deconstruct old gadget")
	//ToDo: Former gadgets are destroyed without testing if there're changes, this should be aborted if GadgetSettingsState == GetDeployedGadgetSettings()
	gm.DestroyGadget(USB_GADGET_NAME)

	var usesUSBEthernet bool

	gadgetRoot := USB_GADGET_DIR_BASE

	//check if root exists, return error otherwise
	if _, err := os.Stat(gadgetRoot); os.IsNotExist(err) {
		return errors.New("configfs path for gadget doesn't exist")
	}

	//ToDo: check if UDC is present and usable

	//create gadget folder
	os.Mkdir(USB_GADGET_DIR, os.ModePerm)
	log.Printf("Creating composite gadget '%s'\nSettings:\n%+v", USB_GADGET_NAME, settings)

	//set vendor ID, product ID
	ioutil.WriteFile(USB_GADGET_DIR+"/idVendor", []byte(settings.Vid), os.ModePerm)
	ioutil.WriteFile(USB_GADGET_DIR+"/idProduct", []byte(settings.Pid), os.ModePerm)

	//set USB mode to 2.0 and device version to 1.0
	ioutil.WriteFile(USB_GADGET_DIR+"/bcdUSB", []byte(USB_bcdUSB), os.ModePerm)
	ioutil.WriteFile(USB_GADGET_DIR+"/bcdDevice", []byte(USB_bcdDevice), os.ModePerm)

	//composite class / subclass / proto (needs single configuration)
	ioutil.WriteFile(USB_GADGET_DIR+"/bDeviceClass", []byte(USB_bDeviceClass), os.ModePerm)
	ioutil.WriteFile(USB_GADGET_DIR+"/bDeviceSubClass", []byte(USB_bDeviceSubClass), os.ModePerm)
	ioutil.WriteFile(USB_GADGET_DIR+"/bDeviceProtocol", []byte(USB_bDeviceProtocol), os.ModePerm)

	// set device descriptions
	os.Mkdir(USB_GADGET_DIR+"/strings/0x409", os.ModePerm) // English language strings
	ioutil.WriteFile(USB_GADGET_DIR+"/strings/0x409/serialnumber", []byte(settings.Serial), os.ModePerm)
	ioutil.WriteFile(USB_GADGET_DIR+"/strings/0x409/manufacturer", []byte(settings.Manufacturer), os.ModePerm)
	ioutil.WriteFile(USB_GADGET_DIR+"/strings/0x409/product", []byte(settings.Product), os.ModePerm)

	// create configuration instance (only one, as multiple configs aren't valid for Windows composite devices)
	os.MkdirAll(USB_GADGET_DIR+"/configs/c.1/strings/0x409", os.ModePerm) // English language strings
	ioutil.WriteFile(USB_GADGET_DIR+"/configs/c.1/strings/0x409/configuration", []byte("Config 1: Composite"), os.ModePerm)
	ioutil.WriteFile(USB_GADGET_DIR+"/configs/c.1/MaxPower", []byte(USB_CONFIGURATION_MaxPower), os.ModePerm)
	ioutil.WriteFile(USB_GADGET_DIR+"/configs/c.1/bmAttributes", []byte(USB_CONFIGURATION_bmAttributes), os.ModePerm)

	// RNDIS has to be the first interface on Composite device for Windows (first function initialized)
	if settings.Use_RNDIS {
		log.Printf("... creating USB RNDIS function")
		usesUSBEthernet = true
		os.Mkdir(USB_GADGET_DIR+"/functions/rndis.usb0", os.ModePerm) //create RNDIS function
		ioutil.WriteFile(USB_GADGET_DIR+"/functions/rndis.usb0/host_addr", []byte(settings.RndisSettings.HostAddr), os.ModePerm)
		ioutil.WriteFile(USB_GADGET_DIR+"/functions/rndis.usb0/dev_addr", []byte(settings.RndisSettings.DevAddr), os.ModePerm)

		/*
			add OS specific device descriptors to force Windows to load RNDIS drivers
			=============================================================================
			Witout this additional descriptors, most Windows system detect the RNDIS interface as "Serial COM port"
			To prevent this, the Microsoft specific OS descriptors are added in here
			!! Important:
			If the device already has been connected to the Windows System without providing the
			OS descriptor, Windows never asks again for them and thus never installs the RNDIS driver
			This behavior is driven by creation of an registry hive, the first time a device without
			OS descriptors is attached. The key is build like this:

			HKLM\SYSTEM\CurrentControlSet\Control\usbflags\[USB_VID+USB_PID+bcdRelease\osvc

			To allow Windows to read the OS descriptors again, the according registry hive has to be
			deleted manually or USB descriptor values have to be cahnged (f.e. USB_PID).
		*/

		//set OS descriptors for Windows
		ioutil.WriteFile(USB_GADGET_DIR+"/os_desc/use", []byte(USB_FUNCTION_RNDIS_os_desc_use), os.ModePerm)
		ioutil.WriteFile(USB_GADGET_DIR+"/os_desc/b_vendor_code", []byte(USB_FUNCTION_RNDIS_os_desc_b_vendor_code), os.ModePerm)
		ioutil.WriteFile(USB_GADGET_DIR+"/os_desc/qw_sign", []byte(USB_FUNCTION_RNDIS_os_desc_qw_sign), os.ModePerm)
		ioutil.WriteFile(USB_GADGET_DIR+"/functions/rndis.usb0/os_desc/interface.rndis/compatible_id", []byte(USB_FUNCTION_RNDIS_os_desc_interface_compatible_id), os.ModePerm)
		ioutil.WriteFile(USB_GADGET_DIR+"/functions/rndis.usb0/os_desc/interface.rndis/sub_compatible_id", []byte(USB_FUNCTION_RNDIS_os_desc_interface_sub_compatible_id), os.ModePerm)

		//activate function by symlinking to config 1
		err := os.Symlink(USB_GADGET_DIR+"/functions/rndis.usb0", USB_GADGET_DIR+"/configs/c.1/rndis.usb0")
		if err != nil {
			log.Println(err)
		}

		// add config 1 to OS descriptors
		err = os.Symlink(USB_GADGET_DIR+"/configs/c.1", USB_GADGET_DIR+"/os_desc/c.1")
		if err != nil {
			log.Println(err)
		}
	}

	if settings.Use_CDC_ECM {
		log.Printf("... creating USB CDC ECM function")
		usesUSBEthernet = true
		os.Mkdir(USB_GADGET_DIR+"/functions/ecm.usb1", os.ModePerm) //create CDC ECM function
		ioutil.WriteFile(USB_GADGET_DIR+"/functions/ecm.usb1/host_addr", []byte(settings.CdcEcmSettings.HostAddr), os.ModePerm)
		ioutil.WriteFile(USB_GADGET_DIR+"/functions/ecm.usb1/dev_addr", []byte(settings.CdcEcmSettings.DevAddr), os.ModePerm)

		//activate function by symlinking to config 1
		err := os.Symlink(USB_GADGET_DIR+"/functions/ecm.usb1", USB_GADGET_DIR+"/configs/c.1/ecm.usb1")
		if err != nil {
			log.Println(err)
		}
	}

	if settings.Use_SERIAL {
		log.Printf("... creating USB serial function")
		os.Mkdir(USB_GADGET_DIR+"/functions/acm.GS0", os.ModePerm) //create ACM function

		//activate function by symlinking to config 1
		err := os.Symlink(USB_GADGET_DIR+"/functions/acm.GS0", USB_GADGET_DIR+"/configs/c.1/acm.GS0")
		if err != nil {
			log.Println(err)
		}

	}

	if settings.Use_HID_KEYBOARD {
		log.Printf("... creating USB HID Keyboard function")
		funcdir := USB_GADGET_DIR + "/functions/" + USB_FUNCTION_HID_KEYBOARD_name
		os.Mkdir(funcdir, os.ModePerm) //create HID function for keyboard

		ioutil.WriteFile(funcdir+"/protocol", []byte(USB_FUNCTION_HID_KEYBOARD_protocol), os.ModePerm)
		ioutil.WriteFile(funcdir+"/subclass", []byte(USB_FUNCTION_HID_KEYBOARD_subclass), os.ModePerm)
		ioutil.WriteFile(funcdir+"/report_length", []byte(USB_FUNCTION_HID_KEYBOARD_report_length), os.ModePerm)
		ioutil.WriteFile(funcdir+"/report_desc", []byte(USB_FUNCTION_HID_KEYBOARD_report_desc), os.ModePerm)

		err := os.Symlink(funcdir, USB_GADGET_DIR+"/configs/c.1/"+USB_FUNCTION_HID_KEYBOARD_name)
		if err != nil {
			log.Println(err)
		}
	}

	if settings.Use_HID_MOUSE {
		log.Printf("... creating USB HID Mouse function")
		funcdir := USB_GADGET_DIR + "/functions/" + USB_FUNCTION_HID_MOUSE_name
		os.Mkdir(funcdir, os.ModePerm) //create HID function for mouse

		ioutil.WriteFile(funcdir+"/protocol", []byte(USB_FUNCTION_HID_MOUSE_protocol), os.ModePerm)
		ioutil.WriteFile(funcdir+"/subclass", []byte(USB_FUNCTION_HID_MOUSE_subclass), os.ModePerm)
		ioutil.WriteFile(funcdir+"/report_length", []byte(USB_FUNCTION_HID_MOUSE_report_length), os.ModePerm)
		ioutil.WriteFile(funcdir+"/report_desc", []byte(USB_FUNCTION_HID_MOUSE_report_desc), os.ModePerm)

		err := os.Symlink(funcdir, USB_GADGET_DIR+"/configs/c.1/"+USB_FUNCTION_HID_MOUSE_name)
		if err != nil {
			log.Println(err)
		}
	}

	if settings.Use_HID_RAW {
		log.Printf("... creating USB HID Generic device function")
		funcdir := USB_GADGET_DIR + "/functions/" + USB_FUNCTION_HID_RAW_name
		os.Mkdir(funcdir, os.ModePerm) //create HID function for mouse

		ioutil.WriteFile(funcdir+"/protocol", []byte(USB_FUNCTION_HID_RAW_protocol), os.ModePerm)
		ioutil.WriteFile(funcdir+"/subclass", []byte(USB_FUNCTION_HID_RAW_subclass), os.ModePerm)
		ioutil.WriteFile(funcdir+"/report_length", []byte(USB_FUNCTION_HID_RAW_report_length), os.ModePerm)
		ioutil.WriteFile(funcdir+"/report_desc", []byte(USB_FUNCTION_HID_RAW_report_desc), os.ModePerm)

		err := os.Symlink(funcdir, USB_GADGET_DIR+"/configs/c.1/"+USB_FUNCTION_HID_RAW_name)
		if err != nil {
			log.Println(err)
		}
	}

	if settings.Use_UMS {
		log.Printf("... creating USB Mass Storage device function")
		funcdir := USB_GADGET_DIR + "/functions/mass_storage.ms1"
		os.Mkdir(funcdir, os.ModePerm) //create HID function for mouse

		ioutil.WriteFile(funcdir+"/stall", []byte("1"), os.ModePerm) // Allow bulk Endpoints
		if settings.UmsSettings.Cdrom {
			ioutil.WriteFile(funcdir+"/lun.0/cdrom", []byte("1"), os.ModePerm) // CD-Rom
		} else {
			ioutil.WriteFile(funcdir+"/lun.0/cdrom", []byte("0"), os.ModePerm) // Writable flashdrive
		}

		ioutil.WriteFile(funcdir+"/lun.0/ro", []byte("0"), os.ModePerm) // Don't restrict to read-only (is implied by cdrom=1 if needed, but causes issues on backend FS if enabled)

		// enable Force Unit Access (FUA) to make Windows write synchronously
		// this is slow, but unplugging the stick without unmounting works
		ioutil.WriteFile(funcdir+"/lun.0/nofua", []byte("0"), os.ModePerm) // Don't restrict to read-only (is implied by cdrom=1 if needed, but causes issues on backend FS if enabled)

		//Provide the backing image
		file := settings.UmsSettings.File
		if settings.UmsSettings.Cdrom {
			file = common.PATH_IMAGE_CDROM + "/" + file
		} else {
			file = common.PATH_IMAGE_FLASHDRIVE + "/" + file
		}
		ioutil.WriteFile(funcdir+"/lun.0/file", []byte(file), os.ModePerm) // Set backing file (or block device) for USB Mass Storage

		err := os.Symlink(funcdir, USB_GADGET_DIR+"/configs/c.1/"+"mass_storage.ms1")
		if err != nil {
			log.Println(err)
		}
	}

	//clear device path for HID devices
	gm.State.DevicePath[USB_FUNCTION_HID_KEYBOARD_name] = ""
	gm.State.DevicePath[USB_FUNCTION_HID_MOUSE_name] = ""
	gm.State.DevicePath[USB_FUNCTION_HID_RAW_name] = ""

	//get UDC driver name and bind to gadget
	if settings.Enabled {
		udcName, err := getUDCName()
		if err != nil {
			return err
		}
		log.Printf("Enabeling gadget for UDC: %s\n", udcName)
		if err = ioutil.WriteFile(USB_GADGET_DIR+"/UDC", []byte(udcName), os.ModePerm); err != nil {
			fmt.Println("... error enabling gadget:", err)
			return err
		}

		//update device path
		log.Println("Retrieving path to HID devices")
		if devPath,errF := enumDevicePath(USB_FUNCTION_HID_KEYBOARD_name); errF == nil  { gm.State.DevicePath[USB_FUNCTION_HID_KEYBOARD_name] = devPath }
		if devPath,errF := enumDevicePath(USB_FUNCTION_HID_MOUSE_name); errF == nil  { gm.State.DevicePath[USB_FUNCTION_HID_MOUSE_name] = devPath }
		if devPath,errF := enumDevicePath(USB_FUNCTION_HID_RAW_name); errF == nil  { gm.State.DevicePath[USB_FUNCTION_HID_RAW_name] = devPath }

		//if Keyboard or Mouse are deployed, grab a HIDController Instance else set it to nil (the old HIDController object won't be destroyed)
		if settings.Use_HID_KEYBOARD || settings.Use_HID_MOUSE {
			devPathKeyboard := gm.State.DevicePath[USB_FUNCTION_HID_KEYBOARD_name]
			devPathMouse := gm.State.DevicePath[USB_FUNCTION_HID_MOUSE_name]

			//log.Printf("Starting HID controller (kbd %s, mouse %s)...\n", devPathKeyboard, devPathMouse)
			var errH error
			gm.hidCtl, errH = hid.NewHIDController(context.Background(), devPathKeyboard, common.PATH_KEYBOARD_LANGUAGE_MAPS, devPathMouse)
			gm.hidCtl.SetEventHandler(gm)
			if errH != nil {
				log.Printf("ERROR: Couldn't bring up an instance of HIDController for keyboard: '%s', mouse: '%s' and mapping path '%s'\nReason: %v\n", devPathKeyboard, devPathMouse, common.PATH_KEYBOARD_LANGUAGE_MAPS, errH)
			} else {
				log.Printf("HIDController for keyboard: '%s', mouse: '%s' and mapping path '%s' initialized\n", devPathKeyboard, devPathMouse, common.PATH_KEYBOARD_LANGUAGE_MAPS)
			}
		} else {
			if gm.hidCtl != nil { gm.hidCtl.Abort() }
			gm.hidCtl = nil
			log.Printf("HIDController for keyboard / mouse disabled\n")
		}
	}




	deleteUSBEthernetBridge() //delete former used bridge, if there's any
	//In case USB ethernet is uesd (RNDIS or CDC ECM), we add a bridge interface
	if usesUSBEthernet && settings.Enabled {
		//wait till "usb0" or "usb1" comes up
		err := pollForUSBEthernet(10*time.Second)
		if err == nil {
			//add USBEthernet bridge including the usb interfaces
			log.Printf("... creating network bridge for USB ethernet devices")
			addUSBEthernetBridge()
			log.Printf("... checking for stored network interface settings for USB ethernet")
			//ReInitNetworkInterface(USB_ETHERNET_BRIDGE_NAME)
			if nim,err := gm.RootSvc.SubSysNetwork.GetManagedInterface(USB_ETHERNET_BRIDGE_NAME); err == nil {
				nim.ReDeploy()
			}

		} else {
			return err
		}

	}

	log.Printf("... done")
	return nil
}

func enumDevicePath(funcName string) (devPath string, err error){
	//cat /sys/dev/char/$(cat /sys/kernel/config/usb_gadget/mame82_gadget/functions/hid.mouse/dev)/uevent | grep DEVNAME
	devfile := USB_GADGET_DIR + "/functions/" + funcName + "/dev"

	var udevNode string
	if res, err := ioutil.ReadFile(devfile); err != nil {
		err1 := errors.New(fmt.Sprintf("Gadget error reading udevname for %s\n", funcName))
		return "", err1
	} else {
		udevNode = strings.TrimSuffix(string(res), "\n")
	}


	ueventPath := fmt.Sprintf("/sys/dev/char/%s/uevent", udevNode)
	if ueventContent, err := ioutil.ReadFile(ueventPath); err != nil {
		err1 := errors.New(fmt.Sprintf("Gadget error reading uevent file '%s' for %s\n", ueventPath, funcName))
		return "", err1
	} else {

		strDevNameSub := rp_usbHidDevName.FindStringSubmatch(string(ueventContent))
		if len(strDevNameSub) > 1 { devPath = "/dev/" + strDevNameSub[1]}
	}

	return
}

func (gm *UsbGadgetManager) DestroyAllGadgets() error {

	//gadgetRoot := "./test"
	gadgetRoot := USB_GADGET_DIR_BASE

	//check if root exists, return error otherwise
	if _, err := os.Stat(gadgetRoot); os.IsNotExist(err) {
		return errors.New("configfs path for gadget doesn't exist")
	}

	gadgetDirs, err := ioutil.ReadDir(gadgetRoot)
	if err != nil {
		return errors.New("no gadgets")
	}

	for _, gadgetDirObj := range gadgetDirs {
		gadgetName := gadgetDirObj.Name()
		log.Println("Found gadget: " + gadgetName)

		err = gm.DestroyGadget(gadgetName)
		if err != nil {
			log.Println(err) //don't return, continue with next
		}
	}

	if gm.hidCtl != nil { gm.hidCtl.Abort() }
	gm.hidCtl = nil
	log.Printf("HIDController for keyboard / mouse disabled\n")

	return nil
}

func (gm *UsbGadgetManager) DestroyGadget(gadgetName string) error {



	gadgetDir := USB_GADGET_DIR_BASE + "/" + gadgetName

	//check if root exists, return error otherwise
	if _, err := os.Stat(USB_GADGET_DIR_BASE); os.IsNotExist(err) {
		return errors.New("Gadget " + gadgetName + " doesn't exist")
	}
	log.Println("Deconstructing gadget " + gadgetName + "...")

	//Assure gadget gets unbound from UDC
	ioutil.WriteFile(gadgetDir+"/UDC", []byte("\x00"), os.ModePerm)

	//Iterate over configurations
	configDirs, _ := ioutil.ReadDir(gadgetDir + "/configs")
	for _, confDirObj := range configDirs {
		confName := confDirObj.Name()
		confDir := gadgetDir + "/configs/" + confName
		log.Println("Found config: " + confName)

		//find linked functions
		confContents, _ := ioutil.ReadDir(confDir)
		for _, function := range confContents {
			//Remove link from function to config
			if function.Mode()&os.ModeSymlink > 0 {
				log.Println("\tRemoving function " + function.Name() + " from config " + confName)
				os.Remove(confDir + "/" + function.Name())
			}
		}

		//find string directories in config
		stringsContents, _ := ioutil.ReadDir(confDir + "/strings")
		for _, str := range stringsContents {
			stringDir := str.Name()
			//Remove string from config
			log.Println("\tRemoving string dir '" + stringDir + "' from configuration")
			os.Remove(confDir + "/strings/" + stringDir)
		}

		//Check if there's an OS descriptor refering this config
		if _, err := os.Stat(gadgetDir + "/os_desc/" + confName); !os.IsNotExist(err) {
			log.Println("\tDeleting link to '" + confName + "' from gadgets OS descriptor")
			os.Remove(gadgetDir + "/os_desc/" + confName)
		}

		// remove config folder, finally
		log.Println("\tDeleting configuration '" + confName + "'")
		os.Remove(confDir)
	}

	// remove functions
	log.Println("Removing functions from '" + gadgetName + "'")
	os.RemoveAll(gadgetDir + "/functions/")

	//find string directories in gadget
	stringsContents, _ := ioutil.ReadDir(gadgetDir + "/strings")
	for _, str := range stringsContents {
		stringDir := str.Name()
		//Remove string from config
		log.Println("Removing string dir '" + stringDir + "' from " + gadgetName)
		os.Remove(gadgetDir + "/strings/" + stringDir)
	}

	//And now remove the gadget itself
	log.Println("Removing gadget " + gadgetName)
	os.Remove(gadgetDir)

	return nil
}
