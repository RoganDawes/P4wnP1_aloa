// +build linux

package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/mame82/P4wnP1_go/common"
	"github.com/mame82/P4wnP1_go/common_web"
	pb "github.com/mame82/P4wnP1_go/proto"
	"github.com/mame82/P4wnP1_go/service/bluetooth"
	"github.com/mame82/mblue-toolz/toolz"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var (
	rpcErrNoHid = errors.New("HIDScript engine disabled, as current USB configuration has mouse and keyboard disable")
)

const (
	cSTORE_PREFIX_WIFI_SETTINGS      = "ws_"
	cSTORE_PREFIX_USB_SETTINGS      = "usbs_"
	cSTORE_PREFIX_ETHERNET_INTERFACE_SETTINGS      = "eis_"
	cSTORE_PREFIX_TRIGGER_ACTION_SET = "tas_"
	cSTORE_PREFIX_BLUETOOTH_SETTINGS = "bt_"
)


func NewRpcServerService(root *Service) *server {
	return &server{
		rootSvc:root,
	}
}

type server struct {
	rootSvc *Service

	listenAddrGrpc string
	listenAddrWeb string
}

func (s *server) WaitTriggerGroupReceive(rpcCtx context.Context, triggerGR *pb.TriggerGroupReceive) (e *pb.Empty, err error) {
	e = &pb.Empty{}
	triggerVal := triggerGR.Value
	triggerGroupName := triggerGR.GroupName

	//register a proper event listener
	evtRcv := s.rootSvc.SubSysEvent.RegisterReceiver(common_web.EVT_TRIGGER)
	defer evtRcv.Cancel()

Outer:
	for {
		select {
		case evt := <- evtRcv.EventQueue:
			// avoid consuming empty messages, because channel is closed
			if evt == nil {
				break Outer // abort loop on "nil" event, as this indicates the EventQueue channel has been closed
			}
			// check if received trigger event applies to TriggerGroupReceive
			if ttEvt := common_web.EvtTriggerType(evt.Values[0].GetTint64()); ttEvt == common_web.TRIGGER_EVT_TYPE_GROUP_RECEIVE {
				evGroupName,evValue,err := DeconstructEventTriggerGroupReceive(evt)
				if err != nil {
					continue // error parsing as groupReceiveEvent --> ignore
				}
				//check if group matches
				if evGroupName != triggerGroupName {
					continue // don't handle on group mismatch, but return without error
				}
				// check if received value matches
				if evValue != triggerVal {
					continue // don't handle on value mismatch, but return without error
				}

				//consume remaining events (shouldn't be necessary)
				//for len(evtRcv.EventQueue) > 0 { <- evtRcv.EventQueue }

				// if here, we have a hit and exit the loop without error
				break Outer
			}
		case <- evtRcv.Ctx.Done():
			// evvent Receiver cancelled or unregistered
			err = errors.New("EventListener for WaitTriggerGroupReceive aborted")
			break Outer
		case <- rpcCtx.Done():
			// evvent Receiver cancelled or unregistered
			err = errors.New("RPC call to WaitTriggerGroupReceive aborted")
			break Outer
		}
	}

/*
	if err != nil {
		fmt.Println("Aborted")
	}
*/
	return
}

func (s *server) FireActionGroupSend(ctx context.Context, gs *pb.ActionGroupSend) (e *pb.Empty, err error) {
	e = &pb.Empty{}
	s.rootSvc.SubSysEvent.Emit(ConstructEventTriggerGroupReceive(gs.GroupName, gs.Value))
	return
}

func (s *server) DeployBluetoothSettings(ctx context.Context, settings *pb.BluetoothSettings) (resultSettings *pb.BluetoothSettings, err error) {
	as := settings.As
	ci := settings.Ci
	resultSettings = &pb.BluetoothSettings{}
	resultSettings.Ci,err = s.DeployBluetoothControllerInformation(ctx, ci)
	if err != nil {
		resultSettings.As,_ = s.GetBluetoothAgentSettings(ctx,&pb.Empty{})
		return
	}
	resultSettings.As,err = s.DeployBluetoothAgentSettings(ctx, as)
	return
}

func (s *server) StoreBluetoothSettings(ctx context.Context, req *pb.BluetoothRequestSettingsStorage) (e *pb.Empty, err error) {
	e = &pb.Empty{}
	err = s.rootSvc.SubSysDataStore.Put(cSTORE_PREFIX_BLUETOOTH_SETTINGS + req.TemplateName, req.Settings, true)
	return
}

func (s *server) GetStoredBluetoothSettings(ctx context.Context, templateName *pb.StringMessage) (result *pb.BluetoothSettings, err error) {
	result = &pb.BluetoothSettings{}
	err = s.rootSvc.SubSysDataStore.Get(cSTORE_PREFIX_BLUETOOTH_SETTINGS + templateName.Msg, result)
	return
}

func (s *server) DeployStoredBluetoothSettings(ctx context.Context, templateName *pb.StringMessage) (e *pb.BluetoothSettings, err error) {
	bts,err := s.GetStoredBluetoothSettings(ctx,templateName)
	if err != nil { return bts,err }
	return s.DeployBluetoothSettings(ctx, bts)
}

