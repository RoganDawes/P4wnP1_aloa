package settings

import(
	"errors"
	"os"
	"fmt"
	"io/ioutil"
)

const (
	USB_GADGET_DIR_BASE = "/sys/kernel/config/usb_gadget"
	USB_GADGET_DIR = USB_GADGET_DIR_BASE + "/mame82_gadget"
	USB_DEFAULT_SERIAL = "deadbeefdeadbeef"
	USB_DEFAULT_MANUFACTURER = "MaMe82"
	USB_DEFAULT_PRODUCT = "P4wnP1 by MaMe82"
	
	USB_bcdDevice = "0x0100" //Version 1.00
	USB_bcdUSB = "0x0200" //mode: USB 2.0

	// composite class / subclass / proto (needs single configuration)
	USB_bDeviceClass = "0xEF"
	USB_bDeviceSubClass = "0x02"
	USB_bDeviceProtocol = "0x01"
	
	USB_CONFIGURATION_MaxPower = "250"
	USB_CONFIGURATION_bmAttributes = "0x80" //should be 0x03 for USB_OTG_SRP | USB_OTG_HNP

	//RNDIS function constants
	USB_FUNCTION_RNDIS_DEFAULT_host_addr = "42:63:65:12:34:56"
	USB_FUNCTION_RNDIS_DEFAULT_dev_addr = "42:63:65:56:34:12"
	//OS descriptors for RNDIS composite function on Windows
	USB_FUNCTION_RNDIS_os_desc_use = "1"
	USB_FUNCTION_RNDIS_os_desc_b_vendor_code = "0xbc"
	USB_FUNCTION_RNDIS_os_desc_qw_sign = "MSFT100"
	USB_FUNCTION_RNDIS_os_desc_interface_compatible_id = "RNDIS"
	USB_FUNCTION_RNDIS_os_desc_interface_sub_compatible_id = "5162001"
	
	//CDC ECM function constants
	USB_FUNCTION_CDC_ECM_DEFAULT_host_addr = "42:63:66:12:34:56"
	USB_FUNCTION_CDC_ECM_DEFAULT_dev_addr = "42:63:66:56:34:12"
	
	//HID function, keyboard constants
	USB_FUNCTION_HID_KEYBOARD_protocol = "1"
	USB_FUNCTION_HID_KEYBOARD_subclass = "1"
	USB_FUNCTION_HID_KEYBOARD_report_length = "8"
	USB_FUNCTION_HID_KEYBOARD_name = "hid.keyboard"
	
	//HID function, mouse constants
	USB_FUNCTION_HID_MOUSE_protocol = "2"
	USB_FUNCTION_HID_MOUSE_subclass = "1"
	USB_FUNCTION_HID_MOUSE_report_length = "6"
	USB_FUNCTION_HID_MOUSE_name = "hid.mouse"
	
	//HID function, custom vendor device constants
	USB_FUNCTION_HID_RAW_protocol = "1"
	USB_FUNCTION_HID_RAW_subclass = "1"
	USB_FUNCTION_HID_RAW_report_length = "64"
	USB_FUNCTION_HID_RAW_name = "hid.raw"
)

type USB struct {
	Vid string
	Pid string
	Manufacturer string
	Product string
	Serial string
	Use_CDC_ECM bool
	Use_RNDIS bool
	Use_HID_KEYBOARD bool
	Use_HID_MOUSE bool
	Use_HID_RAW bool
	Use_UMS bool
	USE_SERIAL bool
}

func (settings USB) CreateGadget() error {
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
		ioutil.WriteFile(USB_GADGET_DIR+"/functions/rndis.usb0/host_addr", []byte(USB_FUNCTION_RNDIS_DEFAULT_host_addr), os.ModePerm)
		ioutil.WriteFile(USB_GADGET_DIR+"/functions/rndis.usb0/dev_addr", []byte(USB_FUNCTION_RNDIS_DEFAULT_dev_addr), os.ModePerm)

		//set OS descriptors for Windows
		ioutil.WriteFile(USB_GADGET_DIR+"/os_desc/use", []byte(USB_FUNCTION_RNDIS_os_desc_use), os.ModePerm)
		ioutil.WriteFile(USB_GADGET_DIR+"/os_desc/b_vendor_code", []byte(USB_FUNCTION_RNDIS_os_desc_b_vendor_code), os.ModePerm)
		ioutil.WriteFile(USB_GADGET_DIR+"/os_desc/qw_sign", []byte(USB_FUNCTION_RNDIS_os_desc_qw_sign), os.ModePerm)
		
		ioutil.WriteFile(USB_GADGET_DIR+"/functions/rndis.usb0/os_desc/interface.rndis/compatible_id", []byte(USB_FUNCTION_RNDIS_os_desc_interface_compatible_id), os.ModePerm)
		ioutil.WriteFile(USB_GADGET_DIR+"/functions/rndis.usb0/os_desc/interface.rndis/sub_compatible_id", []byte(USB_FUNCTION_RNDIS_os_desc_interface_sub_compatible_id), os.ModePerm)

		err := os.Symlink(USB_GADGET_DIR+"/functions/rndis.usb0", USB_GADGET_DIR+"/configs/c.1/rndis.usb0")
		if err != nil {
			fmt.Println(err)
		}
		
		// add config 1 to OS descriptors
		err = os.Symlink(USB_GADGET_DIR+"/configs/c.1", USB_GADGET_DIR+"/os_desc/c.1")
		if err != nil {
			fmt.Println(err)
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

func NewUSB() *USB {
	ust := &USB{
		Vid: "0x1d6b",
		Pid: "0x1338",
		Manufacturer: USB_DEFAULT_MANUFACTURER,
		Product: USB_DEFAULT_PRODUCT,
		Serial: USB_DEFAULT_SERIAL,
	}

	return ust
}

func NewSettings() Settings {
	st := Settings {
		Usb: NewUSB(),
	}
	
	return st
}

type Settings struct {
	Usb *USB
}

