// +build js

package main

import (
	"context"
	"github.com/johanbrandhorst/protobuf/grpcweb"
	"github.com/mame82/P4wnP1_go/common_web"
	pb "github.com/mame82/P4wnP1_go/proto/gopherjs"
	"io"
	"sync"
	"time"
	"errors"
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

func (rpc *Rpc) UploadBytesToFile(timeout time.Duration, filename string, folder pb.AccessibleFolder, content []byte, allowOverwrite bool) (err error) {
	ctx := context.Background()
	if timeout > 0 {
		newCtx, cancel := context.WithTimeout(ctx, timeout)
		ctx = newCtx
		defer cancel()
	}

	_, err = rpc.Client.FSWriteFile(ctx, &pb.WriteFileRequest{
		Data: content,
		Folder: folder,
		Filename: filename,
		Append: false,
		MustNotExist: !allowOverwrite,
	})
	return err
}

// Warning, this method reads content completely to RAM
func (rpc *Rpc) DownloadFileToBytes(timeout time.Duration, filename string, folder pb.AccessibleFolder) (content []byte, err error) {
	ctx := context.Background()
	if timeout > 0 {
		newCtx, cancel := context.WithTimeout(ctx, timeout)
		ctx = newCtx
		defer cancel()
	}

	chunksize := int64(1 << 15)
	readCount := chunksize

	for readCount >= chunksize {
		resp, err := rpc.Client.FSReadFile(ctx, &pb.ReadFileRequest{
			Filename: filename,
			Folder: folder,
			Start: int64(len(content)),
			Len:   chunksize,
		})
		if err != nil { return content,err }
		content = append(content, resp.Data...)
		readCount = resp.ReadCount
	}

	return
}


func (rpc *Rpc) GetStoredBashScriptsList(timeout time.Duration) (ws []string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ma, err := rpc.Client.ListStoredBashScripts(ctx, &pb.Empty{})
	if err != nil { return ws, err }
	return ma.MsgArray, err
}

func (rpc *Rpc) GetStoredHIDScriptsList(timeout time.Duration) (ws []string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ma, err := rpc.Client.ListStoredHIDScripts(ctx, &pb.Empty{})
	if err != nil { return ws, err }
	return ma.MsgArray, err
}


func (rpc *Rpc) DeployedEthernetInterfaceSettings(timeout time.Duration, settings *pb.EthernetInterfaceSettings) (err error) {
	// ToDo: The RPC call has to return an error in case deployment fails
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err = rpc.Client.DeployEthernetInterfaceSettings(ctx, settings)
	return
}


func (rpc *Rpc) GetStoredWifiSettingsList(timeout time.Duration) (ws []string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ma, err := rpc.Client.ListStoredWifiSettings(ctx, &pb.Empty{})
	if err != nil { return ws, err }
	return ma.MsgArray, err
}

func (rpc *Rpc) DeployWifiSettings(timeout time.Duration, settings *pb.WiFiSettings) (state *pb.WiFiState, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	state, err = rpc.Client.DeployWiFiSettings(ctx, settings)
	return
}

func (rpc *Rpc) StoreWifiSettings(timeout time.Duration, req *pb.WifiRequestSettingsStorage) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err = rpc.Client.StoreWifiSettings(ctx, req)

	return
}

func (rpc *Rpc) GetStoredWifiSettings(timeout time.Duration, req *pb.StringMessage) (settings *pb.WiFiSettings, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	settings, err = rpc.Client.GetStoredWifiSettings(ctx, req)

	return
}


func (rpc *Rpc) GetWifiState(timeout time.Duration) (state *jsWiFiState, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ws,err := rpc.Client.GetWiFiState(ctx, &pb.Empty{})
	if err != nil {
		println("Error GetDeployedWifiSettings", err)
		return nil, err
	}

	println("GetWifiState: ", ws)

	// Update state

	state = &jsWiFiState{Object:O()}
	state.fromGo(ws)

	return

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

func (rpc *Rpc) GetRunningHidJobStates(timeout time.Duration) (states []*pb.HIDRunningJobStateResult, err error) {
	println("GetRunningHidJobStates called")

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

func (rpc *Rpc) ListStoredTriggerActionSets(timeout time.Duration) (tasNames []string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	msgArray,err := rpc.Client.ListStoredTriggerActionSets(ctx, &pb.Empty{})
	if err != nil { return tasNames, err}
	return msgArray.MsgArray, nil
}

func (rpc *Rpc) StoreTriggerActionSet(timeout time.Duration, set *pb.TriggerActionSet) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_,err = rpc.Client.StoreTriggerActionSet(ctx, set)
	return err
}

func (rpc *Rpc) GetDeployedTriggerActionSet(timeout time.Duration) (res *pb.TriggerActionSet, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return rpc.Client.GetDeployedTriggerActionSet(ctx, &pb.Empty{})
}

func (rpc *Rpc) DeployTriggerActionsSetReplace(timeout time.Duration, set *pb.TriggerActionSet) (res *pb.TriggerActionSet, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return rpc.Client.DeployTriggerActionSetReplace(ctx, set)
}

func (rpc *Rpc) DeployTriggerActionsSetRemove(timeout time.Duration, set *pb.TriggerActionSet) (res *pb.TriggerActionSet, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return rpc.Client.DeployTriggerActionSetRemove(ctx, set)
}

func (rpc *Rpc) DeployTriggerActionsSetAdd(timeout time.Duration, set *pb.TriggerActionSet) (res *pb.TriggerActionSet, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return rpc.Client.DeployTriggerActionSetAdd(ctx, set)
}

func (rpc *Rpc) DeployStoredTriggerActionsSetReplace(timeout time.Duration, name *pb.StringMessage) (res *pb.TriggerActionSet, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return rpc.Client.DeployStoredTriggerActionSetReplace(ctx, name)
}

func (rpc *Rpc) DeployStoredTriggerActionsSetAdd(timeout time.Duration, name *pb.StringMessage) (res *pb.TriggerActionSet, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return rpc.Client.DeployStoredTriggerActionSetAdd(ctx, name)
}



func (rpc *Rpc) GetDeployedGadgetSettings(timeout time.Duration) (*pb.GadgetSettings, error) {
	//gs := vue.GetVM(c).Get("gadgetSettings")
	println("GetDeployedGadgetSettings called")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return rpc.Client.GetDeployedGadgetSetting(ctx, &pb.Empty{})

}

func (rpc *Rpc) SetRemoteGadgetSettings(targetGS *pb.GadgetSettings, timeout time.Duration) (err error) {
	//gs := vue.GetVM(c).Get("gadgetSettings")
	println("SetRemoteGadgetSettings called")

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

func (rpc *Rpc) DeployRemoteGadgetSettings(timeout time.Duration) (*pb.GadgetSettings, error) {
	//gs := vue.GetVM(c).Get("gadgetSettings")
	println("DeployRemoteGadgetSettings called")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return rpc.Client.DeployGadgetSetting(ctx, &pb.Empty{})

}

func (rpc *Rpc) ConnectionTest(timeout time.Duration) (err error) {
	//gs := vue.GetVM(c).Get("gadgetSettings")
	println("DeployRemoteGadgetSettings called")

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