func (s *server) DeleteStoredBluetoothSettings(ctx context.Context, templateName *pb.StringMessage) (e *pb.Empty, err error) {
	e = &pb.Empty{}
	err = s.rootSvc.SubSysDataStore.Delete(cSTORE_PREFIX_BLUETOOTH_SETTINGS + templateName.Msg)
	return
}

func (s *server) StoreDeployedBluetoothSettings(ctx context.Context, templateName *pb.StringMessage) (e *pb.Empty, err error) {
	e = &pb.Empty{}
	currentSettings := &pb.BluetoothSettings{}
	currentSettings.Ci,err = s.GetBluetoothControllerInformation(ctx, e)
	if err != nil { return e,err }
	currentSettings.As,err = s.GetBluetoothAgentSettings(ctx,e)
	if err != nil { return e,err }

	return s.StoreBluetoothSettings(ctx, &pb.BluetoothRequestSettingsStorage{
		Settings: currentSettings,
		TemplateName: templateName.Msg,
	})
}

func (s *server) ListStoredBluetoothSettings(ctx context.Context, e *pb.Empty) (sa *pb.StringMessageArray, err error) {
	sa = &pb.StringMessageArray{}
	res,err := s.rootSvc.SubSysDataStore.KeysPrefix(cSTORE_PREFIX_BLUETOOTH_SETTINGS, true)
	if err != nil { return sa,err }
	sa.MsgArray = res
	return
}

func (s *server) DeleteStoredUSBSettings(ctx context.Context, name *pb.StringMessage) (e *pb.Empty, err error) {
	e = &pb.Empty{}
	err = s.rootSvc.SubSysDataStore.Delete(cSTORE_PREFIX_USB_SETTINGS + name.Msg)
	return
}

func (s *server) DeleteStoredWifiSettings(ctx context.Context, name *pb.StringMessage) (e *pb.Empty, err error) {
	e = &pb.Empty{}
	err = s.rootSvc.SubSysDataStore.Delete(cSTORE_PREFIX_WIFI_SETTINGS + name.Msg)
	return
}

func (s *server) DeleteStoredEthernetInterfaceSettings(ctx context.Context, name *pb.StringMessage) (e *pb.Empty, err error) {
	e = &pb.Empty{}
	err = s.rootSvc.SubSysDataStore.Delete(cSTORE_PREFIX_ETHERNET_INTERFACE_SETTINGS + name.Msg)
	return
}

func (s *server) DeleteStoredTriggerActionSet(ctx context.Context, name *pb.StringMessage) (e *pb.Empty, err error) {
	e = &pb.Empty{}
	err = s.rootSvc.SubSysDataStore.Delete(cSTORE_PREFIX_TRIGGER_ACTION_SET + name.Msg)
	return
}

func (s *server) DBBackup(ctx context.Context, filename *pb.StringMessage) (e *pb.Empty, err error) {
	e = &pb.Empty{}
	fname := filename.Msg
	ext := filepath.Ext(fname)
	if lext := strings.ToLower(ext); lext != ".db" {
		fname = fname + ".db"
	}

	err = s.rootSvc.SubSysDataStore.Backup(PATH_DATA_STORE_BACKUP + "/" + fname)
	return
}

func (s *server) DBRestore(ctx context.Context, filename *pb.StringMessage) (e *pb.Empty, err error) {
	fmt.Println("DB restore: ", filename.Msg)
	e = &pb.Empty{}
	fname := filename.Msg
	ext := filepath.Ext(fname)
	if lext := strings.ToLower(ext); lext != ".db" {
		fname = fname + ".db"
	}
	err = s.rootSvc.SubSysDataStore.Restore(PATH_DATA_STORE_BACKUP + "/" + fname, true)
	return
}

func (s *server) ListStoredDBBackups(ctx context.Context, e *pb.Empty) (ma *pb.StringMessageArray, err error) {
	ma = &pb.StringMessageArray{}
	scripts,err := ListFilesOfFolder(PATH_DATA_STORE_BACKUP, ".db")
	if err != nil { return ma,err }
	ma.MsgArray = scripts
	return
}



func (s *server) GetBluetoothAgentSettings(ctx context.Context, e *pb.Empty) (as *pb.BluetoothAgentSettings, err error) {
	as = &pb.BluetoothAgentSettings{}

	pin,err := s.rootSvc.SubSysBluetooth.GetPIN()
	if err != nil { return as,err }
	as.Pin = pin
	return
}

func (s *server) DeployBluetoothAgentSettings(ctx context.Context, src *pb.BluetoothAgentSettings) (res *pb.BluetoothAgentSettings, err error) {
	err = s.rootSvc.SubSysBluetooth.SetPIN(src.Pin)
	return s.GetBluetoothAgentSettings(ctx, nil)
}

