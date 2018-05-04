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
	if err != nil {
		return
	}

	ls, err = client.GetLEDSettings(ctx, &pb.Empty{})
	if err != nil {
		log.Printf("Error getting LED blink count: %v", err)
	}

	ClientDisconnectServer(cancel, conn)
	return
}

func ClientGetGadgetSettings(host string, port string) (gs *pb.GadgetSettings, err error) {
	conn, client, ctx, cancel, err := ClientConnectServer(host, port)
	if err != nil {
		return
	}

	gs, err = client.GetGadgetSettings(ctx, &pb.Empty{})
	if err != nil {
		log.Printf("Error getting USB Gadget Settings count: %+v", err)
	}

	ClientDisconnectServer(cancel, conn)
	return
}


func ClientSetGadgetSettings(host string, port string, gs pb.GadgetSettings) (err error) {
	conn, client, ctx, cancel, err := ClientConnectServer(host, port)
	if err != nil {
		return
	}

	_, err = client.SetGadgetSettings(ctx, &gs)
	if err != nil {
		log.Printf("Error setting GadgetSettings %d: %+v", gs, err)
	}

	ClientDisconnectServer(cancel, conn)
	return
}

func ClientSetLED(host string, port string, ls pb.LEDSettings) (err error) {
	conn, client, ctx, cancel, err := ClientConnectServer(host, port)
	if err != nil {
		return
	}

	_, err = client.SetLEDSettings(ctx, &ls)
	if err != nil {
		log.Printf("Error setting LED blink count %d: %v", ls.BlinkCount, err)
	}

	ClientDisconnectServer(cancel, conn)
	return
}

func ClientDisconnectServer(cancel context.CancelFunc, connection *grpc.ClientConn) error {
	defer connection.Close()
	defer cancel()
	return nil
}
