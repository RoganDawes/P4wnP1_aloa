// +build linux

package service

import (
	pb "github.com/mame82/P4wnP1_go/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/improbable-eng/grpc-web/go/grpcweb"

	"net/http"
	"path"
	"strings"
	"time"

	"github.com/mame82/P4wnP1_go/common"
	"io/ioutil"
	"os"
)

var (
	rpcErrNoHid = errors.New("HIDScript engine disabled, as current USB configuration has mouse and keyboard disable")
)

const (
	cSTORE_PREFIX_WIFI_SETTINGS      = "ws_"
	cSTORE_PREFIX_TRIGGER_ACTION_SET = "tas_"
)

type server struct {
	rootSvc *Service

	listenAddrGrpc string
	listenAddrWeb string
}

func (s *server) StoreTriggerActionSet(ctx context.Context, set *pb.TriggerActionSet) (e *pb.Empty, err error) {
	e = &pb.Empty{}
	err = s.rootSvc.SubSysDataStore.Put(cSTORE_PREFIX_TRIGGER_ACTION_SET+ set.Name, set, true)
	return
}

func (s *server) ListStoredTriggerActionSets(ctx context.Context, e *pb.Empty) (tas *pb.StringMessageArray, err error) {
	res, err := s.rootSvc.SubSysDataStore.KeysPrefix(cSTORE_PREFIX_TRIGGER_ACTION_SET, true)
	if err != nil {
		return tas, err
	}
	tas = &pb.StringMessageArray{
		MsgArray: res,
	}
	return
}

func (s *server) GetTriggerActionsState(context.Context, *pb.Empty) (*pb.TriggerActionSet, error) {
	return s.rootSvc.SubSysTriggerActions.GetCurrentTriggerActionSet(), nil
}

func (s *server) DeployTriggerActionSetReplace(ctx context.Context, tas *pb.TriggerActionSet) (resTas *pb.TriggerActionSet, err error) {
	// Clear old set, but keep immutables
	s.rootSvc.SubSysTriggerActions.ClearTriggerActions(true)
	// Add the new set
	_,err = s.DeployTriggerActionSetAdd(ctx, tas)
	if err != nil { return s.rootSvc.SubSysTriggerActions.GetCurrentTriggerActionSet(),err }
	return s.GetTriggerActionsState(ctx, &pb.Empty{})
}

func (s *server) DeployTriggerActionSetAdd(ctx context.Context, tas *pb.TriggerActionSet) (resTas *pb.TriggerActionSet, err error) {
	addedTA := make([]*pb.TriggerAction, len(tas.TriggerActions))
	for idx,ta := range tas.TriggerActions {
		addedTA[idx],err = s.rootSvc.SubSysTriggerActions.AddTriggerAction(ta)
		if err != nil { return s.rootSvc.SubSysTriggerActions.GetCurrentTriggerActionSet(),err }
	}

	return &pb.TriggerActionSet{TriggerActions:addedTA},nil
}

func NewRpcServerService(root *Service) *server {
	return &server{
		rootSvc:root,
	}
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
	if err != nil { return st,err }
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
	res,err := s.rootSvc.SubSysDataStore.KeysPrefix(cSTORE_PREFIX_WIFI_SETTINGS, true)
	if err != nil { return sa,err }
	sa = &pb.StringMessageArray{
		MsgArray: res,
	}
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
	return &pb.Empty{}, common.WriteFile(req.Path, req.MustNotExist, req.Append, req.Data)

}

func (s *server) FSReadFile(ctx context.Context, req *pb.ReadFileRequest) (resp *pb.ReadFileResponse, err error) {
	n,err := common.ReadFile(req.Path, req.Start, req.Data)
	resp = &pb.ReadFileResponse{ReadCount: int64(n)}
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
	targetJob,err := s.rootSvc.SubSysUSB.HidCtl.GetBackgroundJobByID(int(req.Id))
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
	if !s.rootSvc.SubSysUSB.Usable { return nil, ErrUsbNotUsable }

	if s.rootSvc.SubSysUSB.HidCtl == nil { return nil, rpcErrNoHid}

	retJobs,err := s.rootSvc.SubSysUSB.HidCtl.GetAllBackgroundJobs()
	if err != nil { return nil, err }
	jobs = &pb.HIDScriptJobList{}
	for _, aJob := range retJobs {
		jobs.Ids = append(jobs.Ids, uint32(aJob.Id))
	}
	return
}