// Unused, Server services are deployed via BluetoothControllerInformation
func (s *server) SetBluetoothNetworkService(ctx context.Context, btNwSvc *pb.BluetoothNetworkService) (e *pb.Empty, err error) {
	e = &pb.Empty{}
	uuid := toolz.UUID_NETWORK_SERVER_NAP
	switch btNwSvc.Type {
	case pb.BluetoothNetworkServiceType_NAP:
		uuid = toolz.UUID_NETWORK_SERVER_NAP
	case pb.BluetoothNetworkServiceType_PANU:
		uuid = toolz.UUID_NETWORK_SERVER_PANU
	case pb.BluetoothNetworkServiceType_GN:
		uuid = toolz.UUID_NETWORK_SERVER_GN
	}
	if btNwSvc.ServerOrConnect {
		// start server for given network service
		if btNwSvc.RegisterOrUnregister {
			return e,s.rootSvc.SubSysBluetooth.RegisterNetworkServer(uuid)
		} else {
			return e,s.rootSvc.SubSysBluetooth.UnregisterNetworkServer(uuid)
		}
	} else {
		//(dis)connect from/to given network network service of given remote device

		if btNwSvc.RegisterOrUnregister {
			// register == connect
			return e,s.rootSvc.SubSysBluetooth.ConnectNetwork(btNwSvc.MacOrName, uuid)
		} else {
			// unregister == disconnect
			return e,s.rootSvc.SubSysBluetooth.DisconnectNetwork(btNwSvc.MacOrName)
		}
	}
}

func (s *server) DeployBluetoothControllerInformation(ctx context.Context, newBtCiRpc *pb.BluetoothControllerInformation) (updateBtCiRpc *pb.BluetoothControllerInformation, err error) {
	btCi := bluetooth.BluetoothControllerInformationFromRpc(newBtCiRpc)
	bridgeNameNap := BT_ETHERNET_BRIDGE_NAME
	bridgeNamePanu := BT_ETHERNET_BRIDGE_NAME
	bridgeNameGn := BT_ETHERNET_BRIDGE_NAME
	updatedCi,err := s.rootSvc.SubSysBluetooth.Controller.UpdateSettingsFromChangedControllerInformation(btCi, bridgeNameNap, bridgeNamePanu, bridgeNameGn)
	fmt.Printf("Deployed bluetooth settings\n%+v\n%v\n", updatedCi, err)
	if err != nil { return &pb.BluetoothControllerInformation{},err }
	updateBtCiRpc = bluetooth.BluetoothControllerInformationToRpc(updatedCi)
	return updateBtCiRpc, nil
}

func (s *server) GetBluetoothControllerInformation(ctx context.Context, e *pb.Empty) (res *pb.BluetoothControllerInformation, err error) {
	res = &pb.BluetoothControllerInformation{}
	btCi,err := s.rootSvc.SubSysBluetooth.Controller.ReadControllerInformation()
	if err != nil { return res,err }
	btCiRpc := bluetooth.BluetoothControllerInformationToRpc(btCi)
	btCiRpc.IsAvailable = s.rootSvc.SubSysBluetooth.IsServiceAvailable()

	return btCiRpc, nil
}

func (s *server) StoreUSBSettings(ctx context.Context, r *pb.USBRequestSettingsStorage) (e *pb.Empty, err error) {
	e = &pb.Empty{}
	err = s.rootSvc.SubSysDataStore.Put(cSTORE_PREFIX_USB_SETTINGS + r.TemplateName, r.Settings, true)
	return
}

func (s *server) GetStoredUSBSettings(ctx context.Context, m *pb.StringMessage) (gs *pb.GadgetSettings, err error) {
	gs = &pb.GadgetSettings{}
	err = s.rootSvc.SubSysDataStore.Get(cSTORE_PREFIX_USB_SETTINGS + m.Msg, gs)
	return
}

func (s *server) DeployStoredUSBSettings(ctx context.Context, m *pb.StringMessage) (st *pb.GadgetSettings, err error) {
	ws,err := s.GetStoredUSBSettings(ctx,m)
	if err != nil { return &pb.GadgetSettings{},err }
	st,err = s.SetGadgetSettings(ctx, ws)
	if err != nil {
		return
	}
	_,err = s.DeployGadgetSetting(ctx, &pb.Empty{})
	return
}

func (s *server) StoreDeployedUSBSettings(ctx context.Context, m *pb.StringMessage) (e *pb.Empty, err error) {
	gstate, err := ParseGadgetState(USB_GADGET_NAME)
	if err != nil { return &pb.Empty{},err }

	return s.StoreUSBSettings(ctx, &pb.USBRequestSettingsStorage{
		Settings: gstate,
		TemplateName: m.Msg,
	})
}

func (s *server) ListStoredUSBSettings(ctx context.Context, e *pb.Empty) (sa *pb.StringMessageArray, err error) {
	sa = &pb.StringMessageArray{}
	res,err := s.rootSvc.SubSysDataStore.KeysPrefix(cSTORE_PREFIX_USB_SETTINGS, true)
	if err != nil { return sa,err }
	sa.MsgArray = res
	return
}

func (s *server) ListStoredHIDScripts(context.Context, *pb.Empty) (sa *pb.StringMessageArray, err error) {
	sa = &pb.StringMessageArray{}
	scripts,err := ListFilesOfFolder(PATH_HID_SCRIPTS, ".js", ".javascript")
	if err != nil { return sa,err }
	sa.MsgArray = scripts
	return
}

func (s *server) ListStoredBashScripts(context.Context, *pb.Empty) (sa *pb.StringMessageArray, err error) {
	sa = &pb.StringMessageArray{}
	scripts,err := ListFilesOfFolder(PATH_BASH_SCRIPTS, ".sh", ".bash")
	if err != nil { return sa,err }
	sa.MsgArray = scripts
	return
}

