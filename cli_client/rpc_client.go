package cli_client

import (
	"log"
	"google.golang.org/grpc"

	pb "../proto"
	"time"
	"golang.org/x/net/context"
)

func ClientConnectServer(rpcHost string, rpcPort string) (
	connection *grpc.ClientConn,
	client pb.P4WNP1Client,
	ctx context.Context,
	cancel context.CancelFunc,
	err error) {
	// Set up a connection to the server.
	address := rpcHost + ":" + rpcPort
	//log.Printf("Connecting %s ...", address)
	connection, err = grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect to P4wnP1 RPC server: %v", err)
	}
	//defer conn.Close()

	client = pb.NewP4WNP1Client(connection)

	// Contact the server
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	//defer cancel()

	err = nil
	return
}

func ClientGetLED(host string, port string) (ls *pb.LEDSettings, err error) {
	conn, client, ctx, cancel, err := ClientConnectServer(host, port)
	defer conn.Close()
	defer cancel()
	if err != nil { return }

	ls, err = client.GetLEDSettings(ctx, &pb.Empty{})
	if err != nil {
		log.Printf("Error getting LED blink count: %v", err)
	}

	return
}

func ClientGetGadgetSettings(host string, port string) (gs *pb.GadgetSettings, err error) {
	conn, client, ctx, cancel, err := ClientConnectServer(host, port)
	defer conn.Close()
	defer cancel()
	if err != nil { return 	}

	gs, err = client.GetGadgetSettings(ctx, &pb.Empty{})
	if err != nil {
		log.Printf("Error getting USB Gadget Settings: %+v", err)
	}

	return
}

func ClientDeployGadgetSettings(host string, port string) (gs *pb.GadgetSettings, err error) {
	conn, client, ctx, cancel, err := ClientConnectServer(host, port)
	defer conn.Close()
	defer cancel()
	if err != nil { return 	}


	gs, err = client.DeployGadgetSetting(ctx, &pb.Empty{})
	if err != nil {
		log.Printf("Error deploying current USB Gadget Settings: %+v", err)
		//We have an error case, thus gs isn't submitted by the gRPC server (even if the value is provided)
		//in case of an error `gs`should reflect the Gadget Settings which are deployed, now that deployment of the
		//new settings failed. So we fetch the result manually
		gs, _ = client.GetDeployedGadgetSetting(ctx, &pb.Empty{}) //We ignore a new error this time, if it occures `gs` will be nil
	}

	return
}

func ClientGetDeployedGadgetSettings(host string, port string) (gs *pb.GadgetSettings, err error) {
	conn, client, ctx, cancel, err := ClientConnectServer(host, port)
	defer conn.Close()
	defer cancel()
	if err != nil {	return }

	gs, err = client.GetDeployedGadgetSetting(ctx, &pb.Empty{})
	if err != nil {
		log.Printf("Error getting USB Gadget Settings count: %+v", err)
	}

	return
}


func ClientSetGadgetSettings(host string, port string, gs pb.GadgetSettings) (err error) {
	conn, client, ctx, cancel, err := ClientConnectServer(host, port)
	defer conn.Close()
	defer cancel()
	if err != nil { return }

	_, err = client.SetGadgetSettings(ctx, &gs)
	//Only forward the error
	/*
	if err != nil {
		log.Printf("Error setting GadgetSettings %d: %+v", gs, err)
	}
	*/
	return
}

func ClientSetLED(host string, port string, ls pb.LEDSettings) (err error) {
	conn, client, ctx, cancel, err := ClientConnectServer(host, port)
	defer conn.Close()
	defer cancel()
	if err != nil { return }

	_, err = client.SetLEDSettings(ctx, &ls)
	if err != nil {
		log.Printf("Error setting LED blink count %d: %v", ls.BlinkCount, err)
	}


	return
}

func ClientDeployEthernetInterfaceSettings(host string, port string, settings *pb.EthernetInterfaceSettings) (err error) {
	// Set up a connection to the server.
	address := host + ":" + port
	//log.Printf("Connecting %s ...", address)
	connection, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect to P4wnP1 RPC server: %v", err)
	}
	defer connection.Close()

	rpcClient := pb.NewP4WNP1Client(connection)

	// Contact the server
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 30)
	defer cancel()

	_, err = rpcClient.DeployEthernetInterfaceSettings(ctx, settings)
	return

}

func ClientDeployWifiSettings(host string, port string, settings *pb.WiFiSettings) (err error) {
	// Set up a connection to the server.
	address := host + ":" + port
	//log.Printf("Connecting %s ...", address)
	connection, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect to P4wnP1 RPC server: %v", err)
	}
	defer connection.Close()

	rpcClient := pb.NewP4WNP1Client(connection)

	// Contact the server
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 30)
	defer cancel()

	_,err = rpcClient.DeployWifiSettings(ctx, settings)


	return err
}

/*
func ClientDisconnectServer(cancel context.CancelFunc, connection *grpc.ClientConn) error {
	defer connection.Close()
	defer cancel()
	return nil
}
*/