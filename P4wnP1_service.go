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


	service.InitLed(false) //Set LED to manual triger
	//core.StartRpcServer("127.0.0.1", "50051") //start gRPC service
	service.StartRpcServer("", "50051") //start gRPC service
}
