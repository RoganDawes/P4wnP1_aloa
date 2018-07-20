package main

import (
	pb "../proto/gopherjs"
	"context"
	"io"
	"sync"
	"errors"
)

type Rpc struct {
	*sync.Mutex
	Client pb.P4WNP1Client
	eventListeningOn bool
	eventListeningCtx *context.Context
	eventListeningCancel *context.CancelFunc
}

func (rpc *Rpc) StartListenEvents(evtType int64) (err error) {
	rpc.Lock()
	if rpc.eventListeningOn {
		rpc.Unlock()
		return errors.New("Already listening to events")
	}
	// shouldn't happen
	if rpc.eventListeningCancel != nil {
		//Cancel old eventListeners
		cancel := *rpc.eventListeningCancel
		cancel()
	}

	ctx,cancel := context.WithCancel(context.Background())
	rpc.eventListeningCtx = &ctx
	rpc.eventListeningCancel = &cancel
	rpc.eventListeningOn = true
	rpc.Unlock()


	evStream, err := rpc.Client.EventListen(ctx, &pb.EventRequest{ListenType: evtType})
	if err != nil { return err }

	go func() {
		defer rpc.StopEventListening()
		for {
			event, err := evStream.Recv()
			if err == io.EOF { break }
			if err != nil { return }

			println("Event: ", event)
		}
		return
	}()

	return nil
}

func (rpc *Rpc) StopEventListening() {
	rpc.Lock()
	if rpc.eventListeningCancel != nil {
		(*rpc.eventListeningCancel)()
	}
	rpc.eventListeningCancel = nil
	rpc.eventListeningCtx= nil
	rpc.eventListeningOn = false
	rpc.Unlock()
}

func NewRpcClient(addr string) Rpc {
	rcl := Rpc{}
	rcl.Mutex = &sync.Mutex{}
	cl := pb.NewP4WNP1Client(addr)
	rcl.Client = cl
	return rcl
}

/*
func ClientRegisterEvent(host string, port string,  evtType int64) (err error) {
	// open gRPC Client
	address := host + ":" + port
	connection, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {return}
	defer connection.Close()
	client := pb.NewP4WNP1Client(connection)
	evStream, err := client.EventListen(context.Background(), &pb.EventRequest{ListenType: evtType})
	if err != nil { return err }

	for {
		event, err := evStream.Recv()
		if err == io.EOF { break }
		if err != nil { return err }

		log.Printf("Event: %+v", event)
	}
	return nil
}
*/
