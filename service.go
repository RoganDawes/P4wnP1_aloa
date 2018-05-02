package main

import (
	"log"

	"./core"
)

func main() {
	var err error
	err = core.CheckLibComposite()
	if err != nil {
		log.Fatalf("Couldn't load libcomposite: %v", err)
	}

	err = core.DestroyAllGadgets()
	if err != nil {
		log.Fatalf("Error while rolling back existing USB gadgets: %v", err)
	}

	err = core.InitDefaultGadgetSettings()
	if err != nil {
		log.Fatalf("Error while setting up the default gadget: %v", err)
	}

	core.InitLed(false) //Set LED to manual triger
	//core.StartRpcServer("127.0.0.1", "50051") //start gRPC service
	core.StartRpcServer("", "50051") //start gRPC service
}
