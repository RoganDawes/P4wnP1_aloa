// +build js

package main

import (
	pb "github.com/mame82/P4wnP1_go/proto/gopherjs"
	"context"
	"sync"
	"github.com/johanbrandhorst/protobuf/grpcweb"
	"time"
	"errors"
	"github.com/mame82/P4wnP1_go/common_web"
	"io"
)

type Rpc struct {
	*sync.Mutex
	Client               pb.P4WNP1Client
	eventListeningOn     bool
	eventListeningCtx    *context.Context
	eventListeningCancel context.CancelFunc
}

func NewRpcClient(addr string) Rpc {
	println("Bringing up RPC client for address:", addr)
	rcl := Rpc{}
	rcl.Mutex = &sync.Mutex{}
	cl := pb.NewP4WNP1Client(addr, grpcweb.WithDefaultCallOptions(grpcweb.ForceWebsocketTransport()))
	rcl.Client = cl
	return rcl
}

func (rpc *Rpc) DeployedEthernetInterfaceSettings(timeout time.Duration, settings *pb.EthernetInterfaceSettings) (err error) {
	// ToDo: The RPC call has to return an error in case deployment fails
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err = rpc.Client.DeployEthernetInterfaceSettings(ctx, settings)
	return
}


func (rpc *Rpc) DeployWifiSettings(timeout time.Duration, settings *pb.WiFiSettings) (state *pb.WiFiState, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	state, err = rpc.Client.DeployWiFiSettings(ctx, settings)
	return
}


func (rpc *Rpc) GetDeployedWiFiSettings(timeout time.Duration) (settingsList *jsWiFiSettings, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ws,err := rpc.Client.GetWiFiState(ctx, &pb.Empty{})
	if err != nil {
		println("Error GetDeployedWifiSettings", err)
		return nil, err
	}

	println("GetDeployedWifiSettings: ", ws)

	// Update state

	jsWs := &jsWiFiSettings{Object: O()}
	jsWs.fromGo(ws.CurrentSettings)
	return jsWs, nil

}


func (rpc *Rpc) GetAllDeployedEthernetInterfaceSettings(timeout time.Duration) (settingsList *jsEthernetSettingsList, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	deployedSettings, err := rpc.Client.GetAllDeployedEthernetInterfaceSettings(ctx, &pb.Empty{})
	if err != nil {
		println("Error GetAllDeployedEthernetInterfaceSettings", err)
		return nil, err
	}

	settingsList = &jsEthernetSettingsList{Object: O()}
	settingsList.fromGo(deployedSettings)

	return settingsList, nil

	/*
	js.Global.Set("ds", deployedSettings.List)

	for idx,is := range deployedSettings.List {
		jis := &jsEthernetInterfaceSettings{Object:O()}
		jis.fromGo(is)
		name := "ds"+strconv.Itoa(idx)
		println("Globalizing " + name)
		js.Global.Set(name, jis)
	}
	*/
}

func (rpc *Rpc) RpcGetRunningHidJobStates(timeout time.Duration) (states []*pb.HIDRunningJobStateResult, err error) {
	println("RpcGetRunningHidJobStates called")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// get running job IDs
	joblist, err := rpc.Client.HIDGetRunningScriptJobs(ctx, &pb.Empty{})
	if err != nil {
		return nil, err
	}

	states = make([]*pb.HIDRunningJobStateResult, len(joblist.Ids))
	for idx, jobid := range joblist.Ids {
		jobstate, err := rpc.Client.HIDGetRunningJobState(ctx, &pb.HIDScriptJob{Id: jobid})
		if err != nil {
			return nil, err
		}
		states[idx] = jobstate
	}

	return states, nil
}

func (rpc *Rpc) RpcGetDeployedGadgetSettings(timeout time.Duration) (*pb.GadgetSettings, error) {
	//gs := vue.GetVM(c).Get("gadgetSettings")
	println("RpcGetDeployedGadgetSettings called")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return rpc.Client.GetDeployedGadgetSetting(ctx, &pb.Empty{})

}

