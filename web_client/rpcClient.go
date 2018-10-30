// +build js

package main

import (
	"context"
	"errors"
	"github.com/johanbrandhorst/protobuf/grpcweb"
	"github.com/mame82/P4wnP1_go/common_web"
	pb "github.com/mame82/P4wnP1_go/proto/gopherjs"
	"sync"
	"time"
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

func (rpc *Rpc) UploadContentToTempFile(timeout time.Duration, content []byte) (filename string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	//create hex string of content MD5 sum
	filename = BytesToMD5(content)

	//upload file to `/tmp/{md5_hash_hex}`
	_,err = rpc.Client.FSWriteFile(ctx,
		&pb.WriteFileRequest{
			Data:content,
			Append:false,
			Filename:filename,
			Folder: pb.AccessibleFolder_TMP,
			MustNotExist:false,
		})

	return
}

func (rpc *Rpc) RunHIDScriptJob(timeout time.Duration, filepath string) (job *pb.HIDScriptJob, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	//upload file to `/tmp/{md5_hash_hex}`
	return rpc.Client.HIDRunScriptJob(
		ctx,
		&pb.HIDScriptRequest{
			ScriptPath:     filepath,
			TimeoutSeconds: uint32(0),
		},
	)
}

func (rpc *Rpc) CancelHIDScriptJob(timeout time.Duration, jobID uint32) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_,err = rpc.Client.HIDCancelScriptJob(ctx, &pb.HIDScriptJob{
		Id:jobID,
	})
	return
}

func (rpc *Rpc) CancelAllHIDScriptJobs(timeout time.Duration) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_,err = rpc.Client.HIDCancelAllScriptJobs(ctx, &pb.Empty{})
	return
}

func (rpc *Rpc) GetStoredBluetoothSettingsList(timeout time.Duration) (ws []string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ma, err := rpc.Client.ListStoredBluetoothSettings(ctx, &pb.Empty{})
	if err != nil { return ws, err }
	return ma.MsgArray, err
}

func (rpc *Rpc) StoreBluetoothSettings(timeout time.Duration, req *pb.BluetoothRequestSettingsStorage) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err = rpc.Client.StoreBluetoothSettings(ctx, req)

	return
}

func (rpc *Rpc) GetStoredBluetoothSettings(timeout time.Duration, req *pb.StringMessage) (settings *pb.BluetoothSettings, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	settings, err = rpc.Client.GetStoredBluetoothSettings(ctx, req)

	return
}

func (rpc *Rpc) DeployStoredBluetoothSettings(timeout time.Duration, req *pb.StringMessage) (state *pb.BluetoothSettings, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	state, err = rpc.Client.DeployStoredBluetoothSettings(ctx, req)

	return
}


func (rpc *Rpc) DeleteStoredBluetoothSettings(timeout time.Duration, req *pb.StringMessage) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_, err = rpc.Client.DeleteStoredBluetoothSettings(ctx, req)
	return
}

func (rpc *Rpc) DeleteStoredUSBSettings(timeout time.Duration, req *pb.StringMessage) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_, err = rpc.Client.DeleteStoredUSBSettings(ctx, req)
	return
}

func (rpc *Rpc) DeleteStoredWifiSettings(timeout time.Duration, req *pb.StringMessage) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_, err = rpc.Client.DeleteStoredWifiSettings(ctx, req)
	return
}

func (rpc *Rpc) DeleteStoredEthernetInterfaceSettings(timeout time.Duration, req *pb.StringMessage) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_,err = rpc.Client.DeleteStoredEthernetInterfaceSettings(ctx, req)
	return
}

func (rpc *Rpc) DeleteStoredTriggerActionsSet(timeout time.Duration, name *pb.StringMessage) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_,err = rpc.Client.DeleteStoredTriggerActionSet(ctx, name)
	return
}


func (rpc *Rpc) GetBluetoothAgentSettings(timeout time.Duration) (res *jsBluetoothAgentSettings, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resRpc, err := rpc.Client.GetBluetoothAgentSettings(ctx, &pb.Empty{})
	if err != nil { return res, err }
	res = &jsBluetoothAgentSettings{Object:O()}
	res.fromGo(resRpc)
	return
}

func (rpc *Rpc) DeployBluetoothAgentSettings(timeout time.Duration, newSettings *jsBluetoothAgentSettings) (res *jsBluetoothAgentSettings, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resRpc, err := rpc.Client.DeployBluetoothAgentSettings(ctx, newSettings.toGo())
	if err != nil { return res, err }
	res = &jsBluetoothAgentSettings{Object:O()}
	res.fromGo(resRpc)
	return
}

func (rpc *Rpc) GetBluetoothControllerInformation(timeout time.Duration) (res *jsBluetoothControllerInformation, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	btCtlInfo, err := rpc.Client.GetBluetoothControllerInformation(ctx, &pb.Empty{})
	if err != nil { return res, err }
	res = &jsBluetoothControllerInformation{Object:O()}
	res.fromGo(btCtlInfo)
	return
}