func (s *server) DeployStoredTriggerActionSetReplace(ctx context.Context, msg *pb.StringMessage) (tas *pb.TriggerActionSet, err error) {
	// load set from store
	tas = &pb.TriggerActionSet{}
	err = s.rootSvc.SubSysDataStore.Get(cSTORE_PREFIX_TRIGGER_ACTION_SET + msg.Msg, tas)
	if err != nil { return }

	return s.DeployTriggerActionSetReplace(ctx,tas)
}

func (s *server) DeployStoredTriggerActionSetAdd(ctx context.Context, msg *pb.StringMessage) (tas *pb.TriggerActionSet, err error) {
	// load set from store
	tas = &pb.TriggerActionSet{}
	err = s.rootSvc.SubSysDataStore.Get(cSTORE_PREFIX_TRIGGER_ACTION_SET + msg.Msg, tas)
	if err != nil { return }

	return s.DeployTriggerActionSetAdd(ctx,tas)
}

func (s *server) StoreTriggerActionSet(ctx context.Context, set *pb.TriggerActionSet) (e *pb.Empty, err error) {
	e = &pb.Empty{}
	err = s.rootSvc.SubSysDataStore.Put(cSTORE_PREFIX_TRIGGER_ACTION_SET+ set.Name, set, true)
	return
}

func (s *server) ListStoredTriggerActionSets(ctx context.Context, e *pb.Empty) (tas *pb.StringMessageArray, err error) {
	tas = &pb.StringMessageArray{}
	res, err := s.rootSvc.SubSysDataStore.KeysPrefix(cSTORE_PREFIX_TRIGGER_ACTION_SET, true)
	if err != nil {
		return tas, err
	}
	tas.MsgArray = res
	return
}

func (s *server) GetDeployedTriggerActionSet(context.Context, *pb.Empty) (*pb.TriggerActionSet, error) {
	return s.rootSvc.SubSysTriggerActions.GetCurrentTriggerActionSet(), nil
}

func (s *server) DeployTriggerActionSetReplace(ctx context.Context, tas *pb.TriggerActionSet) (resTas *pb.TriggerActionSet, err error) {
	// Clear old set, but keep immutables
	s.rootSvc.SubSysTriggerActions.ClearTriggerActions(true)
	// Add the new set
	_,err = s.DeployTriggerActionSetAdd(ctx, tas)
	if err != nil { return s.rootSvc.SubSysTriggerActions.GetCurrentTriggerActionSet(),err }
	return s.GetDeployedTriggerActionSet(ctx, &pb.Empty{})
}

func (s *server) DeployTriggerActionSetAdd(ctx context.Context, tas *pb.TriggerActionSet) (resTas *pb.TriggerActionSet, err error) {
	addedTA := make([]*pb.TriggerAction, 0)
	for _,ta := range tas.TriggerActions {
		// we don't allow adding immutable settings via RPC call
		if !ta.Immutable {
			added,err := s.rootSvc.SubSysTriggerActions.AddTriggerAction(ta)
			if err != nil { return s.rootSvc.SubSysTriggerActions.GetCurrentTriggerActionSet(),err }
			addedTA = append(addedTA, added)
		}
	}

	resTas = &pb.TriggerActionSet{TriggerActions:addedTA, Name: "Added TriggerActions"}
	return
}

func (s *server) DeployTriggerActionSetRemove(ctx context.Context, removeTas *pb.TriggerActionSet) (removedTas *pb.TriggerActionSet, err error) {
	removedOnes := make([]*pb.TriggerAction,0)
	for _,removeTa := range removeTas.TriggerActions {
		removed,err := s.rootSvc.SubSysTriggerActions.RemoveTriggerAction(removeTa)
		if err != nil { return s.rootSvc.SubSysTriggerActions.GetCurrentTriggerActionSet(),err }
		removedOnes = append(removedOnes, removed)
	}

	removedTas = &pb.TriggerActionSet{TriggerActions:removedOnes, Name:"removed TriggerActions"}
	return
}

func (s *server) Start() error {
	return nil
}

func (s *server) Stop() error {
	return nil
}

func (s *server) StoreDeployedWifiSettings(ctx context.Context, m *pb.StringMessage) (e *pb.Empty, err error) {
	return s.StoreWifiSettings(ctx, &pb.WifiRequestSettingsStorage{
		Settings: s.rootSvc.SubSysWifi.State.CurrentSettings,
		TemplateName: m.Msg,
	})
}

func (s *server) DeployStoredWifiSettings(ctx context.Context, m *pb.StringMessage) (st *pb.WiFiState, err error) {
	ws,err := s.GetStoredWifiSettings(ctx,m)
	if err != nil { return &pb.WiFiState{},err }
	return s.DeployWiFiSettings(ctx, ws)
}

func (s *server) StoreWifiSettings(ctx context.Context, r *pb.WifiRequestSettingsStorage) (e *pb.Empty, err error) {
	e = &pb.Empty{}
	err = s.rootSvc.SubSysDataStore.Put(cSTORE_PREFIX_WIFI_SETTINGS + r.TemplateName, r.Settings, true)
	return
}

