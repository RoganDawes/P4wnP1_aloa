// +build js

package main

import (
	"errors"
	"github.com/gopherjs/gopherjs/js"
	"github.com/mame82/P4wnP1_go/common_web"
	pb "github.com/mame82/P4wnP1_go/proto/gopherjs"
	"github.com/mame82/hvue"
	"strconv"
)

var eNoLogEvent = errors.New("No log event")
var eNoHidEvent = errors.New("No HID event")


type jsMasterTemplate struct {
	*js.Object
	TemplateNameBluetooth      string   `js:"TemplateNameBluetooth"`
	TemplateNameUSB            string   `js:"TemplateNameUSB"`
	TemplateNameWiFi           string   `js:"TemplateNameWiFi"`
	TemplateNameTriggerActions string   `js:"TemplateNameTriggerActions"`
	TemplateNamesNetwork       []string `js:"TemplateNamesNetwork"`
}

func (target *jsMasterTemplate) fromGo(src *pb.MasterTemplate) {
	target.TemplateNamesNetwork = src.TemplateNamesNetwork
	target.TemplateNameBluetooth = src.TemplateNameBluetooth
	target.TemplateNameWiFi = src.TemplateNameWifi
	target.TemplateNameUSB = src.TemplateNameUsb
	target.TemplateNameTriggerActions = src.TemplateNameTriggerActions
}

func (src *jsMasterTemplate) toGo() (target *pb.MasterTemplate) {
	target = &pb.MasterTemplate{}
	target.TemplateNamesNetwork = src.TemplateNamesNetwork
	target.TemplateNameBluetooth = src.TemplateNameBluetooth
	target.TemplateNameWifi = src.TemplateNameWiFi
	target.TemplateNameUsb = src.TemplateNameUSB
	target.TemplateNameTriggerActions = src.TemplateNameTriggerActions
	return target
}

func NewMasterTemplate() (res *jsMasterTemplate) {
	res = &jsMasterTemplate{Object: O()}
	res.TemplateNameBluetooth = ""
	res.TemplateNameWiFi = ""
	res.TemplateNameUSB = ""
	res.TemplateNameTriggerActions = ""
	res.TemplateNamesNetwork = []string{}

	return res
}

type jsRequestMasterTemplateStorage struct {
	*js.Object
	TemplateName string `js:"TemplateName"`
	Template     *jsMasterTemplate `js:"Template"`
}

func (rs *jsRequestMasterTemplateStorage) toGo() *pb.RequestMasterTemplateStorage {
	return &pb.RequestMasterTemplateStorage{
		TemplateName: rs.TemplateName,
		Template: rs.Template.toGo(),
	}
}

func (rs *jsRequestMasterTemplateStorage) fromGo(src *pb.RequestMasterTemplateStorage) {
	rs.TemplateName = src.TemplateName
	rs.Template = NewMasterTemplate()
	rs.Template.fromGo(src.Template)
}

func NewRequestMasterTemplateStorage() (res *jsRequestMasterTemplateStorage) {
	res = &jsRequestMasterTemplateStorage{Object:O()}
	res.Template = NewMasterTemplate()
	res.TemplateName = ""
	return res
}

type jsLoadHidScriptSourceMode int
const (
	HID_SCRIPT_SOURCE_LOAD_MODE_PREPEND jsLoadHidScriptSourceMode = iota
	HID_SCRIPT_SOURCE_LOAD_MODE_APPEND
	HID_SCRIPT_SOURCE_LOAD_MODE_REPLACE
)
type jsLoadHidScriptSourceReq struct {
	*js.Object
	FileName string `js:"FileName"`
	Mode jsLoadHidScriptSourceMode `js:"Mode"`
}

type jsWifiRequestSettingsStorage struct {
	*js.Object
	TemplateName string `js:"TemplateName"`
	Settings     *jsWiFiSettings `js:"Settings"`
}

func (rs *jsWifiRequestSettingsStorage) toGo() *pb.WifiRequestSettingsStorage {
	return &pb.WifiRequestSettingsStorage{
		Settings: rs.Settings.toGo(),
		TemplateName: rs.TemplateName,
	}
}

func (rs *jsWifiRequestSettingsStorage) fromGo(src *pb.WifiRequestSettingsStorage) {
	rs.TemplateName = src.TemplateName
	rs.Settings = NewWifiSettings()
	rs.Settings.fromGo(src.Settings)
}

func NewWifiRequestSettingsStorage() *jsWifiRequestSettingsStorage {
	res := &jsWifiRequestSettingsStorage{Object:O()}
	res.TemplateName = ""
	res.Settings = NewWifiSettings()
	return res
}

/* USB Gadget types corresponding to gRPC messages */

