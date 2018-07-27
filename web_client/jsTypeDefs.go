package main

import (
	pb "../proto/gopherjs"
	"github.com/gopherjs/gopherjs/js"
	"errors"
	"../common"
	"sync"
	"context"
	"io"
)

/* USB Gadget types corresponding to gRPC messages */

type jsGadgetSettings struct {
	*js.Object
	Enabled          bool  `js:"Enabled"`
	Vid              string  `js:"Vid"`
	Pid              string  `js:"Pid"`
	Manufacturer     string `js:"Manufacturer"`
	Product          string `js:"Product"`
	Serial           string `js:"Serial"`
	Use_CDC_ECM      bool `js:"Use_CDC_ECM"`
	Use_RNDIS        bool `js:"Use_RNDIS"`
	Use_HID_KEYBOARD bool `js:"Use_HID_KEYBOARD"`
	Use_HID_MOUSE    bool `js:"Use_HID_MOUSE"`
	Use_HID_RAW      bool `js:"Use_HID_RAW"`
	Use_UMS          bool `js:"Use_UMS"`
	Use_SERIAL       bool `js:"Use_SERIAL"`
	RndisSettings    *VGadgetSettingsEthernet `js:"RndisSettings"`
	CdcEcmSettings   *VGadgetSettingsEthernet `js:"CdcEcmSettings"`
	UmsSettings      *VGadgetSettingsUMS `js:"UmsSettings"`
}

type VGadgetSettingsEthernet struct {
	*js.Object
	HostAddr string `js:"HostAddr"`
	DevAddr  string `js:"DevAddr"`
}


type VGadgetSettingsUMS struct {
	*js.Object
	Cdrom bool `js:"Cdrom"`
	File  string `js:"File"`
}

func (jsGS jsGadgetSettings) toGS() (gs *pb.GadgetSettings) {
	return &pb.GadgetSettings{
		Serial:           jsGS.Serial,
		Use_SERIAL:       jsGS.Use_SERIAL,
		Use_UMS:          jsGS.Use_UMS,
		Use_HID_RAW:      jsGS.Use_HID_RAW,
		Use_HID_MOUSE:    jsGS.Use_HID_MOUSE,
		Use_HID_KEYBOARD: jsGS.Use_HID_KEYBOARD,
		Use_RNDIS:        jsGS.Use_RNDIS,
		Use_CDC_ECM:      jsGS.Use_CDC_ECM,
		Product:          jsGS.Product,
		Manufacturer:     jsGS.Manufacturer,
		Vid:              jsGS.Vid,
		Pid:              jsGS.Pid,
		Enabled:          jsGS.Enabled,
		UmsSettings: &pb.GadgetSettingsUMS{
			Cdrom: jsGS.UmsSettings.Cdrom,
			File:  jsGS.UmsSettings.File,
		},
		CdcEcmSettings: &pb.GadgetSettingsEthernet{
			DevAddr:  jsGS.CdcEcmSettings.DevAddr,
			HostAddr: jsGS.CdcEcmSettings.HostAddr,
		},
		RndisSettings: &pb.GadgetSettingsEthernet{
			DevAddr:  jsGS.RndisSettings.DevAddr,
			HostAddr: jsGS.RndisSettings.HostAddr,
		},
	}
}

func (jsGS *jsGadgetSettings) fromGS(gs *pb.GadgetSettings) {
	println(gs)

	jsGS.Enabled = gs.Enabled
	jsGS.Vid = gs.Vid
	jsGS.Pid = gs.Pid
	jsGS.Manufacturer = gs.Manufacturer
	jsGS.Product = gs.Product
	jsGS.Serial = gs.Serial
	jsGS.Use_CDC_ECM = gs.Use_CDC_ECM
	jsGS.Use_RNDIS = gs.Use_RNDIS
	jsGS.Use_HID_KEYBOARD = gs.Use_HID_KEYBOARD
	jsGS.Use_HID_MOUSE = gs.Use_HID_MOUSE
	jsGS.Use_HID_RAW = gs.Use_HID_RAW
	jsGS.Use_UMS = gs.Use_UMS
	jsGS.Use_SERIAL = gs.Use_SERIAL

	jsGS.RndisSettings = &VGadgetSettingsEthernet{
		Object: O(),
	}
	if gs.RndisSettings != nil {
		jsGS.RndisSettings.HostAddr = gs.RndisSettings.HostAddr
		jsGS.RndisSettings.DevAddr = gs.RndisSettings.DevAddr
	}

	jsGS.CdcEcmSettings = &VGadgetSettingsEthernet{
		Object: O(),
	}
	if gs.CdcEcmSettings != nil {
		jsGS.CdcEcmSettings.HostAddr = gs.CdcEcmSettings.HostAddr
		jsGS.CdcEcmSettings.DevAddr = gs.CdcEcmSettings.DevAddr
	}

	jsGS.UmsSettings = &VGadgetSettingsUMS{
		Object: O(),
	}
	if gs.UmsSettings != nil {
		jsGS.UmsSettings.File = gs.UmsSettings.File
		jsGS.UmsSettings.Cdrom = gs.UmsSettings.Cdrom
	}
}