func (s *server) HIDCancelAllScriptJobs(ctx context.Context, rEmpty *pb.Empty) (empty *pb.Empty, err error) {
	empty = &pb.Empty{}
	if s.rootSvc.SubSysUSB.HidCtl == nil { return empty, rpcErrNoHid}

	// Try to find script
	s.rootSvc.SubSysUSB.HidCtl.CancelAllBackgroundJobs()
	return
}



func (s *server) HIDCancelScriptJob(ctx context.Context, sJob *pb.HIDScriptJob) (empty *pb.Empty, err error) {
	empty = &pb.Empty{}
	if s.rootSvc.SubSysUSB.HidCtl == nil { return empty, rpcErrNoHid}

	// Try to find script
	job,err := s.rootSvc.SubSysUSB.HidCtl.GetBackgroundJobByID(int(sJob.Id))
	if err != nil { return empty, err }

	job.Cancel()
	return
}

func (s *server) HIDRunScript(ctx context.Context, scriptReq *pb.HIDScriptRequest) (scriptRes *pb.HIDScriptResult, err error) {
	if s.rootSvc.SubSysUSB.HidCtl == nil { return nil, rpcErrNoHid}



	if scriptFile, err := ioutil.ReadFile(scriptReq.ScriptPath); err != nil {
		return nil, errors.New(fmt.Sprintf("Couldn't load HIDScript '%s': %v\n", scriptReq.ScriptPath, err))
	} else {
		//jobCtx := context.Background()
		jobCtx := ctx //we want to interrupt the script if the gRPC client cancels
		// ToDo: we don't retrieve the cancelFunc which should be called to free resources. Solution: use withCancel context and call cancel by go routine on timeout
		if scriptReq.TimeoutSeconds > 0 { jobCtx,_ = context.WithTimeout(jobCtx, time.Second * time.Duration(scriptReq.TimeoutSeconds))}


		scriptVal,err := s.rootSvc.SubSysUSB.HidCtl.RunScript(jobCtx, string(scriptFile))
		if err != nil { return nil,err }
		val,_ := scriptVal.Export() //Convert to Go representation, error is always nil
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
}

func (s *server) HIDRunScriptJob(ctx context.Context, scriptReq *pb.HIDScriptRequest) (rJob *pb.HIDScriptJob, err error) {
	if s.rootSvc.SubSysUSB.HidCtl == nil { return nil, rpcErrNoHid}

	if scriptFile, err := ioutil.ReadFile(scriptReq.ScriptPath); err != nil {
		return nil, errors.New(fmt.Sprintf("Couldn't load HIDScript '%s': %v\n", scriptReq.ScriptPath, err))
	} else {
		//Note: Don't use the gRPC context, it would cancel after this call and thus interrupt the job immediately
		jobCtx := context.Background()
		// ToDo: we don't retrieve the cancelFunc which should be called to free resources. Solution: use withCancel context and call cancel by go routine on timeout
		if scriptReq.TimeoutSeconds > 0 { jobCtx,_ = context.WithTimeout(jobCtx, time.Second * time.Duration(scriptReq.TimeoutSeconds))}
		job,err := s.rootSvc.SubSysUSB.HidCtl.StartScriptAsBackgroundJob(jobCtx, string(scriptFile))
		if err != nil { return nil,err }

		rJob = &pb.HIDScriptJob{
			Id: uint32(job.Id),
		}
		return rJob,nil
	}
	return
}

func (s *server) HIDGetScriptJobResult(ctx context.Context, sJob *pb.HIDScriptJob) (scriptRes *pb.HIDScriptResult, err error) {
	if s.rootSvc.SubSysUSB.HidCtl == nil { return nil, rpcErrNoHid}

	// Try to find script
	job,err := s.rootSvc.SubSysUSB.HidCtl.GetBackgroundJobByID(int(sJob.Id))
	if err != nil { return scriptRes, err }


	//ToDo: check impact/behavior, because ctx is provided by gRPC server
	scriptVal,err := s.rootSvc.SubSysUSB.HidCtl.WaitBackgroundJobResult(ctx, job)
	if err != nil { return nil,err }
	val,_ := scriptVal.Export() //Convert to Go representation, error is always nil
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
	return
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
			fmt.Printf("legacy web req:\n %v\n", req)
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
