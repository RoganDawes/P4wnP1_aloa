// +build js

package main

import (
	pb "github.com/mame82/P4wnP1_go/proto/gopherjs"
	"context"
	"sync"
	"github.com/johanbrandhorst/protobuf/grpcweb"
	"time"
	"errors"
)

type Rpc struct {
	*sync.Mutex
	Client pb.P4WNP1Client
	eventListeningOn bool
	eventListeningCtx *context.Context
	eventListeningCancel *context.CancelFunc
}

/*
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
*/

func NewRpcClient(addr string) Rpc {
	rcl := Rpc{}
	rcl.Mutex = &sync.Mutex{}
	cl := pb.NewP4WNP1Client(addr, grpcweb.WithDefaultCallOptions(grpcweb.ForceWebsocketTransport()))
	rcl.Client = cl
	return rcl
}

func RpcGetRunningHidJobStates(timeout time.Duration) (states []*pb.HIDRunningJobStateResult, err error) {
	println("RpcGetRunningHidJobStates called")


	ctx,cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// get running job IDs
	joblist, err := Client.Client.HIDGetRunningScriptJobs(ctx,&pb.Empty{})
	if err != nil { return nil, err }

	states = make([]*pb.HIDRunningJobStateResult, len(joblist.Ids))
	for idx,jobid := range joblist.Ids {
		jobstate, err := Client.Client.HIDGetRunningJobState(ctx, &pb.HIDScriptJob{Id:jobid})
		if err != nil { return nil, err }
		states[idx] = jobstate
	}

	return states,nil
}

func RpcGetDeployedGadgetSettings(timeout time.Duration) (*pb.GadgetSettings, error) {
	//gs := vue.GetVM(c).Get("gadgetSettings")
	println("RpcGetDeployedGadgetSettings called")

	ctx,cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()


	return Client.Client.GetDeployedGadgetSetting(ctx, &pb.Empty{})

}


func RpcSetRemoteGadgetSettings(targetGS *pb.GadgetSettings, timeout time.Duration) (err error) {
	//gs := vue.GetVM(c).Get("gadgetSettings")
	println("RpcSetRemoteGadgetSettings called")

	ctx,cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()


	//Set gadget settings
	_, err = Client.Client.SetGadgetSettings(ctx, targetGS)
	if err != nil {
		//js.Global.Call("alert", "Error setting given gadget settings: " + status.Convert(err).Message())
		//println(err)
		//c.UpdateFromDeployedGadgetSettings(vm)
		return err
	}

	return nil
}

func RpcDeployRemoteGadgetSettings(timeout time.Duration) (*pb.GadgetSettings, error) {
	//gs := vue.GetVM(c).Get("gadgetSettings")
	println("RpcDeployRemoteGadgetSettings called")

	ctx,cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()


	return Client.Client.DeployGadgetSetting(ctx, &pb.Empty{})

}


func ConnectionTest(timeout time.Duration) (err error) {
	//gs := vue.GetVM(c).Get("gadgetSettings")
	println("RpcDeployRemoteGadgetSettings called")

	ctx,cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req := &pb.StringMessage{Msg:"ping"}
	resp,err := Client.Client.EchoRequest(ctx, req)
	if err != nil { return err }
	if resp.Msg != req.Msg { errors.New("Unexpected response")}
	return nil
}