type jsUSBRequestSettingsStorage struct {
	*js.Object
	TemplateName string `js:"TemplateName"`
	Settings     *jsGadgetSettings `js:"Settings"`
}

func (rs *jsUSBRequestSettingsStorage) toGo() *pb.USBRequestSettingsStorage {
	return &pb.USBRequestSettingsStorage{
		Settings: rs.Settings.toGo(),
		TemplateName: rs.TemplateName,
	}
}

func (rs *jsUSBRequestSettingsStorage) fromGo(src *pb.USBRequestSettingsStorage) {
	rs.TemplateName = src.TemplateName
	rs.Settings = NewUSBGadgetSettings()
	rs.Settings.fromGo(src.Settings)
}

func NewUSBRequestSettingsStorage() *jsUSBRequestSettingsStorage {
	res := &jsUSBRequestSettingsStorage{Object:O()}
	res.TemplateName = ""
	res.Settings = NewUSBGadgetSettings()
	return res
}


type jsGadgetSettings struct {
	*js.Object
	Enabled          bool                     `js:"Enabled"`
	Vid              string                   `js:"Vid"`
	Pid              string                   `js:"Pid"`
	Manufacturer     string                   `js:"Manufacturer"`
	Product          string                   `js:"Product"`
	Serial           string                    `js:"Serial"`
	Use_CDC_ECM      bool                      `js:"Use_CDC_ECM"`
	Use_RNDIS        bool                      `js:"Use_RNDIS"`
	Use_HID_KEYBOARD bool                      `js:"Use_HID_KEYBOARD"`
	Use_HID_MOUSE    bool                      `js:"Use_HID_MOUSE"`
	Use_HID_RAW      bool                      `js:"Use_HID_RAW"`
	Use_UMS          bool                      `js:"Use_UMS"`
	Use_SERIAL       bool                      `js:"Use_SERIAL"`
	RndisSettings    *jsGadgetSettingsEthernet `js:"RndisSettings"`
	CdcEcmSettings   *jsGadgetSettingsEthernet `js:"CdcEcmSettings"`
	UmsSettings      *jsGadgetSettingsUMS      `js:"UmsSettings"`
}

type jsGadgetSettingsEthernet struct {
	*js.Object
	HostAddr string `js:"HostAddr"`
	DevAddr  string `js:"DevAddr"`
}

type jsGadgetSettingsUMS struct {
	*js.Object
	Cdrom bool   `js:"Cdrom"`
	File  string `js:"File"`
}