func (s *server) GetStoredWifiSettings(ctx context.Context, m *pb.StringMessage) (ws *pb.WiFiSettings, err error) {
	ws = &pb.WiFiSettings{}
	err = s.rootSvc.SubSysDataStore.Get(cSTORE_PREFIX_WIFI_SETTINGS + m.Msg, ws)
	return
}

func (s *server) ListStoredWifiSettings(ctx context.Context, e *pb.Empty) (sa *pb.StringMessageArray, err error) {
	sa = &pb.StringMessageArray{}
	res,err := s.rootSvc.SubSysDataStore.KeysPrefix(cSTORE_PREFIX_WIFI_SETTINGS, true)
	if err != nil { return sa,err }
	sa.MsgArray = res
	return
}

func (s *server) DeployWiFiSettings(ctx context.Context, wset *pb.WiFiSettings) (wstate *pb.WiFiState, err error) {
	return s.rootSvc.SubSysWifi.DeploySettings(wset)
}

func (s *server) GetWiFiState(ctx context.Context, empty *pb.Empty) (wstate *pb.WiFiState, err error) {
	// Update state before transmitting back
	s.rootSvc.SubSysWifi.UpdateStateFromIw()

	st := s.rootSvc.SubSysWifi.State
	return st, nil
}

func (s *server) ListenWiFiStateChanges(ctx context.Context, empty *pb.Empty) (wstate *pb.WiFiState, err error) {
	panic("implement me")
}

func (s *server) EchoRequest(ctx context.Context, req *pb.StringMessage) (resp *pb.StringMessage, err error) {
	return &pb.StringMessage{Msg:req.Msg}, nil
}

func (s *server) EventListen(eReq *pb.EventRequest, eStream pb.P4WNP1_EventListenServer) (err error) {
	//ToDo: check dependency from state (EvMgr initialized)
	rcv := s.rootSvc.SubSysEvent.RegisterReceiver(eReq.ListenType)

	for {
		select {
			case ev := <- rcv.EventQueue:
				//fmt.Printf("Event dequed to send: %+v\n", ev)

				//send Event to stream
				err = eStream.Send(ev)
				if err != nil {
					rcv.Cancel()
					log.Println(err)
					return err
				}

			case <-eStream.Context().Done():
				fmt.Println("Receiver aborted ...")
				rcv.Cancel()
				return errors.New("Event listening request aborted")
			case <-rcv.Ctx.Done():
				return errors.New("Service stopped event manager")
		}
	}
}

func (s *server) FSWriteFile(ctx context.Context, req *pb.WriteFileRequest) (empty *pb.Empty, err error) {
	filePath := "/" + req.Filename
	switch req.Folder {
	case pb.AccessibleFolder_TMP:
		filePath = "/tmp" + filePath
	case pb.AccessibleFolder_BASH_SCRIPTS:
		filePath = PATH_BASH_SCRIPTS + filePath
	case pb.AccessibleFolder_HID_SCRIPTS:
		filePath = PATH_HID_SCRIPTS + filePath
	default:
		err = errors.New("Unknown folder")
		return
	}

	return &pb.Empty{}, common.WriteFile(filePath, req.MustNotExist, req.Append, req.Data)

}

func (s *server) FSReadFile(ctx context.Context, req *pb.ReadFileRequest) (resp *pb.ReadFileResponse, err error) {
	//ToDo: check filename for path traversal attempts (don't care for security, currently - hey, we allow executing bash scripts as root - so what)

	filePath := "/" + req.Filename
	switch req.Folder {
	case pb.AccessibleFolder_TMP:
		filePath = "/tmp" + filePath
	case pb.AccessibleFolder_BASH_SCRIPTS:
		filePath = PATH_BASH_SCRIPTS + filePath
	case pb.AccessibleFolder_HID_SCRIPTS:
		filePath = PATH_HID_SCRIPTS + filePath
	default:
		err = errors.New("Unknown folder")
		return
	}

	chunk := make([]byte, req.Len)
	n,err := common.ReadFile(filePath, req.Start, chunk)
	if err == io.EOF { err = nil } //we ignore eof error, as eof is indicated by n = 0
	if err != nil {	return nil,err	}
	resp = &pb.ReadFileResponse{ReadCount: int64(n), Data: chunk[:n]}
	return
}

func (s *server) FSGetFileInfo(ctx context.Context, req *pb.FileInfoRequest) (resp *pb.FileInfoResponse, err error) {
	fi, err := os.Stat(req.Path)
	resp = &pb.FileInfoResponse{}
	if err != nil { return }
	resp.Name = fi.Name()
	resp.IsDir = fi.IsDir()
	resp.Mode = uint32(fi.Mode())
	resp.ModTime = fi.ModTime().Unix()
	resp.Size = fi.Size()
	return
}

func (s *server) FSCreateTempDirOrFile(ctx context.Context, req *pb.TempDirOrFileRequest) (resp *pb.TempDirOrFileResponse, err error) {
	resp = &pb.TempDirOrFileResponse{}
	if req.OnlyFolder {
		name, err := ioutil.TempDir(req.Dir, req.Prefix)
		if err != nil { return resp, err }
		resp.ResultPath = name
		return resp, err
	} else {
		var f *os.File
		f,err = ioutil.TempFile(req.Dir, req.Prefix)
		if err != nil { return resp,err }
		defer f.Close()
		resp.ResultPath = f.Name()
		return resp, err
	}
}

