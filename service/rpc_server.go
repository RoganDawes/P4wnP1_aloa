//package rpcserv
package service

import (
	"log"
	pb "../proto"
	"golang.org/x/net/context"
	"net"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/improbable-eng/grpc-web/go/grpcweb"

	"net/http"
	"strings"
	"path"
	"time"
)

type server struct {}

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

	errg := DeployGadgetSettings(GadgetSettingsState)
	err = nil
	if errg != nil {
		err = errors.New(fmt.Sprintf("Deploying new gadget settings failed, reverted to old ones: %v", errg))
		DeployGadgetSettings(*gs_backup) //We don't catch the error, as the old settings should have been working
	}

	gs, _ = ParseGadgetState(USB_GADGET_NAME) //Return settings from deployed gadget
	return
}

func (s *server) GetGadgetSettings(context.Context, *pb.Empty) (*pb.GadgetSettings, error) {
	return &GadgetSettingsState, nil
}

func (s *server) SetGadgetSettings(ctx context.Context, gs *pb.GadgetSettings) (res *pb.GadgetSettings, err error) {
	if err = ValidateGadgetSetting(*gs); err != nil {
		//We return the validation error and the current (unchanged) GadgetSettingsState
		res = &GadgetSettingsState
		return
	}
	GadgetSettingsState = *gs
	res = &GadgetSettingsState
	return
}

func (s *server) GetLEDSettings(context.Context, *pb.Empty) (res *pb.LEDSettings, err error) {
	res, err = GetLed()
	log.Printf("GetLEDSettings, result: %+v", res)
	return
}

func (s *server) SetLEDSettings(ctx context.Context, ls *pb.LEDSettings) (*pb.Empty, error) {
	log.Printf("SetLEDSettings %+v", ls)
	SetLed(*ls)
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
	// Register reflection service on gRPC server.
	reflection.Register(s)
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

func StartRpcServerAndWeb(host string, port string) {
	listen_address := host + ":" + port
	webserver_path := "/home/pi/P4wnP1_go/www" //ToDo: Change this to an absolute path which could be used after installation

	//Create gRPC Server
	s := grpc.NewServer()
	pb.RegisterP4WNP1Server(s, &server{})

	//Wrap the server into a gRPC-web server
	grpc_web_srv := grpcweb.WrapServer(s) //Wrap server to improbable grpc-web with websockets
	//define a handler for a HTTP web server using the gRPC-web proxy
	http_gRPC_web_handler := func(resp http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.Header.Get("Content-Type"), "application/grpc") || req.Method == "OPTIONS" {
			fmt.Printf("gRPC-web req:\n %v\n", req)
			grpc_web_srv.ServeHTTP(resp, req) // if content type indicates grpc or REQUEST METHOD IS OPTIONS (pre-flight) serve gRPC-web
		} else {
			fmt.Printf("legacy web req:\n %v\n", req)
			http.FileServer(http.Dir((webserver_path))).ServeHTTP(resp, req)
		}
	}

	//Open TCP listener
	log.Printf("P4wnP1 gRPC server listening on " + listen_address)
	lis, err := net.Listen("tcp", listen_address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// run gRPC server in go routine
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	//Setup our HTTP server
	http_srv := &http.Server{
		Addr: host + ":80", //listen on port 80 with webservice
		Handler: http.HandlerFunc(http_gRPC_web_handler),
		ReadHeaderTimeout: 5*time.Second,
		IdleTimeout: 120*time.Second,
	}
	log.Printf("P4wnP1 gRPC-web server listening on " + http_srv.Addr)
	err_http := http_srv.ListenAndServe()
	if err_http != nil {
		log.Fatal(err)
	}

}
