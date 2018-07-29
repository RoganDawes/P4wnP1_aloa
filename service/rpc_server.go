//package rpcserv
package service

import (
	"log"
	pb "github.com/mame82/P4wnP1_go/proto"
	"context"
	"net"
	"google.golang.org/grpc"
//	"google.golang.org/grpc/reflection"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/improbable-eng/grpc-web/go/grpcweb"

	"net/http"
	"strings"
	"path"
	"time"

	"../common"
	"os"
	"io/ioutil"
)

var (
	rpcErrNoHid = errors.New("HIDScript engine disabled, as current USB configuration has mouse and keyboard disable")
)

type server struct {}

func (s *server) EventListen(eReq *pb.EventRequest, eStream pb.P4WNP1_EventListenServer) (err error) {
	//ToDo: check dependency from state (EvMgr initialized)
	rcv := ServiceState.EvMgr.RegisterReceiver(eReq.ListenType)

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
		f,err := ioutil.TempFile(req.Dir, req.Prefix)
		if err != nil { return resp,err }
		defer f.Close()
		resp.ResultPath = f.Name()
		return resp, err
	}
}

func (s *server) HIDGetRunningScriptJobs(ctx context.Context, rEmpty *pb.Empty) (jobs *pb.HIDScriptJobList, err error) {
	if ServiceState.UsbGM.HidCtl == nil { return nil, rpcErrNoHid}

	retJobs,err := ServiceState.UsbGM.HidCtl.GetAllBackgroundJobs()
	if err != nil { return nil, err }
	jobs = &pb.HIDScriptJobList{}
	jobs.Ids = retJobs
	return
}

func (s *server) HIDCancelAllScriptJobs(ctx context.Context, rEmpty *pb.Empty) (empty *pb.Empty, err error) {
	empty = &pb.Empty{}
	if ServiceState.UsbGM.HidCtl == nil { return empty, rpcErrNoHid}

	// Try to find script
	ServiceState.UsbGM.HidCtl.CancelAllBackgroundJobs()
	return
}



func (s *server) HIDCancelScriptJob(ctx context.Context, sJob *pb.HIDScriptJob) (empty *pb.Empty, err error) {
	empty = &pb.Empty{}
	if ServiceState.UsbGM.HidCtl == nil { return empty, rpcErrNoHid}

	// Try to find script
	job,err := ServiceState.UsbGM.HidCtl.GetBackgroundJobByID(int(sJob.Id))
	if err != nil { return empty, err }

	job.Cancel()
	return
}