func (s *server) HIDGetRunningJobState(ctx context.Context, req *pb.HIDScriptJob) (res *pb.HIDRunningJobStateResult, err error) {
	targetJob,err := s.rootSvc.SubSysUSB.HidScriptGetBackgroundJobByID(int(req.Id))
	if err != nil { return nil, err }

	vmID,_ := targetJob.GetVMId() // ignore error, as VM ID would be -1 in error case

	//try to convert source to string
	source,ok := targetJob.Source.(string)
	if !ok { source = "Couldn't retrieve job's script source" }

	return &pb.HIDRunningJobStateResult{
		Id: int64(targetJob.Id),
		VmId: int64(vmID),
		Source: source,
	}, nil

}

func (s *server) HIDGetRunningScriptJobs(ctx context.Context, rEmpty *pb.Empty) (jobs *pb.HIDScriptJobList, err error) {
	retJobs,err := s.rootSvc.SubSysUSB.HidScriptGetAllRunningBackgroundJobs()
	if err != nil { return nil, err }
	jobs = &pb.HIDScriptJobList{}
	for _, aJob := range retJobs {
		jobs.Ids = append(jobs.Ids, uint32(aJob.Id))
	}
	return
}

func (s *server) HIDCancelAllScriptJobs(ctx context.Context, rEmpty *pb.Empty) (empty *pb.Empty, err error) {
	err = s.rootSvc.SubSysUSB.HidScriptCancelAllRunningBackgroundJobs()
	return
}

func (s *server) HIDCancelScriptJob(ctx context.Context, sJob *pb.HIDScriptJob) (empty *pb.Empty, err error) {
	job,err := s.rootSvc.SubSysUSB.HidScriptGetBackgroundJobByID(int(sJob.Id))
	if err != nil { return nil, err }

	job.Cancel()
	return
}

func (s *server) HIDRunScript(ctx context.Context, scriptReq *pb.HIDScriptRequest) (scriptRes *pb.HIDScriptResult, err error) {
	err = s.rootSvc.SubSysUSB.HidScriptUsable()
	if err != nil { return }

	scriptFile, err := ioutil.ReadFile(scriptReq.ScriptPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Couldn't load HIDScript '%s': %v\n", scriptReq.ScriptPath, err))
	}

	// ToDo: we don't retrieve the cancelFunc which should be called to free resources. Solution: use withCancel context and call cancel by go routine on timeout
	if scriptReq.TimeoutSeconds > 0 { ctx,_ = context.WithTimeout(ctx, time.Second * time.Duration(scriptReq.TimeoutSeconds))}

	val,err := s.rootSvc.SubSysUSB.HidScriptRun(ctx, string(scriptFile))
	if err != nil { return nil,err }

	if jsonVal,err := json.Marshal(val); err == nil {
		scriptRes = &pb.HIDScriptResult{
			IsFinished: true,
			Job: &pb.HIDScriptJob{Id:0},
			ResultJson: string(jsonVal),
		}
		return scriptRes,nil
	} else {
		return nil, errors.New(fmt.Sprintf("Script seems to have succeeded but result couldn't be converted to JSON: %v\n", err))
	}

}

func (s *server) HIDRunScriptJob(ctx context.Context, scriptReq *pb.HIDScriptRequest) (rJob *pb.HIDScriptJob, err error) {
	err = s.rootSvc.SubSysUSB.HidScriptUsable()
	if err != nil { return }

	scriptFile, err := ioutil.ReadFile(scriptReq.ScriptPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Couldn't load HIDScript '%s': %v\n", scriptReq.ScriptPath, err))
	}

	//Note: Don't use the gRPC context, it would cancel after this call and thus interrupt the job immediately
	jobCtx := context.Background()
	// ToDo: we don't retrieve the cancelFunc which should be called to free resources. Solution: use withCancel context and call cancel by go routine on timeout
	if scriptReq.TimeoutSeconds > 0 { jobCtx,_ = context.WithTimeout(jobCtx, time.Second * time.Duration(scriptReq.TimeoutSeconds))}
	job,err := s.rootSvc.SubSysUSB.HidScriptStartBackground(jobCtx, string(scriptFile))
	if err != nil { return nil,err }

	rJob = &pb.HIDScriptJob{
		Id: uint32(job.Id),
	}
	return rJob,nil
}

func (s *server) HIDGetScriptJobResult(ctx context.Context, sJob *pb.HIDScriptJob) (scriptRes *pb.HIDScriptResult, err error) {
	// Try to find script
	job,err := s.rootSvc.SubSysUSB.HidScriptGetBackgroundJobByID(int(sJob.Id))
	if err != nil { return nil, err }

	val,err := s.rootSvc.SubSysUSB.HidScriptWaitBackgroundJobResult(ctx, job)
	if err != nil { return nil,err }
	jsonVal,err := json.Marshal(val)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Script seems to have succeeded but result couldn't be converted to JSON: %v\n", err))
	}
	scriptRes = &pb.HIDScriptResult{
		IsFinished: true,
		Job: &pb.HIDScriptJob{Id:0},
		ResultJson: string(jsonVal),
	}
	return scriptRes,nil
}