func (rpc *Rpc) DeployBluetoothControllerInformation(timeout time.Duration, newSettings *jsBluetoothControllerInformation) (res *jsBluetoothControllerInformation, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	btCtlInfo, err := rpc.Client.DeployBluetoothControllerInformation(ctx, newSettings.toGo())
	if err != nil { return res, err }
	res = &jsBluetoothControllerInformation{Object:O()}
	res.fromGo(btCtlInfo)
	return
}

func (rpc *Rpc) GetStoredUSBSettingsList(timeout time.Duration) (ws []string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ma, err := rpc.Client.ListStoredUSBSettings(ctx, &pb.Empty{})
	if err != nil { return ws, err }
	return ma.MsgArray, err
}

func (rpc *Rpc) StoreUSBSettings(timeout time.Duration, req *pb.USBRequestSettingsStorage) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err = rpc.Client.StoreUSBSettings(ctx, req)

	return
}

func (rpc *Rpc) GetStoredUSBSettings(timeout time.Duration, req *pb.StringMessage) (settings *pb.GadgetSettings, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	settings, err = rpc.Client.GetStoredUSBSettings(ctx, req)

	return
}

func (rpc *Rpc) DeployStoredUSBSettings(timeout time.Duration, req *pb.StringMessage) (state *pb.GadgetSettings, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	state, err = rpc.Client.DeployStoredUSBSettings(ctx, req)

	return
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


func (rpc *Rpc) GetStoredEthernetInterfaceSettingsList(timeout time.Duration) (eis []string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ma, err := rpc.Client.ListStoredEthernetInterfaceSettings(ctx, &pb.Empty{})
	if err != nil { return eis, err }
	return ma.MsgArray, err
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

func (rpc *Rpc) DeployStoredWifiSettings(timeout time.Duration, req *pb.StringMessage) (state *pb.WiFiState, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	state, err = rpc.Client.DeployStoredWifiSettings(ctx, req)

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


func (rpc *Rpc) GetAllDeployedEthernetInterfaceSettings(timeout time.Duration) (settingsList *jsEthernetSettingsArray, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	deployedSettings, err := rpc.Client.GetAllDeployedEthernetInterfaceSettings(ctx, &pb.Empty{})
	if err != nil {
		println("Error GetAllDeployedEthernetInterfaceSettings", err)
		return nil, err
	}

	settingsList = &jsEthernetSettingsArray{Object: O()}
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

func (rpc *Rpc) StoreEthernetInterfaceSettings(timeout time.Duration, req *pb.EthernetRequestSettingsStorage) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err = rpc.Client.StoreEthernetInterfaceSettings(ctx, req)

	return
}

func (rpc *Rpc) GetStoredEthernetInterfaceSettings(timeout time.Duration, req *pb.StringMessage) (settings *pb.EthernetInterfaceSettings, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	settings, err = rpc.Client.GetStoredEthernetInterfaceSettings(ctx, req)

	return
}

func (rpc *Rpc) DeployStoredEthernetInterfaceSettings(timeout time.Duration, req *pb.StringMessage) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_,err = rpc.Client.DeployStoredEthernetInterfaceSettings(ctx, req)

	return
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

/*
func (rpc *Rpc) StartListening() {

	println("Start listening called", globalState.EventProcessor)

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
			for RpcClient.ConnectionTest(time.Millisecond * 2500) != nil {
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
					globalState.EventProcessor.HandleEvent(event)
				}
				// we end here on connection error
				evStream.CloseSend() // fix for half-open websockets, for which the server wouldn't send a TCP RST after crash/restart, as no active client to server communication takes place
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
*/

func (rpc *Rpc) StartEventListening(timeout time.Duration) (eventStream pb.P4WNP1_EventListenClient, cancel context.CancelFunc, err error) {

	println("Start listening called", globalState.EventProcessor)

	// Notes:
	// - rpc.Client.EventListen doesn't return an error if the gRPC server is not running or not reachable (we can't
	// cancel the context based on a timeout, as eventListen is meant to read an endless stream)
	// - in contrast, a call to a RPC method which isn't meant for server streaming, could fail after timeout
	// - to determine if the server is connectible at all, a connection test RPC method is called upfront
	// - additionally it should be noted, that even if the server streaming gRPC call to `EventListen` couldn't
	// detect that the server isn't connectible, a call to the `Recv()` method of the resulting stream object errors
	// in case an already existing server connection is lost (the server resets the underlying socket, but has to be running to do so)


	//Check if server is reachable (with timeout)
	for RpcClient.ConnectionTest(timeout) != nil {
		return eventStream,cancel,errors.New("Server not reachable")
	}


	println("... success")
	globalState.IsConnected = true
	globalState.FailedConnectionAttempts = 0

	ctx, cancel := context.WithCancel(context.Background())
	rpc.eventListeningCancel = cancel

	// try RPC call
	evStream, err := rpc.Client.EventListen(ctx, &pb.EventRequest{ListenType: common_web.EVT_ANY}) //No error if Websocket connection fails

	return evStream,cancel,err
}