func NewUSBGadgetSettings() *jsGadgetSettings {
	gs := &jsGadgetSettings{
		Object: O(),
	}
	gs.fromGS(&pb.GadgetSettings{}) //start with empty settings, but create nested structs

	return gs
}

/** Events **/

//Log event
type jsEventLog struct {
	*js.Object
	EvLogSource  string `js:"source"`
	EvLogLevel   int  `js:"level"`
	EvLogMessage string `js:"message"`
	EvLogTime string `js:"time"`
}

func DeconstructEventLog(gRPCEv *pb.Event) (res *jsEventLog, err error) {
	if gRPCEv.Type != common.EVT_LOG { return nil,errors.New("No log event")}

	res = &jsEventLog{Object:O()}
	switch vT := gRPCEv.Values[0].Val.(type) {
	case *pb.EventValue_Tstring:
		res.EvLogSource = vT.Tstring
	default:
		return nil, errors.New("Value at position 0 has wrong type for a log event")
	}
	switch vT := gRPCEv.Values[1].Val.(type) {
	case *pb.EventValue_Tint64:
		res.EvLogLevel = int(vT.Tint64)
	default:
		return nil, errors.New("Value at position 1 has wrong type for a log event")
	}
	switch vT := gRPCEv.Values[2].Val.(type) {
	case *pb.EventValue_Tstring:
		res.EvLogMessage = vT.Tstring
	default:
		return nil, errors.New("Value at position 2 has wrong type for a log event")
	}
	switch vT := gRPCEv.Values[3].Val.(type) {
	case *pb.EventValue_Tstring:
		res.EvLogTime = vT.Tstring
	default:
		return nil, errors.New("Value at position 3 has wrong type for a log event")
	}

	return res, nil
}


/* EVENT LOGGER */

type jsLoggerData struct {
	*js.Object
	LogArray *js.Object `js:"logArray"`
	cancel context.CancelFunc
	*sync.Mutex
	MaxEntries int `js:"maxEntries"`
}


/* This method gets internalized and therefor the mutex won't be accessible*/
func (data *jsLoggerData) AddEntry(ev *pb.Event ) {
//	println("ADD ENTRY", ev)
	go func() {
/*
		data.Lock()
		defer data.Unlock()
*/

		logEv, err := DeconstructEventLog(ev)
		if err != nil {
			println("Logger: Error adding log entry, provided event couldn't be converted to log event")
			return
		}

		data.LogArray.Call("push", logEv)

		//reduce to length (note: kebab case 'max-entries' is translated to camel case 'maxEntries' by vue)
		for data.LogArray.Length() > data.MaxEntries {
			data.LogArray.Call("shift") // remove first element
		}

	}()


}

func (data *jsLoggerData) StartListening() {

	println("Start listening called", data)
	ctx,cancel := context.WithCancel(context.Background())
	data.cancel = cancel

	evStream, err := Client.Client.EventListen(ctx, &pb.EventRequest{ListenType: common.EVT_LOG})
	if err != nil {
		cancel()
		println("Error listening fo Log events", err)
		return
	}

	go func() {
		defer cancel()
		for {
			event, err := evStream.Recv()
			if err == io.EOF { break }
			if err != nil { return }

			//println("Event: ", event)
			data.AddEntry(event)

		}
		return
	}()
}


func (data *jsLoggerData) StopListening() {
	data.cancel()
}

func NewLogger(maxEntries int) *jsLoggerData {
	loggerVmData := &jsLoggerData{
		Object: js.Global.Get("Object").New(),
	}

	loggerVmData.Mutex = &sync.Mutex{}
	loggerVmData.LogArray = js.Global.Get("Array").New()
	loggerVmData.MaxEntries = maxEntries

	return loggerVmData
}