func (s *server) GetDeployedEthernetInterfaceSettings(ctx context.Context, req *pb.StringMessage) (resp *pb.EthernetInterfaceSettings, err error) {
	if mi,err := s.rootSvc.SubSysNetwork.GetManagedInterface(req.Msg); err == nil {
		return mi.GetState().CurrentSettings, nil
	} else {
		return nil, errors.New(fmt.Sprintf("No stored (or used) settings for ethernet interface '%s'", req.Msg))
	}
	/*
	if settings,exist := ServiceState.StoredNetworkSettings[req.Msg]; exist && settings.SettingsInUse {
		return settings, nil
	} else {
		return nil, errors.New(fmt.Sprintf("No stored (or used) settings for ethernet interface '%s'", req.Msg))
	}
	*/
}

func (s *server) GetAllDeployedEthernetInterfaceSettings(ctx context.Context, empty *pb.Empty) (resp *pb.DeployedEthernetInterfaceSettings, err error) {
	miList := s.rootSvc.SubSysNetwork.GetManagedInterfaceNames()
	deployed := make([]*pb.EthernetInterfaceSettings,len(miList))
	for idx,name := range miList {
		mi,err := s.rootSvc.SubSysNetwork.GetManagedInterface(name)
		if err != nil { return nil,err }
		deployed[idx] = mi.GetState().CurrentSettings
	}
	resp = &pb.DeployedEthernetInterfaceSettings{
		List: deployed,
	}
	return resp, nil
}

func (s *server) DeployEthernetInterfaceSettings(ctx context.Context, es *pb.EthernetInterfaceSettings) (empty *pb.Empty, err error) {
	log.Printf("Trying to deploy ethernet interface settings %v", es)

	empty = &pb.Empty{}
	iname := es.Name
	nim,err := s.rootSvc.SubSysNetwork.GetManagedInterface(iname)
	if err != nil { return empty,err }

	err = nim.DeploySettings(es)
	if err != nil {
		log.Printf("Error deploying ethernet interface settings %v", err)
	}
	return
}

func (s *server) StoreEthernetInterfaceSettings(ctx context.Context, req *pb.EthernetRequestSettingsStorage) (empty *pb.Empty, err error) {
	empty = &pb.Empty{}
	ifName := req.Settings.Name
	storageKey := cSTORE_PREFIX_ETHERNET_INTERFACE_SETTINGS + ifName + "_" + req.TemplateName
	err = s.rootSvc.SubSysDataStore.Put(storageKey, req.Settings, true)
	return
}

func (s *server) GetStoredEthernetInterfaceSettings(ctx context.Context, m *pb.StringMessage) (eis *pb.EthernetInterfaceSettings, err error) {
	eis = &pb.EthernetInterfaceSettings{}
	err = s.rootSvc.SubSysDataStore.Get(cSTORE_PREFIX_ETHERNET_INTERFACE_SETTINGS + m.Msg, eis)
	return
}

func (s *server) DeployStoredEthernetInterfaceSettings(ctx context.Context, msg *pb.StringMessage) (empty *pb.Empty, err error) {
	eis,err := s.GetStoredEthernetInterfaceSettings(ctx, msg)
	if err != nil { return empty,err }
	return s.DeployEthernetInterfaceSettings(ctx, eis)
}

func (s *server) ListStoredEthernetInterfaceSettings(ctx context.Context, empty *pb.Empty) (messages *pb.StringMessageArray, err error) {
	messages = &pb.StringMessageArray{}
	res,err := s.rootSvc.SubSysDataStore.KeysPrefix(cSTORE_PREFIX_ETHERNET_INTERFACE_SETTINGS, true)
	if err != nil { return messages,err }
	messages.MsgArray = res
	return
}

func (s *server) MountUMSFile(ctx context.Context, gsu *pb.GadgetSettingsUMS) (*pb.Empty, error) {
	log.Printf("Trying to mount iamge `%s` to UMS ...", gsu.File)
	err := MountUMSFile(gsu.File)
	return nil, err
}

func (s *server) GetDeployedGadgetSetting(ctx context.Context, e *pb.Empty) (gs *pb.GadgetSettings, err error) {
	gs, err = ParseGadgetState(USB_GADGET_NAME)

	gs.DevPathHidKeyboard = s.rootSvc.SubSysUSB.State.DevicePath[USB_FUNCTION_HID_KEYBOARD_name]
	gs.DevPathHidMouse = s.rootSvc.SubSysUSB.State.DevicePath[USB_FUNCTION_HID_MOUSE_name]
	gs.DevPathHidMouse = s.rootSvc.SubSysUSB.State.DevicePath[USB_FUNCTION_HID_RAW_name]

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

	errg := s.rootSvc.SubSysUSB.DeployGadgetSettings(s.rootSvc.SubSysUSB.State.UndeployedGadgetSettings)
	err = nil
	if errg != nil {
		err = errors.New(fmt.Sprintf("Deploying new gadget settings failed, reverted to old ones: %v", errg))
		s.rootSvc.SubSysUSB.DeployGadgetSettings(gs_backup) //We don't catch the error, as the old settings should have been working
	}

	gs, _ = ParseGadgetState(USB_GADGET_NAME) //Return settings from deployed gadget
	return
}