func (s *server) HIDRunScript(ctx context.Context, scriptReq *pb.HIDScriptRequest) (scriptRes *pb.HIDScriptResult, err error) {
	if ServiceState.UsbGM.HidCtl == nil { return nil, rpcErrNoHid}



	if scriptFile, err := ioutil.ReadFile(scriptReq.ScriptPath); err != nil {
		return nil, errors.New(fmt.Sprintf("Couldn't load HIDScript '%s': %v\n", scriptReq.ScriptPath, err))
	} else {
		//jobCtx := context.Background()
		jobCtx := ctx //we want to interrupt the script if the gRPC client cancels
		// ToDo: we don't retrieve the cancelFunc which should be called to free resources. Solution: use withCancel context and call cancel by go routine on timeout
		if scriptReq.TimeoutSeconds > 0 { jobCtx,_ = context.WithTimeout(jobCtx, time.Second * time.Duration(scriptReq.TimeoutSeconds))}


		scriptVal,err := ServiceState.UsbGM.HidCtl.RunScript(jobCtx, string(scriptFile))
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
	if ServiceState.UsbGM.HidCtl == nil { return nil, rpcErrNoHid}

	if scriptFile, err := ioutil.ReadFile(scriptReq.ScriptPath); err != nil {
		return nil, errors.New(fmt.Sprintf("Couldn't load HIDScript '%s': %v\n", scriptReq.ScriptPath, err))
	} else {
		//Note: Don't use the gRPC context, it would cancel after this call and thus interrupt the job immediately
		jobCtx := context.Background()
		// ToDo: we don't retrieve the cancelFunc which should be called to free resources. Solution: use withCancel context and call cancel by go routine on timeout
		if scriptReq.TimeoutSeconds > 0 { jobCtx,_ = context.WithTimeout(jobCtx, time.Second * time.Duration(scriptReq.TimeoutSeconds))}
		job,err := ServiceState.UsbGM.HidCtl.StartScriptAsBackgroundJob(jobCtx, string(scriptFile))
		if err != nil { return nil,err }

		rJob = &pb.HIDScriptJob{
			Id: uint32(job.Id),
		}
		return rJob,nil
	}
	return
}

func (s *server) HIDGetScriptJobResult(ctx context.Context, sJob *pb.HIDScriptJob) (scriptRes *pb.HIDScriptResult, err error) {
	if ServiceState.UsbGM.HidCtl == nil { return nil, rpcErrNoHid}

	// Try to find script
	job,err := ServiceState.UsbGM.HidCtl.GetBackgroundJobByID(int(sJob.Id))
	if err != nil { return scriptRes, err }


	//ToDo: check impact/behavior, because ctx is provided by gRPC server
	scriptVal,err := ServiceState.UsbGM.HidCtl.WaitBackgroundJobResult(ctx, job)
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

func (s *server) DeployWifiSettings(ctx context.Context, ws *pb.WiFiSettings) (empty *pb.Empty, err error) {
	log.Printf("Trying to deploy WiFi settings %v", ws)
	empty = &pb.Empty{}
	err = DeployWifiSettings(ws)
	if err != nil {
		log.Printf("Error deploying WiFi settings settings %v", err)
	}
	return
}

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

	errg := ServiceState.UsbGM.DeployGadgetSettings(ServiceState.UsbGM.UndeployedGadgetSettings)
	err = nil
	if errg != nil {
		err = errors.New(fmt.Sprintf("Deploying new gadget settings failed, reverted to old ones: %v", errg))
		ServiceState.UsbGM.DeployGadgetSettings(gs_backup) //We don't catch the error, as the old settings should have been working
	}

	gs, _ = ParseGadgetState(USB_GADGET_NAME) //Return settings from deployed gadget
	return
}

func (s *server) GetGadgetSettings(context.Context, *pb.Empty) (*pb.GadgetSettings, error) {
	return ServiceState.UsbGM.UndeployedGadgetSettings, nil
}

func (s *server) SetGadgetSettings(ctx context.Context, gs *pb.GadgetSettings) (res *pb.GadgetSettings, err error) {
	if err = ValidateGadgetSetting(*gs); err != nil {
		//We return the validation error and the current (unchanged) GadgetSettingsState
		res = ServiceState.UsbGM.UndeployedGadgetSettings
		return
	}
	ServiceState.UsbGM.UndeployedGadgetSettings = gs
	res = ServiceState.UsbGM.UndeployedGadgetSettings
	return
}

func (s *server) GetLEDSettings(context.Context, *pb.Empty) (res *pb.LEDSettings, err error) {
	res, err = ServiceState.Led.GetLed()
	log.Printf("GetLEDSettings, result: %+v", res)
	return
}

func (s *server) SetLEDSettings(ctx context.Context, ls *pb.LEDSettings) (*pb.Empty, error) {
	log.Printf("SetLEDSettings %+v", ls)
	ServiceState.Led.SetLed(ls)
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
	/*
	// Register reflection service on gRPC server.
	reflection.Register(s)
	*/
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func folderReader(fn http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.URL.Path, "/") {
			// Use contents of index.html for directory, if present.
			req.URL.Path = path.Join(req.URL.Path, "index.html")
		}
		fn.ServeHTTP(w, req)
	}
}

func StartRpcWebServer(host string, port string) {
	//Create gRPC Server
	s := grpc.NewServer()
	pb.RegisterP4WNP1Server(s, &server{})

	//grpc_web_srv := grpcweb.WrapServer(s, grpcweb.WithWebsockets(true)) //Wrap server to improbable grpc-web with websockets
	grpc_web_srv := grpcweb.WrapServer(s) //Wrap server to improbable grpc-web with websockets

	/*
	http_handler := func(resp http.ResponseWriter, req *http.Request) {
		if req.ProtoMajor == 2 && strings.Contains(req.Header.Get("Content-Type"), "application/grpc") ||
			websocket.IsWebSocketUpgrade(req) {
			grpc_web_srv.ServeHTTP(resp, req)
		} else {
			//No gRPC request
			folderReader(http.FileServer(http.Dir("/home/pi/P4wnP1_go"))).ServeHTTP(resp, req)
		}
	}
	*/

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

func StartRpcServerAndWeb(host string, gRPCPort string, webPort string, absWebRoot string) () {
	//ToDo: Return servers/TCP listener to allow closing from caller
	listen_address_grpc := host + ":" + gRPCPort
	listen_address_web := host + ":" + webPort


	//Create gRPC Server
	s := grpc.NewServer()
	pb.RegisterP4WNP1Server(s, &server{})



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
