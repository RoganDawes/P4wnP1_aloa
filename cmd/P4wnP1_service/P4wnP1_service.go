// +build linux,arm

package main

import (
	"log"

	"github.com/mame82/P4wnP1_go/service"
	"os"
	"os/signal"
	"syscall"
	"fmt"
	pb "github.com/mame82/P4wnP1_go/proto"
	"github.com/mame82/P4wnP1_go/common"
	"time"
	"strconv"
)



func main() {

	err := service.InitGlobalServiceState()
	if err != nil { panic(err) }

	state := service.ServiceState
	state.StartService()


	//ToDo: Check for root privs


	//service.NewLed(false) //Set LED to manual trigger
	//service.InitDefaultLEDSettings()

	/*
	log.Printf("Keyboard devFile: %s\n", service.HidDevPath[service.USB_FUNCTION_HID_KEYBOARD_name])
	log.Printf("Mouse devFile: %s\n", service.HidDevPath[service.USB_FUNCTION_HID_MOUSE_name])
	log.Printf("HID RAW devFile: %s\n", service.HidDevPath[service.USB_FUNCTION_HID_RAW_name])
	*/

	// ToDo: The webroot has to be changed to /usr/local/P4wnP1/www
	service.StartRpcServerAndWeb("0.0.0.0", "50051", "8000", "/usr/local/P4wnP1/www") //start gRPC service

	//Indicate servers up with LED blink count 1
	state.Led.SetLed(&pb.LEDSettings{1})

	//service.StartEventManager(20)
//	log.SetOutput(state.EvMgr)
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
			//println("Sending log event")
			state.EvMgr.Emit(service.ConstructEventLog("test source", i%5, "message " +strconv.Itoa(i) + ": " + textfill))
			time.Sleep(time.Millisecond *2000)
			i++
		}
	}()



	//use a channel to wait for SIGTERM or SIGINT
	fmt.Println("P4wnP1 service initialized, stop with SIGTERM or SIGINT")
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig
	log.Printf("Signal (%v) received, ending P4wnP1_service ...\n", s)
	state.StopService()
	return
}
