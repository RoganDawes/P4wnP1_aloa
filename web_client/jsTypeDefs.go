package main

import (
	pb "../proto/gopherjs"
	"github.com/gopherjs/gopherjs/js"
	"errors"
	"../common"
	"sync"
	"context"
	"io"
	"fmt"
)

var eNoLogEvent = errors.New("No log event")

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

type jsEvent struct {
	*js.Object
	Type   int64 `js:"type"`
	Values []interface{}
	JSValues *js.Object `js:"values"`
}


func NewJsEventFromNative(event *pb.Event) (res *jsEvent) {
	res = &jsEvent{Object:O()}
	res.JSValues = js.Global.Get("Array").New()
	res.Type = event.Type
	res.Values = make([]interface{}, len(event.Values))
	for idx,val := range event.Values {
		switch valT := val.Val.(type) {
		case *pb.EventValue_Tint64:
			res.Values[idx] = valT.Tint64
			res.JSValues.Call("push", valT.Tint64)
		case *pb.EventValue_Tstring:
			res.Values[idx] = valT.Tstring
			res.JSValues.Call("push", valT.Tstring)
		case *pb.EventValue_Tbool:
			res.Values[idx] = valT.Tbool
			res.JSValues.Call("push", valT.Tbool)
		default:
			println("error parsing event value", valT)
		}
	}
	println("result",res)

	return res
}

//Log event
type jsLogEvent struct {
	*js.Object
	EvLogSource  string `js:"source"`
	EvLogLevel   int  `js:"level"`
	EvLogMessage string `js:"message"`
	EvLogTime string `js:"time"`
}

func (jsEv *jsEvent) toLogEvent() (res *jsLogEvent, err error) {
	if jsEv.Type != common.EVT_LOG || len(jsEv.Values) != 4 { return nil,eNoLogEvent}
	res = &jsLogEvent{Object:O()}

	var ok bool
	res.EvLogSource,ok = jsEv.Values[0].(string)
	if !ok { return nil,eNoLogEvent }

	ll,ok := jsEv.Values[1].(int64)
	if !ok { return nil,eNoLogEvent}
	res.EvLogLevel = int(ll)

	res.EvLogMessage,ok = jsEv.Values[2].(string)
	if !ok { return nil,eNoLogEvent}

	res.EvLogTime,ok = jsEv.Values[3].(string)
	if !ok { return nil,eNoLogEvent}

	return res,nil
}

/*
func DeconstructEventLog(gRPCEv *pb.Event) (res *jsLogEvent, err error) {
	if gRPCEv.Type != common.EVT_LOG { return nil,errors.New("No log event")}

	res = &jsLogEvent{Object:O()}
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
*/

/* EVENT LOGGER */

type jsLoggerData struct {
	*js.Object
	LogArray *js.Object `js:"logArray"`
	EventArray *js.Object `js:"eventArray"`
	cancel context.CancelFunc
	*sync.Mutex
	MaxEntries int `js:"maxEntries"`
}

func NewLogger(maxEntries int) *jsLoggerData {
	loggerVmData := &jsLoggerData{
		Object: js.Global.Get("Object").New(),
	}

	loggerVmData.Mutex = &sync.Mutex{}
	loggerVmData.LogArray = js.Global.Get("Array").New()
	loggerVmData.EventArray = js.Global.Get("Array").New()
	loggerVmData.MaxEntries = maxEntries

	return loggerVmData
}

/* This method gets internalized and therefor the mutex won't be accessible*/
func (data *jsLoggerData) AddEntry(ev *pb.Event ) {
//	println("ADD ENTRY", ev)
	go func() {
/*
		data.Lock()
		defer data.Unlock()
*/

		fmt.Println("LOOOOOG ENTRYYYYYYYYYYYYYYYYY")

		//if LOG event add to logArray
		jsEv := NewJsEventFromNative(ev)
		println("JS from native", jsEv)
		if jsEv.Type == common.EVT_LOG {
			if logEv,err := jsEv.toLogEvent(); err == nil {
				data.LogArray.Call("push", logEv)
			} else {
				println("couldn't convert to LogEvent: ", jsEv)
			}
		} else {
			data.EventArray.Call("push", jsEv)
		}


		/*
		logEv, err := DeconstructEventLog(ev)
		if err != nil {
			println("Logger: Error adding log entry, provided event couldn't be converted to log event")
			return
		}

		data.LogArray.Call("push", logEv)
		*/

		//reduce to length (note: kebab case 'max-entries' is translated to camel case 'maxEntries' by vue)
		for data.LogArray.Length() > data.MaxEntries {
			data.LogArray.Call("shift") // remove first element
		}
		for data.EventArray.Length() > data.MaxEntries {
			data.EventArray.Call("shift") // remove first element
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
		println("EVENTLISTENING ENTERING LOOP")
		for {
			event, err := evStream.Recv()
			if err == io.EOF { break }
			if err != nil { return }

			//println("Event: ", event)
			data.AddEntry(event)
			println(event)
		}
		println("EVENTLISTENING ABORTED")
		return
	}()
}


func (data *jsLoggerData) StopListening() {
	data.cancel()
}

