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
func (s *server) GetGadgetSettings(context.Context, *pb.Empty) (*pb.GadgetSettings, error) {
	usbset := &pb.GadgetSettings{
		Pid:              "0x1337",
		Vid:              "0x1222",
		Manufacturer:     "MaMe82",
		Serial:           "deadbeef13371337",
		Product:          "P4wnP1 by MaMe82",
		Use_RNDIS:        false,
		Use_CDC_ECM:      true,
		Use_HID_KEYBOARD: false,
		Use_HID_MOUSE:    false,
		Use_HID_RAW:      false,
		Use_UMS:          false,
		Use_SERIAL:       false,

		RndisSettings:  &pb.GadgetSettingsEthernet{HostAddr: "11:22:33:44:55:66", DevAddr: "66:55:44:33:22:11"},
		CdcEcmSettings: &pb.GadgetSettingsEthernet{HostAddr: "11:22:33:54:76:98", DevAddr: "66:55:44:98:76:54"},
	}
	j_usbset, _ := json.Marshal(usbset)
	log.Printf("Gadget settings requested %v", string(j_usbset))

	return usbset, nil
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
	log.Printf("P4wnP1 RPC server lsitening on " + listen_address)
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
