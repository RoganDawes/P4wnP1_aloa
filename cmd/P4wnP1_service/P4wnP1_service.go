// +build linux,arm

package main

import (
	"fmt"
	"github.com/mame82/P4wnP1_go/service"
	"log"
	"os"
	"os/signal"
	"syscall"
)



func main() {

	//ToDo: Check for root privs


	/*
	log.Printf("Keyboard devFile: %s\n", service.HidDevPath[service.USB_FUNCTION_HID_KEYBOARD_name])
	log.Printf("Mouse devFile: %s\n", service.HidDevPath[service.USB_FUNCTION_HID_MOUSE_name])
	log.Printf("HID RAW devFile: %s\n", service.HidDevPath[service.USB_FUNCTION_HID_RAW_name])
	*/

	// ToDo: The webroot has to be changed to /usr/local/P4wnP1/www



	svc,err := service.NewService()
	if err != nil {
		panic(err)
	}
	svc.Start()

/*
	// ToDo: Remove this (testing only)
	//Send some log messages for testing
	textfill := "Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea"
	i := 0
	go func() {
		for {
			//println("Sending log event")
			svc.SubSysEvent.Emit(service.ConstructEventLog("test source", i%5, "message " +strconv.Itoa(i) + ": " + textfill))
			time.Sleep(time.Millisecond *3000)
			i++
		}
	}()
*/

	//use a channel to wait for SIGTERM or SIGINT
	fmt.Println("P4wnP1 service initialized, stop with SIGTERM or SIGINT")
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig
	log.Printf("Signal (%v) received, ending P4wnP1_service ...\n", s)
	svc.Stop()
	return
}
