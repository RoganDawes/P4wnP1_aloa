//package rpcserv
package core

import (
	"encoding/json"
	"log"
	"net"

	pb "../proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct{}

//Attach handler function implementing the "GetGadgetSettings" interface to server
func (s *server) GetGadgetSettings(context.Context, *pb.Empty) (usbset *pb.GadgetSettings, err error) {
	usbset, err = ParseGadgetState(USB_GADGET_NAME)

	if err == nil {
		j_usbset, _ := json.Marshal(usbset)
		log.Printf("Gadget settings requested %v", string(j_usbset))
	} else {
		log.Printf("Error parsing current gadget config: %v", err)
	}


	return usbset, err
}

//Attach handler function implementing the "SetGadgetSettings" interface to server
func (s *server) SetGadgetSettings(context.Context, *pb.GadgetSettings) (*pb.Error, error) {
	return &pb.Error{Err: 0}, nil
}

//Attach handler function implementing the "StartGadget" interface to server
func (s *server) StartGadget(context.Context, *pb.Empty) (*pb.Error, error) {
	return &pb.Error{Err: 0}, nil
}

//Attach handler function implementing the "StopGadget" interface to server
func (s *server) StopGadget(context.Context, *pb.Empty) (*pb.Error, error) {
	return &pb.Error{Err: 0}, nil
}

func (s *server) GetLEDSettings(context.Context, *pb.Empty) (*pb.LEDSettings, error) {
	led_settings := &pb.LEDSettings{}
	return led_settings, nil
}

func (s *server) SetLEDSettings(ctx context.Context, ledSettings *pb.LEDSettings) (rpcerr *pb.Error, err error) {
	log.Printf("SetLEDSettings %+v", ledSettings)
	setLed(*ledSettings)
	return &pb.Error{Err: 0}, nil
}

func StartRpcServer(host string, port string) {
	listen_address := host + ":" + port
	//Open TCP listener
	log.Printf("P4wnP1 RPC server listening on " + listen_address)
	lis, err := net.Listen("tcp", listen_address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	//Create gRPC Server
	s := grpc.NewServer()
	pb.RegisterP4WNP1Server(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