func (s *server) GetGadgetSettings(context.Context, *pb.Empty) (*pb.GadgetSettings, error) {
	return s.rootSvc.SubSysUSB.State.UndeployedGadgetSettings, nil
}

func (s *server) SetGadgetSettings(ctx context.Context, gs *pb.GadgetSettings) (res *pb.GadgetSettings, err error) {
	if err = ValidateGadgetSetting(*gs); err != nil {
		//We return the validation error and the current (unchanged) GadgetSettingsState
		res = s.rootSvc.SubSysUSB.State.UndeployedGadgetSettings
		return
	}
	s.rootSvc.SubSysUSB.State.UndeployedGadgetSettings = gs
	res = s.rootSvc.SubSysUSB.State.UndeployedGadgetSettings
	return
}

func (s *server) GetLEDSettings(context.Context, *pb.Empty) (res *pb.LEDSettings, err error) {
//	res, err = ServiceState.Led.GetLed()
	state := s.rootSvc.SubSysLed.GetState()
	res = &pb.LEDSettings{
		BlinkCount: *state.BlinkCount,
	}
	log.Printf("GetLEDSettings, result: %+v", res)
	return
}

func (s *server) SetLEDSettings(ctx context.Context, ls *pb.LEDSettings) (*pb.Empty, error) {
	log.Printf("SetLEDSettings %+v", ls)
	s.rootSvc.SubSysLed.DeploySettings(ls)
	return &pb.Empty{}, nil
}

/*
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
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
*/

func folderReader(fn http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.URL.Path, "/") {
			// Use contents of index.html for directory, if present.
			req.URL.Path = path.Join(req.URL.Path, "index.html")
		}
		fn.ServeHTTP(w, req)
	}
}

/*
func StartRpcWebServer(host string, port string) {
	//Create gRPC Server
	s := grpc.NewServer()
	pb.RegisterP4WNP1Server(s, &server{})

	//grpc_web_srv := grpcweb.WrapServer(s, grpcweb.WithWebsockets(true)) //Wrap server to improbable grpc-web with websockets
	grpc_web_srv := grpcweb.WrapServer(s) //Wrap server to improbable grpc-web with websockets

	http_handler := func(resp http.ResponseWriter, req *http.Request) {
		grpc_web_srv.ServeHTTP(resp, req)
	}

	listen_address := host + ":" + port
	http_srv := &http.Server{
		Addr: listen_address,
		Handler: http.HandlerFunc(http_handler),
		//ReadHeaderTimeout: 5*time.Second,
		//IdleTimeout: 120*time.Second,
	}


	//Open TCP listener
	log.Printf("P4wnP1 gRPC-web server listening on " + listen_address)
	log.Fatal(http_srv.ListenAndServe())
}
*/

func (srv *server) StartRpcServerAndWeb(host string, gRPCPort string, webPort string, absWebRoot string) () {
	//ToDo: Return servers/TCP listener to allow closing from caller
	listen_address_grpc := host + ":" + gRPCPort
	listen_address_web := host + ":" + webPort


	//Create gRPC Server
	s := grpc.NewServer()
	pb.RegisterP4WNP1Server(s, srv)



	//Open TCP listener
	lis, err := net.Listen("tcp", listen_address_grpc)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// run gRPC server in go routine
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
	log.Printf("P4wnP1 gRPC server listening on " + listen_address_grpc)


	//Wrap the server into a gRPC-web server
	grpc_web_srv := grpcweb.WrapServer(s, grpcweb.WithWebsockets(true)) //Wrap server to improbable grpc-web with websockets
	//define a handler for a HTTP web server using the gRPC-web proxy
	http_gRPC_web_handler := func(resp http.ResponseWriter, req *http.Request) {
		//fmt.Printf("===========\nRequest: %s\n %v\n=============\n", req)
		if strings.Contains(req.Header.Get("Content-Type"), "application/grpc") ||
			req.Method == "OPTIONS" ||
			strings.Contains(req.Header.Get("Sec-Websocket-Protocol"), "grpc-websockets") {
			//fmt.Printf("gRPC-web req:\n %v\n", req)
			grpc_web_srv.ServeHTTP(resp, req) // if content type indicates grpc or REQUEST METHOD IS OPTIONS (pre-flight) serve gRPC-web
		} else {
			fmt.Printf("legacy web req: %v\n", req.RequestURI)
			http.FileServer(http.Dir((absWebRoot))).ServeHTTP(resp, req)
		}
	}
	//Setup our HTTP server
	http_srv := &http.Server{
		Addr: listen_address_web, //listen on port 80 with webservice
		Handler: http.HandlerFunc(http_gRPC_web_handler),
		ReadHeaderTimeout: 5*time.Second,
		IdleTimeout: 120*time.Second,
	}

	go func() {
		if err_http := http_srv.ListenAndServe(); err_http != nil {
			log.Fatal(err)
		}
	}()
	log.Printf("P4wnP1 gRPC-web server listening on " + http_srv.Addr)
}
