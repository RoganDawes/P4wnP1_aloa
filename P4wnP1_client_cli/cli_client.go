package main

import (
	"log"
	"time"
	"reflect"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "../proto"
)

const (
	address     = "localhost:50051"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewP4WNP1Client(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.GetGadgetSettings(ctx, &pb.Empty{})
	if err != nil {
		log.Fatalf("could not get GadgetSettings: %v", err)
	}
	log.Printf("USB Settings %s: %+v", reflect.TypeOf(*r), *r)
}
