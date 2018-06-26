package main

import (
	"log"

	"./service"
)

func main() {
	/*
	if gadget, err := core.ParseGadgetState(core.USB_GADGET_NAME); err == nil {
		log.Printf("Gadget config: %+v", gadget)
	} else {
		log.Printf("Gadget %s couldn't be parsed: %s", core.USB_GADGET_NAME, err)
	}
	*/

	//ToDo: Check for root privs

	var err error
	err = service.CheckLibComposite()
	if err != nil {
		log.Fatalf("Couldn't load libcomposite: %v", err)
	}

	err = service.DestroyAllGadgets()
	if err != nil {
		log.Fatalf("Error while rolling back existing USB gadgets: %v", err)
	}

	err = service.InitDefaultGadgetSettings()
	if err != nil {
		log.Fatalf("Error while setting up the default gadget: %v", err)
	}


	service.DeployWifiSettings(service.GetDefaultWiFiSettings())


	service.InitLed(false) //Set LED to manual trigger
	service.InitDefaultLEDSettings()

	log.Printf("Keyboard devFile: %s\n", service.HidDevPath[service.USB_FUNCTION_HID_KEYBOARD_name])
	log.Printf("Mouse devFile: %s\n", service.HidDevPath[service.USB_FUNCTION_HID_MOUSE_name])
	log.Printf("HID RAW devFile: %s\n", service.HidDevPath[service.USB_FUNCTION_HID_RAW_name])


	service.StartRpcServerAndWeb("0.0.0.0", "50051") //start gRPC service
}