func (rpc *Rpc) RpcSetRemoteGadgetSettings(targetGS *pb.GadgetSettings, timeout time.Duration) (err error) {
	//gs := vue.GetVM(c).Get("gadgetSettings")
	println("RpcSetRemoteGadgetSettings called")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	//Set gadget settings
	_, err = rpc.Client.SetGadgetSettings(ctx, targetGS)
	if err != nil {
		//js.Global.Call("alert", "Error setting given gadget settings: " + status.Convert(err).Message())
		//println(err)
		//c.UpdateFromDeployedGadgetSettings(vm)
		return err
	}

	return nil
}

func (rpc *Rpc) RpcDeployRemoteGadgetSettings(timeout time.Duration) (*pb.GadgetSettings, error) {
	//gs := vue.GetVM(c).Get("gadgetSettings")
	println("RpcDeployRemoteGadgetSettings called")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return rpc.Client.DeployGadgetSetting(ctx, &pb.Empty{})

}

func (rpc *Rpc) ConnectionTest(timeout time.Duration) (err error) {
	//gs := vue.GetVM(c).Get("gadgetSettings")
	println("RpcDeployRemoteGadgetSettings called")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req := &pb.StringMessage{Msg: "ping"}
	resp, err := rpc.Client.EchoRequest(ctx, req)
	if err != nil {
		return err
	}
	if resp.Msg != req.Msg {
		errors.New("Unexpected response")
	}
	return nil
}

func (rpc *Rpc) StartListening() {

	println("Start listening called", globalState.EventReceiver)

	//Note: This method is responsible for handling server streaming of events
	// It isn't possible to use the stream for connection watching (heartbeat), for the following reasons
	// 1) A connection loss can be detected in case `evStream.Recv()` fails with an error, but a successful websocket
	// connection can't be detected with this method, as it blocks till a message is received (in case the connection
	// succeeds). Thus `evStream.Recv()` could be used to indicate connection error, but not to indicate successful
	// connections.
	// 2) The initial call to `Client.EventListen` seems to be another place to distinguish between successful and
	// failed Websocket connection establishment. Unfortunately this method doesn't return an error for a failed
	// Websocket connection attempt, even if the target host isn't reachable at all.
	// --> Solution: A unary call is used to check if the server is reachable

	go func() {
		for {
			println("Try to connect server ...")
			for RpcClient.ConnectionTest(time.Second*3) != nil {
				println("... failed, retry for 3 seconds")
				globalState.FailedConnectionAttempts++
			}
			println("... success")
			globalState.IsConnected = true
			globalState.FailedConnectionAttempts = 0

			ctx, cancel := context.WithCancel(context.Background())
			rpc.eventListeningCancel = cancel

			// try RPC call
			evStream, err := rpc.Client.EventListen(ctx, &pb.EventRequest{ListenType: common_web.EVT_ANY}) //No error if Websocket connection fails
			if err == nil {
				println("EVENTLISTENING ENTERING LOOP")
			Inner:
				for {
					//Note:
					event, err := evStream.Recv() //Error if Websocket connection fails/aborts, but success is indicated only if stream data is received
					if err == io.EOF {
						break Inner
					}
					if err != nil {
						println("EVENTLISTENING ERROR", err)
						break Inner
					}

					//println("Event: ", event)
					globalState.EventReceiver.HandleEvent(event)
				}
				// we end here on connection error
				cancel()
				println("EVENTLISTENING ABORTED")

			} else {
				globalState.IsConnected = false
				// Note: This error case isn't reached when the websocket based RPC call can't establish a connection,
				// instead the error occurs when the evStream.Recv() method is called
				cancel()
				println("Error listening for Log events", err)
			}
			println("Connection to server lost, reconnecting ...")
			globalState.IsConnected = false

			//retry to connect (outer loop)
		}

		return
	}()
}

func (rpc *Rpc) StopListening() {
	rpc.eventListeningCancel()
}
