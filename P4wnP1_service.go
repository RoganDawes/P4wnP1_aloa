package main

import (
	"log"

	"./service"
	"os"
	"os/signal"
	"syscall"
	"fmt"
	pb "./proto"
	"./common"
	"time"
	"strconv"
)



func main() {




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


	//service.DeployWifiSettings(service.GetDefaultWiFiSettings())

	service.InitLed(false) //Set LED to manual trigger
	//service.InitDefaultLEDSettings()

	/*
	log.Printf("Keyboard devFile: %s\n", service.HidDevPath[service.USB_FUNCTION_HID_KEYBOARD_name])
	log.Printf("Mouse devFile: %s\n", service.HidDevPath[service.USB_FUNCTION_HID_MOUSE_name])
	log.Printf("HID RAW devFile: %s\n", service.HidDevPath[service.USB_FUNCTION_HID_RAW_name])
	*/

	service.StartRpcServerAndWeb("0.0.0.0", "50051", "80", "/home/pi/P4wnP1_go/www") //start gRPC service

	//Indicate servers up with LED blink count 1
	service.SetLed(pb.LEDSettings{1})

	service.StartEventManager(20)
	log.SetOutput(service.EvMgr)
	log.Println("TESTMESSAGE")
	go func() {
		err := common.RunBashScript("/usr/local/P4wnP1/scripts/servicestart.sh")
		if err != nil { log.Printf("Error executing service startup script: %v\n", err) }
	}()


	//Send some log messages for testing
	textfill := "Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea"
	i := 0
	go func() {
		for {
			service.EvMgr.Emit(service.ConstructEventLog("test source", i%5, "message " +strconv.Itoa(i) + ": " + textfill))
			time.Sleep(time.Millisecond *100)
			i++
		}
	}()



	//use a channel to wait for SIGTERM or SIGINT
	fmt.Println("P4wnP1 service initialized, stop with SIGTERM or SIGINT")
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig
	log.Printf("Signal (%v) received, ending P4wnP1_service ...\n", s)
	return
}
