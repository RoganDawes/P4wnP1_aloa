//package rpcserv
package service

import (
	"log"
	pb "../proto"
	"golang.org/x/net/context"
	"net"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"encoding/json"
	"errors"
	"fmt"
)

type server struct {}

func (s *server) DeployEthernetInterfaceSettings(ctx context.Context, es *pb.EthernetInterfaceSettings) (empty *pb.Empty, err error) {
	log.Printf("Trying to deploy ethernet interface settings %v", es)

	empty = &pb.Empty{}
	err = ConfigureInterface(es)
	if err != nil {
		log.Printf("Error deploying ethernet interface settings %v", err)
	}
	return
}

func (s *server) MountUMSFile(ctx context.Context, gsu *pb.GadgetSettingsUMS) (*pb.Empty, error) {
	log.Printf("Trying to mount iamge `%s` to UMS ...", gsu.File)
	err := MountUMSFile(gsu.File)
	return nil, err
}

func (s *server) GetDeployedGadgetSetting(ctx context.Context, e *pb.Empty) (gs *pb.GadgetSettings, err error) {
	gs, err = ParseGadgetState(USB_GADGET_NAME)

	if err == nil {
		j_usbset, _ := json.Marshal(gs)
		log.Printf("Gadget settings requested %v", string(j_usbset))
	} else {
		log.Printf("Error parsing current gadget config: %v", err)
	}

	return
}

func (s *server) DeployGadgetSetting(context.Context, *pb.Empty) (gs *pb.GadgetSettings, err error) {
	gs_backup,_ := ParseGadgetState(USB_GADGET_NAME)

	//ToDo: Former gadgets are destroyed without testing if there're changes, this should be aborted if GadgetSettingsState == GetDeployedGadgetSettings()
	DestroyGadget(USB_GADGET_NAME)

	errg := DeployGadgetSettings(GadgetSettingsState)
	err = nil
	if errg != nil {
		err = errors.New(fmt.Sprintf("Deploying new gadget settings failed, reverted to old ones: %v", errg))
		DeployGadgetSettings(*gs_backup) //We don't catch the error, as the old settings should have been working
	}

	gs, _ = ParseGadgetState(USB_GADGET_NAME) //Return settings from deployed gadget
	return
}

func (s *server) GetGadgetSettings(context.Context, *pb.Empty) (*pb.GadgetSettings, error) {
	return &GadgetSettingsState, nil
}

func (s *server) SetGadgetSettings(ctx context.Context, gs *pb.GadgetSettings) (res *pb.GadgetSettings, err error) {
	if err = ValidateGadgetSetting(*gs); err != nil {
		//We return the validation error and the current (unchanged) GadgetSettingsState
		res = &GadgetSettingsState
		return
	}
	GadgetSettingsState = *gs
	res = &GadgetSettingsState
	return
}

func (s *server) GetLEDSettings(context.Context, *pb.Empty) (res *pb.LEDSettings, err error) {
	res, err = GetLed()
	log.Printf("GetLEDSettings, result: %+v", res)
	return
}

func (s *server) SetLEDSettings(ctx context.Context, ls *pb.LEDSettings) (*pb.Empty, error) {
	log.Printf("SetLEDSettings %+v", ls)
	SetLed(*ls)
	return &pb.Empty{}, nil
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
