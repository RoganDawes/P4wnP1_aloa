package main

import (
	"fmt"
	"./core"
)

func main() {
	settings := settings.NewSettings()
	
	
	fmt.Printf("Hello World from P4wnP1\n")
	fmt.Printf("%+v\n", settings)
	fmt.Printf("%+v\n", *settings.Usb)

	settings.Usb.Use_RNDIS = true
	
	err := settings.Usb.CreateGadget()
	if err != nil {
		fmt.Print(err)
	}
}
