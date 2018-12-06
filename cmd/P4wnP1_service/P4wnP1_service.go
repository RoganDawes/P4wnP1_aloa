// +build linux,arm

package main

import (
	"fmt"
	"github.com/mame82/P4wnP1_aloa/common_web"
	"github.com/mame82/P4wnP1_aloa/service"
	"log"
	"os"
	"os/signal"
	"syscall"
)



func main() {
	//ToDo: Check for root privs
	fmt.Println("P4wnP1 A.L.O.A. " + common_web.VERSION)

	svc,err := service.NewService()
	if err != nil {
		panic(err)
	}
	ctx,_ := svc.Start()

/*
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
	select {
	case s := <-sig:
		log.Printf("Signal (%v) received, ending P4wnP1_service ...\n", s)
	case <- ctx.Done():
		log.Printf("Service cancelled, ending P4wnP1_service ...\n")
	}


	svc.Stop()
	return
}
