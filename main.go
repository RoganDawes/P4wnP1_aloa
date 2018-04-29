package main

import (
	"fmt"
	"./core"
)

func main() {
	usb.DestroyAllGadgets()

	usb_gadget := usb.New()
	
	usb_gadget.Pid = "0x1234"
//	usb_gadget.Use_RNDIS = true //2 EP
//	usb_gadget.Use_CDC_ECM	= true // 2 EP
//	usb_gadget.Use_HID_KEYBOARD = true //1 EP
//	usb_gadget.Use_HID_MOUSE = true // 1 EP
//	usb_gadget.Use_HID_RAW = true //1 EP
	usb_gadget.Use_SERIAL = true //2 EP
	
	fmt.Printf("%+v\n", usb_gadget)
	
	
	err := usb_gadget.CreateGadget()
	if err != nil {
		fmt.Print(err)
	}
	
}
