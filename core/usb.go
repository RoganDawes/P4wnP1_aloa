package core

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	pb "../proto"
)

const (
	USB_GADGET_DIR_BASE      = "/sys/kernel/config/usb_gadget"
	USB_GADGET_DIR           = USB_GADGET_DIR_BASE + "/mame82_gadget"
	USB_DEFAULT_SERIAL       = "deadbeefdeadbeef"
	USB_DEFAULT_MANUFACTURER = "MaMe82"
	USB_DEFAULT_PRODUCT      = "P4wnP1 by MaMe82"

	USB_bcdDevice = "0x0100" //Version 1.00
	USB_bcdUSB    = "0x0200" //mode: USB 2.0

	// composite class / subclass / proto (needs single configuration)
	USB_bDeviceClass    = "0xEF"
	USB_bDeviceSubClass = "0x02"
	USB_bDeviceProtocol = "0x01"

	USB_CONFIGURATION_MaxPower     = "250"
	USB_CONFIGURATION_bmAttributes = "0x80" //should be 0x03 for USB_OTG_SRP | USB_OTG_HNP

	/*
		//RNDIS function constants
		USB_FUNCTION_RNDIS_DEFAULT_host_addr = "42:63:65:12:34:56"
		USB_FUNCTION_RNDIS_DEFAULT_dev_addr  = "42:63:65:56:34:12"
	*/
	//OS descriptors for RNDIS composite function on Windows
	USB_FUNCTION_RNDIS_os_desc_use                         = "1"
	USB_FUNCTION_RNDIS_os_desc_b_vendor_code               = "0xbc"
	USB_FUNCTION_RNDIS_os_desc_qw_sign                     = "MSFT100"
	USB_FUNCTION_RNDIS_os_desc_interface_compatible_id     = "RNDIS"
	USB_FUNCTION_RNDIS_os_desc_interface_sub_compatible_id = "5162001"

	/*
		//CDC ECM function constants
		USB_FUNCTION_CDC_ECM_DEFAULT_host_addr = "42:63:66:12:34:56"
		USB_FUNCTION_CDC_ECM_DEFAULT_dev_addr  = "42:63:66:56:34:12"
	*/

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

func InitDefaultGadgetSettings() error {
	return InitGadget(CreateDefaultGadgetSettings())
}

func CreateDefaultGadgetSettings() (res pb.GadgetSettings) {
	res = pb.GadgetSettings{
		Vid:              "0x1d6b",
		Pid:              "0x1337",
		Manufacturer:     "MaMe82",
		Product:          "P4wnP1 by MaMe82",
		Serial:           "deadbeef1337",
		Use_CDC_ECM:      true,
		Use_RNDIS:        true,
		Use_HID_KEYBOARD: false,
		Use_HID_MOUSE:    false,
		Use_HID_RAW:      false,
		Use_UMS:          false,
		Use_SERIAL:       false,
		RndisSettings: &pb.GadgetSettingsEthernet{
			HostAddr: "42:63:65:12:34:56",
			DevAddr:  "42:63:65:56:34:12",
		},
		CdcEcmSettings: &pb.GadgetSettingsEthernet{
			HostAddr: "42:63:66:12:34:56",
			DevAddr:  "42:63:66:56:34:12",
		},
	}

	return res
}

//depends on `bash`, `grep` and `lsmod` binary
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

func InitGadget(settings pb.GadgetSettings) error {
	//gadget_root := "./test"
	gadget_root := USB_GADGET_DIR_BASE

	//check if root exists, return error otherwise
	if _, err := os.Stat(gadget_root); os.IsNotExist(err) {
		return errors.New("Configfs path for gadget doesn't exist")
	}

	//ToDo: check if UDC is present and usable

	//create gadget folder
	os.Mkdir(USB_GADGET_DIR, os.ModePerm)
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
		os.Mkdir(USB_GADGET_DIR+"/functions/acm.GS0", os.ModePerm) //create ACM function

		//activate function by symlinking to config 1
		err := os.Symlink(USB_GADGET_DIR+"/functions/acm.GS0", USB_GADGET_DIR+"/configs/c.1/acm.GS0")
		if err != nil {
			log.Println(err)
		}

	}

	if settings.Use_HID_KEYBOARD {
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

	//get UDC driver name and bind to gadget
	files, err := ioutil.ReadDir("/sys/class/udc")
	if err != nil {
		return errors.New("Couldn't find working UDC driver")
	}
	if len(files) < 1 {
		return errors.New("Couldn't find working UDC driver")
	}
	udc_name := files[0].Name()
	ioutil.WriteFile(USB_GADGET_DIR+"/UDC", []byte(udc_name), os.ModePerm)

	return nil
}

func DestroyAllGadgets() error {
	//gadget_root := "./test"
	gadget_root := USB_GADGET_DIR_BASE

	//check if root exists, return error otherwise
	if _, err := os.Stat(gadget_root); os.IsNotExist(err) {
		return errors.New("Configfs path for gadget doesn't exist")
	}

	gadget_dirs, err := ioutil.ReadDir(gadget_root)
	if err != nil {
		return errors.New("No gadgets")
	}

	for _, gadget_dir_obj := range gadget_dirs {
		gadget_name := gadget_dir_obj.Name()
		log.Println("Found gadget: " + gadget_name)
		err = DestroyGadget(gadget_name)
		if err != nil {
			log.Println(err) //don't return, continue with next
		}
	}

	return nil
}

func DestroyGadget(Gadget_name string) error {
	//gadget_root := "./test"
	gadget_dir := USB_GADGET_DIR_BASE + "/" + Gadget_name

	//check if root exists, return error otherwise
	if _, err := os.Stat(USB_GADGET_DIR_BASE); os.IsNotExist(err) {
		return errors.New("Gadget " + Gadget_name + " doesn't exist")
	}
	log.Println("Deconstructing gadget " + Gadget_name + "...")

	//Assure gadget gets unbound from UDC
	ioutil.WriteFile(gadget_dir+"/UDC", []byte("\x00"), os.ModePerm)

	//Iterate over configurations
	config_dirs, _ := ioutil.ReadDir(gadget_dir + "/configs")
	for _, conf_dir_obj := range config_dirs {
		conf_name := conf_dir_obj.Name()
		conf_dir := gadget_dir + "/configs/" + conf_name
		log.Println("Found config: " + conf_name)

		//find linked functions
		conf_content, _ := ioutil.ReadDir(conf_dir)
		for _, function := range conf_content {
			//Remove link from function to config
			if function.Mode()&os.ModeSymlink > 0 {
				log.Println("\tRemoving function " + function.Name() + " from config " + conf_name)
				os.Remove(conf_dir + "/" + function.Name())
			}
		}

		//find string directories in config
		strings_content, _ := ioutil.ReadDir(conf_dir + "/strings")
		for _, str := range strings_content {
			string_dir := str.Name()
			//Remove string from config
			log.Println("\tRemoving string dir '" + string_dir + "' from configuration")
			os.Remove(conf_dir + "/strings/" + string_dir)
		}

		//Check if there's an OS descriptor refering this config
		if _, err := os.Stat(gadget_dir + "/os_desc/" + conf_name); !os.IsNotExist(err) {
			log.Println("\tDeleting link to '" + conf_name + "' from gadgets OS descriptor")
			os.Remove(gadget_dir + "/os_desc/" + conf_name)
		}

		// remove config folder, finally
		log.Println("\tDeleting configuration '" + conf_name + "'")
		os.Remove(conf_dir)
	}

	// remove functions
	log.Println("Removing functions from '" + Gadget_name + "'")
	os.RemoveAll(gadget_dir + "/functions/")

	//find string directories in gadget
	strings_content, _ := ioutil.ReadDir(gadget_dir + "/strings")
	for _, str := range strings_content {
		string_dir := str.Name()
		//Remove string from config
		log.Println("Removing string dir '" + string_dir + "' from " + Gadget_name)
		os.Remove(gadget_dir + "/strings/" + string_dir)
	}

	//And now remove the gadget itself
	log.Println("Removing gadget " + Gadget_name)
	os.Remove(gadget_dir)

	return nil
}
