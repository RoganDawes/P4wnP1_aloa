// +build js

package main

import (
	pb "github.com/mame82/P4wnP1_go/proto/gopherjs"
	"github.com/gopherjs/gopherjs/js"
	"errors"
	"sync"
	"context"
	"github.com/mame82/P4wnP1_go/common_web"
	"strconv"
	"github.com/mame82/hvue"
	"io"
	"time"
)

var eNoLogEvent = errors.New("No log event")
var eNoHidEvent = errors.New("No HID event")

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
	//println("result",res)

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

//HID event
type jsHidEvent struct {
	*js.Object
	EvType int64 `js:"evtype"`
	VMId   int64  `js:"vmId"`
	JobId   int64  `js:"jobId"`
	HasError  bool  `js:"hasError"`
	Result string `js:"result"`
	Error string `js:"error"`
	Message string `js:"message"`
	EvLogTime string `js:"time"`
}

func (jsEv *jsEvent) toLogEvent() (res *jsLogEvent, err error) {
	if jsEv.Type != common_web.EVT_LOG || len(jsEv.Values) != 4 { return nil,eNoLogEvent}
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

func (jsEv *jsEvent) toHidEvent() (res *jsHidEvent, err error) {
	if jsEv.Type != common_web.EVT_HID || len(jsEv.Values) != 8 { return nil,eNoHidEvent}
	res = &jsHidEvent{Object:O()}

	var ok bool
	res.EvType,ok = jsEv.Values[0].(int64)
	if !ok { return nil,eNoHidEvent }

	res.VMId,ok = jsEv.Values[1].(int64)
	if !ok { return nil,eNoHidEvent}

	res.JobId,ok = jsEv.Values[2].(int64)
	if !ok { return nil,eNoHidEvent}

	res.HasError,ok = jsEv.Values[3].(bool)
	if !ok { return nil,eNoHidEvent}

	res.Result,ok = jsEv.Values[4].(string)
	if !ok { return nil,eNoHidEvent}

	res.Error,ok = jsEv.Values[5].(string)
	if !ok { return nil,eNoHidEvent}

	res.Message,ok = jsEv.Values[6].(string)
	if !ok { return nil,eNoHidEvent}

	res.EvLogTime,ok = jsEv.Values[7].(string)
	if !ok { return nil,eNoHidEvent}


	return res,nil
}


/* HIDJobList */
type jsHidJobState struct {
	*js.Object
	Id             int64  `js:"id"`
	VmId           int64  `js:"vmId"`
	HasFailed      bool   `js:"hasFailed"`
	HasSucceeded   bool   `js:"hasSucceeded"`
	LastMessage    string `js:"lastMessage"`
	TextResult     string `js:"textResult"`
//	TextError      string `js:"textError"`
	LastUpdateTime string `js:"lastUpdateTime"` //JSON timestamp from server
	ScriptSource   string `js:"textSource"`
}

type jsHidJobStateList struct {
	*js.Object
	Jobs *js.Object `js:"jobs"`
}

func NewHIDJobStateList() *jsHidJobStateList {
	jl := &jsHidJobStateList{Object:O()}
	jl.Jobs = O()

	/*
	//ToDo: Delete added a test jobs
	jl.UpdateEntry(99,1,false,false, "This is the latest event message", "current result", "current error","16:00", "type('hello world')")
	jl.UpdateEntry(100,1,false,true, "SUCCESS", "current result", "current error","16:00", "type('hello world')")
	jl.UpdateEntry(101,1,true,false, "FAIL", "current result", "current error","16:00", "type('hello world')")
	jl.UpdateEntry(102,1,true,true, "Error and Success at same time --> UNKNOWN", "current result", "current error","16:00", "type('hello world')")
	jl.UpdateEntry(102,1,true,true, "Error and Success at same time --> UNKNOWN, repeated ID", "current result", "current error","16:00", "type('hello world')")
	*/

	return jl
}

// The Method updates data of the HidJobStateList and is called from two source
// 1) ... from an actively triggered RPC call to "HIDGetRunningJobState", in order to prefill the list with state data of
// running jobs (job id, id of executing vm and job script source). The RPC call only returns information for running jobs,
// not for finished ones (no matter if they failed or succeeded), as the RPC server doesn't keep state for finished jobs
// in order to save memory.
// 2) ... from HID events received via streaming RPC call to EventListen. The events provide additional information on
// results of finished jobs (hasSucceeded, hasError ...). This information is used to update the job state of already
// running jobs. Finished jobs, updated based on such events are kept in the web client's global state, even if the RPC
// server already deleted the job after firing the according events.
//
// Thanks to this two information sources, the client has an overall job state, which is built from RPC server state and
// HID job events (big picture). This mechanism isn't too robust: f.e. if running jobs are received and, let's say a new
// job with id "66" is started before the event receiver is listening, the client will not know that job "66" exists.
// There are a bunch of such race conditions, but it is considered sufficient to fetch the state of all running jobs once
// on client startup with "HIDGetRunningJobState" and update the via events from "EventListen", later on, for the web
// client's muse case.
//
// On Go's end, the job state list is a map, which uses a string representation of the respective job ID as key.
// On JS end, this map has to be represented by a *js.Object, which get's keys added at runtime (every time a new element
// is added to the map). As the job state list is used to present job state data to Vue.JS/Vuex a new problem arises:
// Vue.JS's change detection doesn't work when properties are added or removed from JS objects
// (see https://vuejs.org/v2/guide/list.html#Object-Change-Detection-Caveats). In order to deal with that, the
// "UpdateEntry" method uses the approach proposed by Vue.Js and doesn't update the js.Object, which is backing the job
// state list, directly, but instead uses the "Vue.set()" method to update the object, while making vue aware of it.
// This means: THE "UpdateEntry" METHOD RELIES ON THE PRESENCE OF THE "Vue" OBJECT IN JAVASCRIPT GLOBAL SCOPE. This again
// means Vue.JS has to be loaded, BEFORE THIS METHOD IS CALLED"
func (jl *jsHidJobStateList) UpdateEntry(id, vmId int64, hasFailed, hasSucceeded bool, message, textResult, lastUpdateTime, scriptSource string) {
	key := strconv.Itoa(int(id))

	//Check if job exists, update existing one if already present
	var j *jsHidJobState
	if res := jl.Jobs.Get(key);res == js.Undefined {
		j = &jsHidJobState{Object:O()}
	} else {
		j = &jsHidJobState{Object:res}
	}

	//Create job object

	j.Id = id
	j.VmId = vmId
	j.HasFailed = hasFailed
	j.HasSucceeded = hasSucceeded
	j.LastMessage = message
	j.TextResult = textResult
	j.LastUpdateTime = lastUpdateTime
	if len(scriptSource) > 0 {j.ScriptSource = scriptSource}
	//jl.Jobs.Set(strconv.Itoa(int(j.Id)), j) //jobs["j.ID"]=j <--Property addition/update can't be detected by Vue.js, see https://vuejs.org/v2/guide/list.html#Object-Change-Detection-Caveats
	hvue.Set(jl.Jobs, key, j)
}

func (jl *jsHidJobStateList) DeleteEntry(id int64) {
	jl.Jobs.Delete(strconv.Itoa(int(id))) //JS version
	//delete(jl.Jobs, strconv.Itoa(int(id)))
}

/* EVENT LOGGER */
type jsEventReceiver struct {
	*js.Object
	LogArray      *js.Object `js:"logArray"`
	HidEventArray *js.Object `js:"eventHidArray"`
	cancel        context.CancelFunc
	*sync.Mutex
	MaxEntries    int        `js:"maxEntries"`
	JobList       *jsHidJobStateList `js:"jobList"` //Needs to be exposed to JS in order to use JobList.UpdateEntry() from this JS object
}

func NewEventReceiver(maxEntries int, jobList *jsHidJobStateList) *jsEventReceiver {
	eventReceiver := &jsEventReceiver{
		Object: js.Global.Get("Object").New(),
	}

	eventReceiver.Mutex = &sync.Mutex{}
	eventReceiver.LogArray = js.Global.Get("Array").New()
	eventReceiver.HidEventArray = js.Global.Get("Array").New()
	eventReceiver.MaxEntries = maxEntries
	eventReceiver.JobList = jobList

	return eventReceiver
}

func (data *jsEventReceiver) handleHidEvent(hEv *jsHidEvent ) {
	println("Received HID EVENT", hEv)
	switch hEv.EvType {
	case common_web.HidEventType_JOB_STARTED:
		// Note: the JOB_STARTED event carries the script source in the message field, (no need to re-request the job
		// state in order to retrieve the source code of the job, when adding it to the job state list)
		data.JobList.UpdateEntry(hEv.JobId, hEv.VMId, hEv.HasError, false, "Script started", "", hEv.EvLogTime,hEv.Message)
	case common_web.HidEventType_JOB_FAILED:
		data.JobList.UpdateEntry(hEv.JobId, hEv.VMId, hEv.HasError, false, hEv.Message, hEv.Error, hEv.EvLogTime,"")
	case common_web.HidEventType_JOB_SUCCEEDED:
		data.JobList.UpdateEntry(hEv.JobId, hEv.VMId, hEv.HasError, true, hEv.Message, hEv.Result, hEv.EvLogTime,"")
	case common_web.HidEventType_JOB_CANCELLED:
		data.JobList.UpdateEntry(hEv.JobId, hEv.VMId, true, false, hEv.Message, hEv.Message, hEv.EvLogTime,"")
	default:
		println("unhandled hid event " + common_web.EventType_name[hEv.EvType], hEv)
	}

}


/* This method gets internalized and therefor the mutex won't be accessible*/
func (data *jsEventReceiver) HandleEvent(ev *pb.Event ) {
	//	println("ADD ENTRY", ev)
	go func() {
		/*
				data.Lock()
				defer data.Unlock()
		*/


		jsEv := NewJsEventFromNative(ev)
		switch jsEv.Type {
		//if LOG event add to logArray
		case common_web.EVT_LOG:
			if logEv,err := jsEv.toLogEvent(); err == nil {
				data.LogArray.Call("push", logEv)
			} else {
				println("couldn't convert to LogEvent: ", jsEv)
			}
			//if HID event add to eventHidArray
		case common_web.EVT_HID:
			if hidEv,err := jsEv.toHidEvent(); err == nil {
				data.HidEventArray.Call("push", hidEv)

				//handle event
				data.handleHidEvent(hidEv)
			} else {
				println("couldn't convert to HidEvent: ", jsEv)
			}
		}

		//reduce to length (note: kebab case 'max-entries' is translated to camel case 'maxEntries' by vue)
		for data.LogArray.Length() > data.MaxEntries {
			data.LogArray.Call("shift") // remove first element
		}
		for data.HidEventArray.Length() > data.MaxEntries {
			data.HidEventArray.Call("shift") // remove first element
		}

	}()


}


func (data *jsEventReceiver) StartListening() {

	println("Start listening called", data)


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
			for ConnectionTest(time.Second * 3) != nil {
				println("... failed, retry for 3 seconds")
				globalState.FailedConnectionAttempts++
			}
			println("... success")
			globalState.IsConnected = true
			globalState.FailedConnectionAttempts = 0

			ctx,cancel := context.WithCancel(context.Background())
			data.cancel = cancel

			// try RPC call
			evStream, err := Client.Client.EventListen(ctx, &pb.EventRequest{ListenType: common_web.EVT_ANY}) //No error if Websocket connection fails
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
					data.HandleEvent(event)
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


func (data *jsEventReceiver) StopListening() {
	data.cancel()
}

