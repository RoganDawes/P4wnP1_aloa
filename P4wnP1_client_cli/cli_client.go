package main

import (
	"log"
	"time"
//	"reflect"
	"flag" //will be replaced with cobra 

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "../proto"
	"reflect"
)

func main() {
	//Parse cli flags, should be replaced with cobra
	getUsbGadgetConf := flag.Bool("get_gadget_state", false, "Retrieves the current USB gadget state")
	blinkCountPtr := flag.Int("blink", -1, "LED blink count (0 = LED off, 255 = LED solid, 1..254 blink n times)")
	var rpcHostPtr string
	var rpcPortPtr string
	flag.StringVar(&rpcHostPtr, "host", "localhost", "The remote RPC host running P4wnP1 service")
	flag.StringVar(&rpcPortPtr, "port", "50051", "The remote RPC port of P4wnP1 service")
	flag.Parse()
	
	
	// Set up a connection to the server.
	address := rpcHostPtr + ":" + rpcPortPtr
	log.Printf("Connecting %s ...", address)
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewP4WNP1Client(conn)

	// Contact the server
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if *blinkCountPtr >= 0 {
		_, err1 := c.SetLEDSettings(ctx, &pb.LEDSettings{BlinkCount: uint32(*blinkCountPtr)})
		if err1 != nil {
			log.Printf("Error setting LED blink count %d: %v", *blinkCountPtr, err1)
		}
	}

	if *getUsbGadgetConf {
		r, err := c.GetGadgetSettings(ctx, &pb.Empty{})
		if err != nil {
			log.Fatalf("could not get GadgetSettings: %v", err)
		}
		log.Printf("USB Settings %s: %+v", reflect.TypeOf(*r), *r)
	}
	
	/*
	r, err := c.GetGadgetSettings(ctx, &pb.Empty{})
	if err != nil {
		log.Fatalf("could not get GadgetSettings: %v", err)
	}
	log.Printf("USB Settings %s: %+v", reflect.TypeOf(*r), *r)
	
	log.Printf("Set LED to blink count 3")
	c.SetLEDSettings(ctx, &pb.LEDSettings{ BlinkCount: 3})
	*/
}