func (jsGS jsGadgetSettings) toGo() (gs *pb.GadgetSettings) {
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

func (jsGS *jsGadgetSettings) fromGo(gs *pb.GadgetSettings) {
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

	jsGS.RndisSettings = &jsGadgetSettingsEthernet{
		Object: O(),
	}
	if gs.RndisSettings != nil {
		jsGS.RndisSettings.HostAddr = gs.RndisSettings.HostAddr
		jsGS.RndisSettings.DevAddr = gs.RndisSettings.DevAddr
	}

	jsGS.CdcEcmSettings = &jsGadgetSettingsEthernet{
		Object: O(),
	}
	if gs.CdcEcmSettings != nil {
		jsGS.CdcEcmSettings.HostAddr = gs.CdcEcmSettings.HostAddr
		jsGS.CdcEcmSettings.DevAddr = gs.CdcEcmSettings.DevAddr
	}

	jsGS.UmsSettings = &jsGadgetSettingsUMS{
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
	gs.fromGo(&pb.GadgetSettings{}) //start with empty settings, but create nested structs

	return gs
}

/** Events **/

type jsEvent struct {
	*js.Object
	Type     int64 `js:"type"`
	Values   []interface{}
	JSValues *js.Object `js:"values"`
}

func NewJsEventFromNative(event *pb.Event) (res *jsEvent) {
	res = &jsEvent{Object: O()}
	res.JSValues = js.Global.Get("Array").New()
	res.Type = event.Type
	res.Values = make([]interface{}, len(event.Values))
	for idx, val := range event.Values {
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
	EvLogLevel   int    `js:"level"`
	EvLogMessage string `js:"message"`
	EvLogTime    int64 `js:"time"`
}

//HID event
type jsHidEvent struct {
	*js.Object
	EvType    int64  `js:"evtype"`
	VMId      int64  `js:"vmId"`
	JobId     int64  `js:"jobId"`
	HasError  bool   `js:"hasError"`
	Result    string `js:"result"`
	Error     string `js:"error"`
	Message   string `js:"message"`
	EvLogTime int64 `js:"time"`
}

func (jsEv *jsEvent) toLogEvent() (res *jsLogEvent, err error) {
	if jsEv.Type != common_web.EVT_LOG || len(jsEv.Values) != 4 {
		return nil, eNoLogEvent
	}
	res = &jsLogEvent{Object: O()}

	var ok bool
	res.EvLogSource, ok = jsEv.Values[0].(string)
	if !ok {
		return nil, eNoLogEvent
	}

	ll, ok := jsEv.Values[1].(int64)
	if !ok {
		return nil, eNoLogEvent
	}
	res.EvLogLevel = int(ll)

	res.EvLogMessage, ok = jsEv.Values[2].(string)
	if !ok {
		return nil, eNoLogEvent
	}

	res.EvLogTime, ok = jsEv.Values[3].(int64)
	if !ok {
		return nil, eNoLogEvent
	}
	println("EvLogTime", res.EvLogTime)

	return res, nil
}

func (jsEv *jsEvent) toHidEvent() (res *jsHidEvent, err error) {
	if jsEv.Type != common_web.EVT_HID || len(jsEv.Values) != 8 {
		return nil, eNoHidEvent
	}
	res = &jsHidEvent{Object: O()}

	var ok bool
	res.EvType, ok = jsEv.Values[0].(int64)
	if !ok {
		return nil, eNoHidEvent
	}

	res.VMId, ok = jsEv.Values[1].(int64)
	if !ok {
		return nil, eNoHidEvent
	}

	res.JobId, ok = jsEv.Values[2].(int64)
	if !ok {
		return nil, eNoHidEvent
	}

	res.HasError, ok = jsEv.Values[3].(bool)
	if !ok {
		return nil, eNoHidEvent
	}

	res.Result, ok = jsEv.Values[4].(string)
	if !ok {
		return nil, eNoHidEvent
	}

	res.Error, ok = jsEv.Values[5].(string)
	if !ok {
		return nil, eNoHidEvent
	}

	res.Message, ok = jsEv.Values[6].(string)
	if !ok {
		return nil, eNoHidEvent
	}

	res.EvLogTime, ok = jsEv.Values[7].(int64)
	if !ok {
		return nil, eNoHidEvent
	}

	return res, nil
}

/* HIDJobList */
type jsHidJobState struct {
	*js.Object
	Id           int64  `js:"id"`
	VmId         int64  `js:"vmId"`
	HasFailed    bool   `js:"hasFailed"`
	HasSucceeded bool   `js:"hasSucceeded"`
	LastMessage  string `js:"lastMessage"`
	TextResult   string `js:"textResult"`
	//	TextError      string `js:"textError"`
	LastUpdateTime int64 `js:"lastUpdateTime"` //JSON timestamp from server
	ScriptSource   string `js:"textSource"`
}

type jsHidJobStateList struct {
	*js.Object
	Jobs *js.Object `js:"jobs"`
}

func NewHIDJobStateList() *jsHidJobStateList {
	jl := &jsHidJobStateList{Object: O()}
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
func (jl *jsHidJobStateList) UpdateEntry(id, vmId int64, hasFailed, hasSucceeded bool, message string, textResult string, lastUpdateTime int64, scriptSource string) {
	key := strconv.Itoa(int(id))

	//Check if job exists, update existing one if already present
	var j *jsHidJobState
	if res := jl.Jobs.Get(key); res == js.Undefined {
		j = &jsHidJobState{Object: O()}
	} else {
		j = &jsHidJobState{Object: res}
	}

	//Create job object

	j.Id = id
	j.VmId = vmId
	j.HasFailed = hasFailed
	j.HasSucceeded = hasSucceeded
	j.LastMessage = message
	j.TextResult = textResult
	j.LastUpdateTime = lastUpdateTime
	if len(scriptSource) > 0 {
		j.ScriptSource = scriptSource
	}
	//jl.Jobs.Set(strconv.Itoa(int(j.Id)), j) //jobs["j.ID"]=j <--Property addition/update can't be detected by Vue.js, see https://vuejs.org/v2/guide/list.html#Object-Change-Detection-Caveats
	hvue.Set(jl.Jobs, key, j)
}

func (jl *jsHidJobStateList) Clear() {
	hvue.Set(jl,"jobs",O())
}


func (jl *jsHidJobStateList) DeleteEntry(id int64) {
	key := strconv.Itoa(int(id))
	hvue.Delete(jl.Jobs, key) // vue reactive version
	//jl.Jobs.Delete(key) //JS version
	//delete(jl.Jobs, key) // go version
}

/* WiFi settings */

type jsWiFiState struct {
	*js.Object
	Mode     int          `js:"mode"`
	Channel  uint32       `js:"channel"`
	Ssid   string `js:"ssid"`
	CurrentSettings *jsWiFiSettings `js:"CurrentSettings"`
}

func (target *jsWiFiState) fromGo(src *pb.WiFiState) {
	target.Mode = int(src.Mode)
	target.Channel = src.Channel
	target.Ssid = src.Ssid
	target.CurrentSettings = NewWifiSettings()
	target.CurrentSettings.fromGo(src.CurrentSettings)
	return
}

func (src *jsWiFiState) toGo() (target *pb.WiFiState) {
	target = &pb.WiFiState{
		Channel:     src.Channel,
		Mode: pb.WiFiStateMode(src.Mode),
		Ssid: src.Ssid,
		CurrentSettings: src.CurrentSettings.toGo(),
	}

	return
}

func (src jsWiFiState) ModeString() (strMode string) {
	switch src.Mode {
	case int(pb.WiFiStateMode_STA_NOT_CONNECTED):
		return "Not connected"
	case int(pb.WiFiStateMode_AP_UP):
		return "Access Point"
	case int(pb.WiFiStateMode_STA_CONNECTED):
		return "Connected"
	default:
		return "UNKNOWN"

	}
}

func NewWiFiState() *jsWiFiState {
	res := &jsWiFiState{Object: O()}
	res.Channel = 0
	res.Mode = int(pb.WiFiStateMode_STA_NOT_CONNECTED)
	res.Ssid = ""
	res.CurrentSettings = NewWifiSettings()
	return res
}

type jsWiFiBSSCfg struct {
	*js.Object
	SSID string `js:"ssid"`
	PSK  string `js:"psk"`
}

type jsWiFiSettings struct {
	*js.Object
	Name     string `js:"name"`
	Disabled bool   `js:"disabled"`
	Reg      string `js:"reg"`
	Mode     int    `js:"mode"`     //AP, STA, Failover
	AuthMode int    `js:"authMode"` //WPA2_PSK, OPEN
	Channel  int    `js:"channel"`
	HideSsid bool   `js:"hideSsid"`
	Nexmon   bool   `js:"nexmon"`

	Ap_BSS         *jsWiFiBSSCfg `js:"apBss"`
	Client_BSSList *js.Object    `js:"staBssList"` //pointer to js array
}

func (src *jsWiFiSettings) toGo() (target *pb.WiFiSettings) {
	// assure undefined strings end up as empty strings

	target = &pb.WiFiSettings{
		Name:        src.Name,
		Disabled:    src.Disabled,
		Regulatory:  src.Reg,
		WorkingMode: pb.WiFiWorkingMode(src.Mode),
		AuthMode:    pb.WiFiAuthMode(src.AuthMode),
		Nexmon:      src.Nexmon,
		Channel:     uint32(src.Channel),
		HideSsid:    src.HideSsid,
		Ap_BSS: &pb.WiFiBSSCfg{
			SSID: src.Ap_BSS.SSID,
			PSK:  src.Ap_BSS.PSK,
		},
	}

	//Check if ranges are present
	if src.Client_BSSList != js.Undefined {
		if numClientBss := src.Client_BSSList.Length(); numClientBss > 0 {
			target.Client_BSSList = make([]*pb.WiFiBSSCfg, numClientBss)
			//iterate over JS array
			for i := 0; i < numClientBss; i++ {
				jsBsscfg := &jsWiFiBSSCfg{Object: src.Client_BSSList.Index(i)}
				target.Client_BSSList[i] = &pb.WiFiBSSCfg{
					SSID: jsBsscfg.SSID,
					PSK:  jsBsscfg.PSK,
				}
			}
		}
	} else {
		// at least on empty entry
		target.Client_BSSList = []*pb.WiFiBSSCfg{
			&pb.WiFiBSSCfg{},
		}
	}

	return target
}

func (target *jsWiFiSettings) fromGo(src *pb.WiFiSettings) {
	// assure undefined strings end up as empty strings

	target.Name = src.Name
	target.Mode = int(src.WorkingMode)
	target.Disabled = src.Disabled
	target.Reg = src.Regulatory
	target.AuthMode = int(src.AuthMode)
	target.Nexmon = src.Nexmon
	target.Channel = int(src.Channel)
	target.HideSsid = src.HideSsid
	// ToDo: Errorcheck existence of nested structs
	target.Ap_BSS = &jsWiFiBSSCfg{Object: O()}
	if src.Ap_BSS != nil {
		target.Ap_BSS.SSID = src.Ap_BSS.SSID
		target.Ap_BSS.PSK = src.Ap_BSS.PSK
	} else {
		target.Ap_BSS.SSID = ""
		target.Ap_BSS.PSK = ""
	}

	target.Client_BSSList = js.Global.Get("Array").New()
	for _, clientBSS := range src.Client_BSSList {
		jsClBSS := jsWiFiBSSCfg{Object: O()}
		jsClBSS.SSID = clientBSS.SSID
		jsClBSS.PSK = clientBSS.PSK
		target.Client_BSSList.Call("push", jsClBSS)
	}
	//if no entry in array, push at least one (needed in vue model)
	if target.Client_BSSList.Length() == 0 {
		jsClBSS := jsWiFiBSSCfg{Object: O()}
		jsClBSS.SSID = ""
		jsClBSS.PSK = ""
		target.Client_BSSList.Call("push", jsClBSS)
	}

	/*
	target.AP_SSID = src.Ap_BSS.SSID
	target.AP_PSK = src.Ap_BSS.PSK
	// ToDo: change STA settings to array
	target.STA_SSID = src.Client_BSSList[0].SSID
	target.STA_PSK = src.Client_BSSList[0].PSK
	*/
}

func NewWifiSettings() *jsWiFiSettings {
	res := &jsWiFiSettings{Object: O()}
	res.Disabled = true
	res.Mode = int(pb.WiFiWorkingMode_UNKNOWN)
	res.Name = "default"
	res.Ap_BSS = &jsWiFiBSSCfg{Object: O()}
	res.Ap_BSS.PSK = ""
	res.Ap_BSS.SSID = ""
	res.AuthMode = int(pb.WiFiAuthMode_OPEN)
	res.Channel = 1
	res.Nexmon = false
	res.Client_BSSList = js.Global.Get("Array").New()
	//no entry in array, push one (needed in vue model)
	jsClBSS := jsWiFiBSSCfg{Object: O()}
	jsClBSS.SSID = ""
	jsClBSS.PSK = ""
	res.Client_BSSList.Call("push", jsClBSS)
	res.HideSsid = false
	res.Reg = "US"

	return res
}

/* Network Settings */
type jsEthernetRequestSettingsStorage struct {
	*js.Object
	TemplateName string `js:"TemplateName"`
	Settings     *jsEthernetInterfaceSettings `js:"Settings"`
}

func (rs *jsEthernetRequestSettingsStorage) toGo() *pb.EthernetRequestSettingsStorage {
	return &pb.EthernetRequestSettingsStorage{
		Settings: rs.Settings.toGo(),
		TemplateName: rs.TemplateName,
	}
}

func (rs *jsEthernetRequestSettingsStorage) fromGo(src *pb.EthernetRequestSettingsStorage) {
	rs.TemplateName = src.TemplateName
	rs.Settings = &jsEthernetInterfaceSettings{Object:O()}
	rs.Settings.fromGo(src.Settings)
}

func NewEthernetRequestSettingsStorage() *jsEthernetRequestSettingsStorage {
	res := &jsEthernetRequestSettingsStorage{Object:O()}
	res.TemplateName = ""
	res.Settings = &jsEthernetInterfaceSettings{Object:O()}
	return res
}



type jsEthernetSettingsArray struct {
	*js.Object
	Interfaces *js.Object `js:"interfaces"` //every object property represents an EthernetSettings struct, the key is the interface name
}

func (isl *jsEthernetSettingsArray) fromGo(src *pb.DeployedEthernetInterfaceSettings) {
	//Options array (converted from map)
	isl.Interfaces = js.Global.Get("Array").New()
	for _, ifSets := range src.List {
		jsIfSets := &jsEthernetInterfaceSettings{Object: O()}
		jsIfSets.fromGo(ifSets)
		isl.Interfaces.Call("push", jsIfSets)
	}


	// sort the resulting array
	/*
	store.state.InterfaceSettings.interfaces.sort(function(a,b) {
		return a.name>b.name ? 1 : (a.name<b.name ? -1 : 0);
	})
	*/

	isl.Interfaces.Call("sort", func(a *jsEthernetInterfaceSettings, b *jsEthernetInterfaceSettings) int {
		switch {
		case a.Name > b.Name:
			return 1
		case a.Name < b.Name:
			return -1
		default:
			return 0
		}
	})
}

func (isl *jsEthernetSettingsArray) updateSingleInterface(updatedSettings *jsEthernetInterfaceSettings) {
	//Options array (converted from map)
	for i:=0; i<isl.Interfaces.Length(); i++ {
		 settings := &jsEthernetInterfaceSettings{Object:isl.Interfaces.Index(i)}
		 if settings.Name == updatedSettings.Name {
		 	// this are the settings to update (== replace)
		 	//isl.Interfaces.SetIndex(i, updatedSettings.Object)
		 	hvue.Set(isl.Interfaces,i, updatedSettings)
		 	return // we are done
		 }
	}

	// if here, the settings haven't been found, so we add them
	isl.Interfaces.Call("push", updatedSettings)

}

func NewEthernetSettingsList() (res *jsEthernetSettingsArray) {
	res = &jsEthernetSettingsArray{Object: O()}
	res.Interfaces = js.Global.Get("Array").New()
	return
}

type jsEthernetInterfaceSettings struct {
	*js.Object
	Name               string                `js:"name"`
	Mode               int                   `js:"mode"`
	IpAddress4         string                `js:"ipAddress4"`
	Netmask4           string                `js:"netmask4"`
	Enabled            bool                  `js:"enabled"`
	DhcpServerSettings *jsDHCPServerSettings `js:"dhcpServerSettings"`
	SettingsInUse      bool                  `js:"settingsInUse"`
}

func (target *jsEthernetInterfaceSettings) fromGo(src *pb.EthernetInterfaceSettings) {
	target.Name = src.Name
	target.Mode = int(src.Mode)
	target.IpAddress4 = src.IpAddress4
	target.Netmask4 = src.Netmask4
	target.Enabled = src.Enabled
	target.SettingsInUse = src.SettingsInUse

	if src.DhcpServerSettings != nil {
		target.DhcpServerSettings = &jsDHCPServerSettings{Object: O()}
		target.DhcpServerSettings.fromGo(src.DhcpServerSettings)
	}
}

func (src *jsEthernetInterfaceSettings) toGo() (target *pb.EthernetInterfaceSettings) {
	target = &pb.EthernetInterfaceSettings{
		Name:          src.Name,
		Mode:          pb.EthernetInterfaceSettings_Mode(src.Mode),
		IpAddress4:    src.IpAddress4,
		Netmask4:      src.Netmask4,
		Enabled:       src.Enabled,
		SettingsInUse: src.SettingsInUse,
	}

	if src.DhcpServerSettings.Object == js.Undefined {
		println("DHCPServerSettings on JS object undefined")
		target.DhcpServerSettings = nil
	} else {
		target.DhcpServerSettings = src.DhcpServerSettings.toGo()
	}
	return
}

func (iface *jsEthernetInterfaceSettings) CreateDhcpSettingsForInterface() {
	//create dhcp server settings
	settings := &jsDHCPServerSettings{Object: O()}
	settings.ListenInterface = iface.Name
	settings.ListenPort = 0 // 0 means DNS is disabled
	settings.LeaseFile = common_web.NameLeaseFileDHCPSrv(iface.Name)
	settings.NotAuthoritative = false
	settings.DoNotBindInterface = false
	settings.CallbackScript = ""
	//ToDo: add missing fields

	//Ranges array
	settings.Ranges = js.Global.Get("Array").New()
	settings.Options = js.Global.Get("Array").New()
	settings.StaticHosts = js.Global.Get("Array").New()

	//add empty option for router and DNS to prevent netmask from promoting itself via DHCP
	optNoRouter := &jsDHCPServerOption{Object: O()}
	optNoRouter.Option = 3
	optNoRouter.Value = ""
	settings.AddOption(optNoRouter)
	optNoDNS := &jsDHCPServerOption{Object: O()}
	optNoDNS.Option = 6
	optNoDNS.Value = ""
	settings.AddOption(optNoDNS)

	//iface.DhcpServerSettings = settings
	// Update the field with Vue in order to have proper setters in place
	hvue.Set(iface, "dhcpServerSettings", settings)
}

type jsDHCPServerSettings struct {
	*js.Object
	ListenPort         int        `js:"listenPort"`
	ListenInterface    string     `js:"listenInterface"`
	LeaseFile          string     `js:"leaseFile"`
	NotAuthoritative   bool       `js:"nonAuthoritative"`
	DoNotBindInterface bool       `js:"doNotBindInterface"`
	CallbackScript     string     `js:"callbackScript"`
	Ranges             *js.Object `js:"ranges"`      //[]*DHCPServerRange
	Options            *js.Object `js:"options"`     //map[uint32]string
	StaticHosts        *js.Object `js:"staticHosts"` //[]*DHCPServerStaticHost
}

func (src *jsDHCPServerSettings) toGo() (target *pb.DHCPServerSettings) {
	target = &pb.DHCPServerSettings{}

	target.ListenPort = uint32(src.ListenPort)
	target.ListenInterface = src.ListenInterface
	target.LeaseFile = src.LeaseFile
	target.NotAuthoritative = src.NotAuthoritative
	target.DoNotBindInterface = src.DoNotBindInterface
	target.CallbackScript = src.CallbackScript

	//println("jsRanges", src.Ranges)

	//Check if ranges are present
	if src.Ranges != js.Undefined {
		if numRanges := src.Ranges.Length(); numRanges > 0 {
			target.Ranges = make([]*pb.DHCPServerRange, numRanges)
			//iterate over JS array
			for i := 0; i < numRanges; i++ {
				jsRange := &jsDHCPServerRange{Object: src.Ranges.Index(i)}
				target.Ranges[i] = &pb.DHCPServerRange{}
				target.Ranges[i].RangeUpper = jsRange.RangeUpper
				target.Ranges[i].RangeLower = jsRange.RangeLower
				target.Ranges[i].LeaseTime = jsRange.LeaseTime
			}
		}
	}

	//Check if options are present
	if src.Options != js.Undefined {
		if numOptions := src.Options.Length(); numOptions > 0 {
			target.Options = make(map[uint32]string)
			//iterate over JS array
			for i := 0; i < numOptions; i++ {
				jsOption := &jsDHCPServerOption{Object: src.Options.Index(i)}
				target.Options[uint32(jsOption.Option)] = jsOption.Value
			}
		}
	}

	//Check if SaticHosts are present
	if src.StaticHosts != js.Undefined {
		if numStaticHosts := src.StaticHosts.Length(); numStaticHosts > 0 {
			target.StaticHosts = make([]*pb.DHCPServerStaticHost, numStaticHosts)
			//iterate over JS array
			for i := 0; i < numStaticHosts; i++ {
				jsStaticHost := &jsDHCPServerStaticHost{Object: src.StaticHosts.Index(i)}
				target.StaticHosts[i] = &pb.DHCPServerStaticHost{}
				target.StaticHosts[i].Mac = jsStaticHost.Mac
				target.StaticHosts[i].Ip = jsStaticHost.Ip
			}
		}
	}
	return target
}

func (settings *jsDHCPServerSettings) AddRange(dhcpRange *jsDHCPServerRange) {
	if settings.Ranges == js.Undefined {
		settings.Ranges = js.Global.Get("Array").New()
	}
	settings.Ranges.Call("push", dhcpRange)
}

func (settings *jsDHCPServerSettings) RemoveRange(dhcpRange *jsDHCPServerRange) {
	if settings.Ranges == js.Undefined {
		return
	}

	//Check if in array
	if idx := settings.Ranges.Call("indexOf", dhcpRange).Int(); idx > -1 {
		settings.Ranges.Call("splice", idx, 1)
	}
}

func (settings *jsDHCPServerSettings) AddOption(dhcpOption *jsDHCPServerOption) {
	if settings.Options == js.Undefined {
		settings.Options = js.Global.Get("Array").New()
	}
	settings.Options.Call("push", dhcpOption)
}

func (settings *jsDHCPServerSettings) RemoveOption(dhcpOption *jsDHCPServerOption) {
	if settings.Options == js.Undefined {
		return
	}

	//Check if in array
	if idx := settings.Options.Call("indexOf", dhcpOption).Int(); idx > -1 {
		settings.Options.Call("splice", idx, 1)
	}
}

func (settings *jsDHCPServerSettings) AddStaticHost(dhcpStaticHost *jsDHCPServerStaticHost) {
	if settings.StaticHosts == js.Undefined {
		settings.StaticHosts = js.Global.Get("Array").New()
	}
	settings.StaticHosts.Call("push", dhcpStaticHost)
}

func (settings *jsDHCPServerSettings) RemoveStaticHost(dhcpStaticHost *jsDHCPServerStaticHost) {
	if settings.StaticHosts == js.Undefined {
		return
	}

	//Check if in array
	if idx := settings.StaticHosts.Call("indexOf", dhcpStaticHost).Int(); idx > -1 {
		settings.StaticHosts.Call("splice", idx, 1)
	}
}

func (target *jsDHCPServerSettings) fromGo(src *pb.DHCPServerSettings) {
	target.ListenPort = int(src.ListenPort)
	target.ListenInterface = src.ListenInterface
	target.LeaseFile = src.LeaseFile
	target.NotAuthoritative = src.NotAuthoritative
	target.DoNotBindInterface = src.DoNotBindInterface
	target.CallbackScript = src.CallbackScript

	//Ranges array
	target.Ranges = js.Global.Get("Array").New()
	for _, dhcpRange := range src.Ranges {
		jsRange := &jsDHCPServerRange{Object: O()}
		jsRange.fromGo(dhcpRange)
		target.Ranges.Call("push", jsRange)
	}

	//Options array (converted from map)
	target.Options = js.Global.Get("Array").New()
	for optId, optVal := range src.Options {
		jsOption := &jsDHCPServerOption{Object: O()}
		jsOption.fromGo(optId, optVal)
		target.Options.Call("push", jsOption)
	}

	//StaticHosts array
	target.StaticHosts = js.Global.Get("Array").New()
	for _, staticHost := range src.StaticHosts {
		jsStaticHost := &jsDHCPServerStaticHost{Object: O()}
		jsStaticHost.fromGo(staticHost)
		target.Ranges.Call("push", jsStaticHost)
	}
}

type jsDHCPServerRange struct {
	*js.Object
	RangeLower string `js:"rangeLower"`
	RangeUpper string `js:"rangeUpper"`
	LeaseTime  string `js:"leaseTime"`
}

func (target *jsDHCPServerRange) fromGo(src *pb.DHCPServerRange) {
	target.RangeLower = src.RangeLower
	target.RangeUpper = src.RangeUpper
	target.LeaseTime = src.LeaseTime
}

type jsDHCPServerOption struct {
	*js.Object
	Option int    `js:"option"`
	Value  string `js:"value"`
}

func (target *jsDHCPServerOption) fromGo(srcID uint32, srcVal string) {
	target.Option = int(srcID)
	target.Value = srcVal
}

type jsDHCPServerStaticHost struct {
	*js.Object
	Mac string `js:"mac"`
	Ip  string `js:"ip"`
}

func (target *jsDHCPServerStaticHost) fromGo(src *pb.DHCPServerStaticHost) {
	target.Mac = src.Mac
	target.Ip = src.Ip
}

/* EVENT LOGGER */
type jsEventProcessor struct {
	*js.Object
	LogArray      *js.Object         `js:"logArray"`
	HidEventArray *js.Object         `js:"eventHidArray"`
	MaxEntries    int                `js:"maxEntries"`
	JobList       *jsHidJobStateList `js:"jobList"` //Needs to be exposed to JS in order to use JobList.UpdateEntry() from this JS object
}

func NewEventProcessor(maxEntries int, jobList *jsHidJobStateList) *jsEventProcessor {
	eventProcesssor := &jsEventProcessor{
		Object: js.Global.Get("Object").New(),
	}

	eventProcesssor.LogArray = js.Global.Get("Array").New()
	eventProcesssor.HidEventArray = js.Global.Get("Array").New()
	eventProcesssor.MaxEntries = maxEntries
	eventProcesssor.JobList = jobList

	return eventProcesssor
}

func (data *jsEventProcessor) handleHidEvent(hEv *jsHidEvent) {
	println("Received HID EVENT", hEv)
	switch hEv.EvType {
	case common_web.HidEventType_JOB_STARTED:
		// Note: the JOB_STARTED event carries the script source in the message field, (no need to re-request the job
		// state in order to retrieve the source code of the job, when adding it to the job state list)
		data.JobList.UpdateEntry(hEv.JobId, hEv.VMId, hEv.HasError, false, "Script started", "", hEv.EvLogTime, hEv.Message)
	case common_web.HidEventType_JOB_FAILED:
		data.JobList.UpdateEntry(hEv.JobId, hEv.VMId, hEv.HasError, false, hEv.Message, hEv.Error, hEv.EvLogTime, "")

		QuasarNotifyError("HIDScript job " + strconv.Itoa(int(hEv.JobId)) + " failed", hEv.Error, QUASAR_NOTIFICATION_POSITION_TOP)
	case common_web.HidEventType_JOB_SUCCEEDED:
		data.JobList.UpdateEntry(hEv.JobId, hEv.VMId, hEv.HasError, true, hEv.Message, hEv.Result, hEv.EvLogTime, "")

		QuasarNotifySuccess("HIDScript job " + strconv.Itoa(int(hEv.JobId)) + " succeeded", hEv.Result, QUASAR_NOTIFICATION_POSITION_TOP)
	case common_web.HidEventType_JOB_CANCELLED:
		data.JobList.UpdateEntry(hEv.JobId, hEv.VMId, true, false, hEv.Message, hEv.Message, hEv.EvLogTime, "")
	default:
		println("unhandled hid event "+common_web.EventTypeHIDName[hEv.EvType], hEv)
	}

}

/* This method gets internalized and therefor the mutex won't be accessible*/
func (data *jsEventProcessor) HandleEvent(ev *pb.Event) {
	go func() {
		jsEv := NewJsEventFromNative(ev)
		switch jsEv.Type {
		//if LOG event add to logArray
		case common_web.EVT_LOG:
			if logEv, err := jsEv.toLogEvent(); err == nil {
				data.LogArray.Call("push", logEv)
			} else {
				println("couldn't convert to LogEvent: ", jsEv)
			}
			//if HID event add to eventHidArray
		case common_web.EVT_HID:
			if hidEv, err := jsEv.toHidEvent(); err == nil {
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


