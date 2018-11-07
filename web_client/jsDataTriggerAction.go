// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	pb "github.com/mame82/P4wnP1_go/proto/gopherjs"
	"github.com/mame82/hvue"
	"strconv"
)

type triggerType int
type actionType int
const (
	TriggerServiceStarted        = triggerType(0)
	TriggerUsbGadgetConnected    = triggerType(1)
	TriggerUsbGadgetDisconnected = triggerType(2)
	TriggerWifiAPStarted         = triggerType(3)
	TriggerWifiConnectedAsSta    = triggerType(4)
	TriggerSshLogin              = triggerType(5)
	TriggerDhcpLeaseGranted      = triggerType(6)
	TriggerGPIOIn                = triggerType(7)
	TriggerGroupReceive          = triggerType(8)
	TriggerGroupReceiveMulti     = triggerType(9)

	ActionLog = actionType(0)
	ActionHidScript = actionType(1)
	ActionDeploySettingsTemplate = actionType(2)
	ActionBashScript = actionType(3)
	ActionGPIOOut = actionType(4)
	ActionGroupSend = actionType(5)
)
var triggerNames = map[triggerType]string{
	TriggerServiceStarted:        "service started",
	TriggerUsbGadgetConnected:    "USB gadget connected to host",
	TriggerUsbGadgetDisconnected: "USB Gadget disconnected from host",
	TriggerWifiAPStarted:         "WiFi Access Point is up",
	TriggerWifiConnectedAsSta:    "joined existing WiFi",
	TriggerSshLogin:              "SSH user login",
	TriggerDhcpLeaseGranted:      "DHCP lease issued",
	TriggerGPIOIn:                "input on GPIO",
	TriggerGroupReceive:          "a value on a group channel",
	TriggerGroupReceiveMulti:     "multiple values on a group channel",
}
var actionNames = map[actionType]string{
	ActionLog: "write log entry",
	ActionHidScript: "start a HIDScript",
	ActionDeploySettingsTemplate: "load and deploy settings template",
	ActionBashScript: "run a bash script",
	ActionGPIOOut: "set output on GPIO",
	ActionGroupSend: "send a value to a group channel",
}
var availableTriggers = []triggerType{
	TriggerServiceStarted,
	TriggerUsbGadgetConnected,
	TriggerUsbGadgetDisconnected,
	TriggerWifiAPStarted,
	TriggerWifiConnectedAsSta,
	TriggerDhcpLeaseGranted,
	TriggerGPIOIn,
	TriggerSshLogin,
	TriggerGroupReceive,
	TriggerGroupReceiveMulti,
}
var availableActions = []actionType {
	ActionLog,
	ActionBashScript,
	ActionHidScript,
	ActionDeploySettingsTemplate,
	ActionGPIOOut,
	ActionGroupSend,
}

type jsTriggerAction struct {
	*js.Object

	Id      uint32 `js:"Id"`
	OneShot bool `js:"OneShot"`
	IsActive bool `js:"IsActive"`
	Immutable bool `js:"Immutable"`


	TriggerType triggerType `js:"TriggerType"`
	ActionType actionType `js:"ActionType"`

	TriggerData *js.Object `js:"TriggerData"`
	ActionData *js.Object `js:"ActionData"`
}

func (dst *jsTriggerAction) fromGo(src *pb.TriggerAction) {
	dst.IsActive = src.IsActive
	dst.Immutable = src.Immutable
	dst.OneShot = src.OneShot
	dst.Id = src.Id

	// convert action
	switch srcAction := src.Action.(type) {
	case *pb.TriggerAction_Log:
		dst.ChangeActionType(ActionLog)
	case *pb.TriggerAction_HidScript:
		dst.ChangeActionType(ActionHidScript)
		dstAction := &jsActionStartHIDScript{Object: dst.ActionData}
		dstAction.ScriptName = srcAction.HidScript.ScriptName
	case *pb.TriggerAction_DeploySettingsTemplate:
		dst.ChangeActionType(ActionDeploySettingsTemplate)
		dstAction := &jsActionDeploySettingsTemplate{Object: dst.ActionData}
		dstAction.TemplateName = srcAction.DeploySettingsTemplate.TemplateName
		dstAction.Type = TemplateType(srcAction.DeploySettingsTemplate.Type)
	case *pb.TriggerAction_BashScript:
		dst.ChangeActionType(ActionBashScript)
		dstAction := &jsActionStartBashScript{Object: dst.ActionData}
		dstAction.ScriptName = srcAction.BashScript.ScriptName
	case *pb.TriggerAction_GpioOut:
		dst.ChangeActionType(ActionGPIOOut)
		dstAction := &jsActionGPIOOut{Object: dst.ActionData}
		dstAction.Value = GPIOOutValue(srcAction.GpioOut.Value)
		dstAction.GpioName = srcAction.GpioOut.GpioName
	case *pb.TriggerAction_GroupSend:
		dst.ChangeActionType(ActionGroupSend)
		dstAction := &jsActionGroupSend{Object: dst.ActionData}
		dstAction.Value = srcAction.GroupSend.Value
		dstAction.GroupName = srcAction.GroupSend.GroupName
	default:
		// do nothing
		// Note: no default case, we don't change any values of jsTriggerAction if there isn't a type match
	}

	// convert trigger
	switch srcTrigger := src.Trigger.(type) {
	case *pb.TriggerAction_SshLogin:
		dst.ChangeTriggerType(TriggerSshLogin)
		dstTrigger := &jsTriggerSSHLogin{Object: dst.TriggerData}
		dstTrigger.LoginUser = srcTrigger.SshLogin.LoginUser
	case *pb.TriggerAction_DhcpLeaseGranted:
		dst.ChangeTriggerType(TriggerDhcpLeaseGranted)
	case *pb.TriggerAction_WifiAPStarted:
		dst.ChangeTriggerType(TriggerWifiAPStarted)
	case *pb.TriggerAction_WifiConnectedAsSta:
		dst.ChangeTriggerType(TriggerWifiConnectedAsSta)
	case *pb.TriggerAction_UsbGadgetConnected:
		dst.ChangeTriggerType(TriggerUsbGadgetConnected)
	case *pb.TriggerAction_UsbGadgetDisconnected:
		dst.ChangeTriggerType(TriggerUsbGadgetDisconnected)
	case *pb.TriggerAction_ServiceStarted:
		dst.ChangeTriggerType(TriggerServiceStarted)
	case *pb.TriggerAction_GpioIn:
		dst.ChangeTriggerType(TriggerGPIOIn)
		dstTrigger := &jsTriggerGPIOIn{Object: dst.TriggerData}
		dstTrigger.GpioName = srcTrigger.GpioIn.GpioName
		dstTrigger.Edge = GPIOInEdge(srcTrigger.GpioIn.GpioInEdge)
		dstTrigger.PullUpDown = GPIOInPullUpDown(srcTrigger.GpioIn.PullUpDown)
	case *pb.TriggerAction_GroupReceive:
		dst.ChangeTriggerType(TriggerGroupReceive)
		dstTrigger := &jsTriggerGroupReceive{Object: dst.TriggerData}
		dstTrigger.GroupName = srcTrigger.GroupReceive.GroupName
		dstTrigger.Value = srcTrigger.GroupReceive.Value
	case *pb.TriggerAction_GroupReceiveMulti:
		dst.ChangeTriggerType(TriggerGroupReceiveMulti)
		dstTrigger := &jsTriggerGroupReceiveMulti{Object: dst.TriggerData}
		dstTrigger.GroupName = srcTrigger.GroupReceiveMulti.GroupName
		dstTrigger.Values = srcTrigger.GroupReceiveMulti.Values
		dstTrigger.Type = GroupReceiveMultiType(srcTrigger.GroupReceiveMulti.Type)
	default:
		// change nothing
	}
}


func (ta *jsTriggerAction) toGo() (res *pb.TriggerAction) {
	res = &pb.TriggerAction{
		OneShot: ta.OneShot,
		Immutable: ta.Immutable,
		IsActive: ta.IsActive,
		Id: ta.Id,
	}

	// Convert action
	switch ta.ActionType {
	case ActionLog:
		res.Action = &pb.TriggerAction_Log{
			Log: &pb.ActionLog{},
		}
	case ActionHidScript:
		actionData := &jsActionStartHIDScript{Object: ta.ActionData}
		res.Action = &pb.TriggerAction_HidScript{
			HidScript: &pb.ActionStartHIDScript{
				ScriptName: actionData.ScriptName,
			},
		}
	case ActionDeploySettingsTemplate:
		actionData := &jsActionDeploySettingsTemplate{Object: ta.ActionData}
		res.Action = &pb.TriggerAction_DeploySettingsTemplate{
			DeploySettingsTemplate: &pb.ActionDeploySettingsTemplate{
				Type: pb.ActionDeploySettingsTemplate_TemplateType(actionData.Type),
				TemplateName: actionData.TemplateName,
			},
		}
	case ActionBashScript:
		actionData := &jsActionStartBashScript{Object: ta.ActionData}
		res.Action = &pb.TriggerAction_BashScript {
			BashScript: &pb.ActionStartBashScript{
				ScriptName: actionData.ScriptName,
			},
		}
	case ActionGPIOOut:
		actionData := &jsActionGPIOOut{Object: ta.ActionData}
		res.Action = &pb.TriggerAction_GpioOut{
			GpioOut: &pb.ActionGPIOOut{
				GpioName: actionData.GpioName,
				Value: pb.GPIOOutValue(actionData.Value),
			},
		}
	case ActionGroupSend:
		actionData := &jsActionGroupSend{Object: ta.ActionData}
		res.Action = &pb.TriggerAction_GroupSend{
			GroupSend: &pb.ActionGroupSend{
				GroupName: actionData.GroupName,
				Value: actionData.Value,
			},
		}
	default:
		println("Unknown action type")
		res.Action = nil
	}

	// convert trigger
	switch ta.TriggerType {
	case TriggerSshLogin:
		triggerData := &jsTriggerSSHLogin{Object: ta.TriggerData}
		res.Trigger = &pb.TriggerAction_SshLogin{
			SshLogin: &pb.TriggerSSHLogin{
				LoginUser: triggerData.LoginUser,
			},
		}
	case TriggerDhcpLeaseGranted:
		//triggerData := &jsTriggerDHCPLeaseGranted{Object: ta.TriggerData}
		res.Trigger = &pb.TriggerAction_DhcpLeaseGranted{
			DhcpLeaseGranted: &pb.TriggerDHCPLeaseGranted{},
		}
	case TriggerWifiAPStarted:
		res.Trigger = &pb.TriggerAction_WifiAPStarted{
			WifiAPStarted: &pb.TriggerWifiAPStarted{},
		}
	case TriggerWifiConnectedAsSta:
		res.Trigger = &pb.TriggerAction_WifiConnectedAsSta{
			WifiConnectedAsSta: &pb.TriggerWifiConnectedAsSta{},
		}
	case TriggerUsbGadgetConnected:
		res.Trigger = &pb.TriggerAction_UsbGadgetConnected{
			UsbGadgetConnected: &pb.TriggerUSBGadgetConnected{},
		}
	case TriggerUsbGadgetDisconnected:
		res.Trigger = &pb.TriggerAction_UsbGadgetDisconnected{
			UsbGadgetDisconnected: &pb.TriggerUSBGadgetDisconnected{},
		}
	case TriggerServiceStarted:
		res.Trigger = &pb.TriggerAction_ServiceStarted{
			ServiceStarted: &pb.TriggerServiceStarted{},
		}
	case TriggerGPIOIn:
		triggerData := &jsTriggerGPIOIn{Object: ta.TriggerData}
		res.Trigger = &pb.TriggerAction_GpioIn{
			GpioIn: &pb.TriggerGPIOIn{
				GpioName: triggerData.GpioName,
				PullUpDown: pb.GPIOInPullUpDown(triggerData.PullUpDown),
				GpioInEdge: pb.GPIOInEdge(triggerData.Edge),
			},
		}
	case TriggerGroupReceive:
		triggerData := &jsTriggerGroupReceive{Object: ta.TriggerData}
		res.Trigger = &pb.TriggerAction_GroupReceive{
			GroupReceive: &pb.TriggerGroupReceive {
				GroupName: triggerData.GroupName,
				Value: triggerData.Value,
			},
		}
	case TriggerGroupReceiveMulti:
		triggerData := &jsTriggerGroupReceiveMulti{Object: ta.TriggerData}
		res.Trigger = &pb.TriggerAction_GroupReceiveMulti{
			GroupReceiveMulti: &pb.TriggerGroupReceiveMulti {
				GroupName: triggerData.GroupName,
				Values: triggerData.Values,
				Type: pb.GroupReceiveMultiType(triggerData.Type),
			},
		}
	default:
		println("Unknown trigger type")
		res.Trigger = nil
	}


	return res
}


func (ta *jsTriggerAction) ChangeActionType(newAt actionType) {
	var data *js.Object
	switch newAt {
	case ActionLog:
		d := &jsActionLog{Object:O()}
		data = d.Object
	case ActionHidScript:
		d := &jsActionStartHIDScript{Object:O()}
		d.ScriptName = ""
		data = d.Object
	case ActionDeploySettingsTemplate:
		d := &jsActionDeploySettingsTemplate{Object:O()}
		d.TemplateName = ""
		d.Type = availableTemplateTypes[0]
		data = d.Object
	case ActionBashScript:
		d := &jsActionStartBashScript{Object:O()}
		d.ScriptName = ""
		data = d.Object
	case ActionGPIOOut:
		d := &jsActionGPIOOut{Object:O()}
		d.GpioName = ""
		d.Value = GPIOOutValueHigh
		data = d.Object
	case ActionGroupSend:
		d := &jsActionGroupSend{Object:O()}
		d.GroupName = "Group1"
		d.Value = 1
		data = d.Object
	default:
		println("Unknown action type")
		data = O()
	}

	ta.ActionData = data
	ta.ActionType = newAt
}

func (ta *jsTriggerAction) IsActionLog() bool {
	return ta.ActionType == ActionLog
}

func (ta *jsTriggerAction) IsActionBashScript() bool {
	return ta.ActionType == ActionBashScript
}

func (ta *jsTriggerAction) IsActionHidScript() bool {
	return ta.ActionType == ActionHidScript
}

func (ta *jsTriggerAction) IsActionDeploySettingsTemplate() bool {
	return ta.ActionType == ActionDeploySettingsTemplate
}
func (ta *jsTriggerAction) IsActionGroupSend() bool {
	return ta.ActionType == ActionGroupSend
}
func (ta *jsTriggerAction) IsActionGPIOOut() bool {
	return ta.ActionType == ActionGPIOOut
}

func (ta *jsTriggerAction) ChangeTriggerType(newTt triggerType) {
	var data *js.Object
	switch newTt {
	case TriggerSshLogin:
		d := &jsTriggerSSHLogin{Object:O()}
		d.LoginUser = "root"
		data = d.Object
	case TriggerDhcpLeaseGranted:
		d := &jsTriggerDHCPLeaseGranted{Object:O()}
		data = d.Object
	case TriggerWifiAPStarted:
		d := &jsTriggerWifiAPStarted{Object:O()}
		data = d.Object
	case TriggerWifiConnectedAsSta:
		d := &jsTriggerWifiConnectedAsSta{Object:O()}
		data = d.Object
	case TriggerUsbGadgetConnected:
		d := &jsTriggerUSBGadgetConnected{Object:O()}
		data = d.Object
	case TriggerUsbGadgetDisconnected:
		d := &jsTriggerUSBGadgetDisconnected{Object:O()}
		data = d.Object
	case TriggerServiceStarted:
		d := &jsTriggerServiceStarted{Object:O()}
		data = d.Object
	case TriggerGPIOIn:
		d := &jsTriggerGPIOIn{Object:O()}
		d.Edge = GPIOInEdgeRising
		d.GpioName = ""
		d.PullUpDown = GPIOInPullUp
		data = d.Object
	case TriggerGroupReceive:
		d := &jsTriggerGroupReceive{Object:O()}
		d.GroupName = "Group1"
		d.Value = 0
		data = d.Object
	case TriggerGroupReceiveMulti:
		d := &jsTriggerGroupReceiveMulti{Object: O()}
		d.GroupName = "Group1"
		d.Type = GroupReceiveMultiType_SEQUENCE
		d.Values = []int32{1,2}
		data = d.Object
	default:
		println("Unknown trigger type")
		data = O()
	}

	ta.TriggerData = data
	ta.TriggerType = newTt
}

func (ta *jsTriggerAction) IsTriggerServiceStarted() bool {
	return ta.TriggerType == TriggerServiceStarted
}
func (ta *jsTriggerAction) IsTriggerSshLogin() bool {
	return ta.TriggerType == TriggerSshLogin
}
func (ta *jsTriggerAction) IsTriggerDhcpLeaseGranted() bool {
	return ta.TriggerType == TriggerDhcpLeaseGranted
}
func (ta *jsTriggerAction) IsTriggerWifiAPStarted() bool {
	return ta.TriggerType == TriggerWifiAPStarted
}
func (ta *jsTriggerAction) IsTriggerWifiConnectedAsSta() bool {
	return ta.TriggerType == TriggerWifiConnectedAsSta
}
func (ta *jsTriggerAction) IsTriggerUsbGadgetConnected() bool {
	return ta.TriggerType == TriggerUsbGadgetConnected
}
func (ta *jsTriggerAction) IsTriggerUsbGadgetDisconnected() bool {
	return ta.TriggerType == TriggerUsbGadgetDisconnected
}
func (ta *jsTriggerAction) IsTriggerGPIOIn() bool {
	return ta.TriggerType == TriggerGPIOIn
}
func (ta *jsTriggerAction) IsTriggerGroupReceive() bool {
	return ta.TriggerType == TriggerGroupReceive
}
func (ta *jsTriggerAction) IsTriggerGroupReceiveMulti() bool {
	return ta.TriggerType == TriggerGroupReceiveMulti
}


func NewTriggerAction() *jsTriggerAction {
	ta := &jsTriggerAction{Object: O()}
	ta.Id = 0
	ta.IsActive = true
	ta.Immutable = false
	ta.OneShot = false
	ta.ActionData = O()
	ta.TriggerData = O()
	ta.TriggerType = availableTriggers[0]
	ta.ActionType = availableActions[0]
	return ta
}

type jsTriggerServiceStarted struct {
	*js.Object
}

type jsTriggerUSBGadgetConnected struct {
	*js.Object
}

type jsTriggerUSBGadgetDisconnected struct {
	*js.Object
}

type jsTriggerWifiAPStarted struct {
	*js.Object
}
type jsTriggerWifiConnectedAsSta struct {
	*js.Object
}
type jsTriggerSSHLogin struct {
	*js.Object
	LoginUser string `js:"LoginUser"`
}
type jsTriggerDHCPLeaseGranted struct {
	*js.Object
}
type jsTriggerGroupReceive struct {
	*js.Object
	GroupName string `js:"GroupName"`
	Value int32 `js:"Value"`
}

type jsTriggerGroupReceiveMulti struct {
	*js.Object
	GroupName string                `js:"GroupName"`
	Type      GroupReceiveMultiType `js:"Type"`
	Values    []int32               `js:"Values"`
}
type GroupReceiveMultiType int
const (
	GroupReceiveMultiType_SEQUENCE       GroupReceiveMultiType = 0
	GroupReceiveMultiType_AND            GroupReceiveMultiType = 1
	GroupReceiveMultiType_OR             GroupReceiveMultiType = 2
	GroupReceiveMultiType_EXACT_SEQUENCE GroupReceiveMultiType = 3
)
var groupReceiveMultiNames = map[GroupReceiveMultiType]string{
	GroupReceiveMultiType_SEQUENCE: "Ordered Sequence (out-of-oreder values allowed)",
	GroupReceiveMultiType_AND: "All (logical AND)",
	GroupReceiveMultiType_OR: "One of (logical OR)",
	GroupReceiveMultiType_EXACT_SEQUENCE: "Exact ordered sequence",
}
var availableGroupReceiveMulti = []GroupReceiveMultiType{GroupReceiveMultiType_SEQUENCE, GroupReceiveMultiType_EXACT_SEQUENCE, GroupReceiveMultiType_AND, GroupReceiveMultiType_OR}



type jsTriggerGPIOIn struct {
	*js.Object
	GpioName string `js:"GpioName"`
	PullUpDown GPIOInPullUpDown `js:"PullUpDown"` //PullUp resistor, pull down otherwise
	Edge GPIOInEdge `js:"Edge"` // 0 == GPIO.RISING, 1 == GPIO.FALLING, every value > 1 == GPIO.BOTH
}
type GPIOInPullUpDown int
const GPIOInPullUp = GPIOInPullUpDown(0)
const GPIOInPullDown = GPIOInPullUpDown(1)
const GPIOInPullOff = GPIOInPullUpDown(2)
var gpioInPullUpDownNames = map[GPIOInPullUpDown]string{
	GPIOInPullUp: "Pull-up",
	GPIOInPullDown: "Pull-down",
	GPIOInPullOff: "Off",
}
var availableGPIOInPullUpDowns = []GPIOInPullUpDown{GPIOInPullUp, GPIOInPullDown, GPIOInPullOff}

type GPIOInEdge int
const GPIOInEdgeRising = GPIOInEdge(0)
const GPIOInEdgeFalling = GPIOInEdge(1)
const GPIOInEdgeBoth = GPIOInEdge(2)
var gpioInEdgeNames = map[GPIOInEdge]string{
	GPIOInEdgeRising: "Rising edge",
	GPIOInEdgeFalling: "Falling edge",
	GPIOInEdgeBoth: "Rising or Falling edge",
}
var availableGPIOInEdges = []GPIOInEdge{GPIOInEdgeRising, GPIOInEdgeFalling, GPIOInEdgeBoth}

type jsActionStartBashScript struct {
	*js.Object
	ScriptName string `js:"ScriptName"`
}
type jsActionStartHIDScript struct {
	*js.Object
	ScriptName string `js:"ScriptName"`
}
type jsActionDeploySettingsTemplate struct {
	*js.Object
	TemplateName string `js:"TemplateName"`
	Type         TemplateType `js:"Type"`
}
type jsActionGPIOOut struct {
	*js.Object
	GpioName string `js:"GpioName"`
	Value GPIOOutValue `js:"Value"` //PullUp resistor, pull down otherwise
}
type GPIOOutValue int
const GPIOOutValueLow = GPIOOutValue(0)
const GPIOOutValueHigh = GPIOOutValue(1)
const GPIOOutValueToggle = GPIOOutValue(2)
var gpioOutValueNames = map[GPIOOutValue]string{
	GPIOOutValueLow: "Low",
	GPIOOutValueHigh: "High",
	GPIOOutValueToggle: "Toggle",
}
var availableGPIOOutValues = []GPIOOutValue{GPIOOutValueLow, GPIOOutValueHigh, GPIOOutValueToggle}


type jsActionGroupSend struct {
	*js.Object
	GroupName string `js:"GroupName"`
	Value int32 `js:"Value"`
}
type jsActionLog struct {
	*js.Object
}


type TemplateType int
const TemplateTypeFullSettings = TemplateType(pb.ActionDeploySettingsTemplate_FULL_SETTINGS)
const TemplateTypeNetwork = TemplateType(pb.ActionDeploySettingsTemplate_NETWORK)
const TemplateTypeWifi = TemplateType(pb.ActionDeploySettingsTemplate_WIFI)
const TemplateTypeUSB = TemplateType(pb.ActionDeploySettingsTemplate_USB)
const TemplateTypeBluetooth = TemplateType(pb.ActionDeploySettingsTemplate_BLUETOOTH)
const TemplateTypeTriggerActions = TemplateType(pb.ActionDeploySettingsTemplate_TRIGGER_ACTIONS)
var templateTypeNames = map[TemplateType]string{
	TemplateTypeFullSettings: "Overall settings",
	TemplateTypeNetwork: "Network interface settings",
	TemplateTypeWifi: "WiFi settings",
	TemplateTypeUSB: "USB settings",
	TemplateTypeBluetooth: "Bluetooth settings",
	TemplateTypeTriggerActions: "Stored TriggerAction set",
}
var availableTemplateTypes = []TemplateType{
	TemplateTypeFullSettings,
	TemplateTypeNetwork,
	TemplateTypeWifi,
	TemplateTypeUSB,
	TemplateTypeBluetooth,
	TemplateTypeTriggerActions,
}



/* TriggerActions */
type jsTriggerActionSet struct {
	*js.Object
	Name string `js:"Name"`
	TriggerActions *js.Object `js:"TriggerActions"`
}

func NewTriggerActionSet() *jsTriggerActionSet {
	tal := &jsTriggerActionSet{Object: O()}
	tal.TriggerActions = O()
	tal.Name = "default_ta_set"
	return tal
}

func (tal *jsTriggerActionSet) UpdateEntry(ta *jsTriggerAction) {
	key := strconv.Itoa(int(ta.Id))

	//Check if job exists, update existing one if already present
	var updateTa *jsTriggerAction
	if res := tal.TriggerActions.Get(key); res == js.Undefined {
		updateTa = &jsTriggerAction{Object: O()}
	} else {
		updateTa = &jsTriggerAction{Object: res}
	}

	//Create job object
	updateTa.Id = ta.Id
	updateTa.IsActive = ta.IsActive
	updateTa.Immutable = ta.Immutable
	updateTa.OneShot = ta.OneShot
	updateTa.ActionType = ta.ActionType
	updateTa.TriggerType = ta.TriggerType
	updateTa.TriggerData = ta.TriggerData
	updateTa.ActionData = ta.ActionData

	hvue.Set(tal.TriggerActions, key, updateTa)
}

func (tal *jsTriggerActionSet) DeleteEntry(id uint32) {
	tal.TriggerActions.Delete(strconv.Itoa(int(id))) //JS version
	//delete(jl.Jobs, strconv.Itoa(int(id)))
}

func (tas *jsTriggerActionSet) Flush() {
	tas.TriggerActions = O()
	//delete(jl.Jobs, strconv.Itoa(int(id)))
}

func (src jsTriggerActionSet) toGo() (target *pb.TriggerActionSet) {
	js_ta_array := js.Global.Get("Object").Call("values", src.TriggerActions)
	count := js_ta_array.Length()
	// println("tal len:", count)
	// iterate over array
	target = &pb.TriggerActionSet{
		Name: src.Name,
	}
	target.TriggerActions = make([]*pb.TriggerAction, count)
	for i:=0;i<count;i++ {
		jsTa := &jsTriggerAction{Object: js_ta_array.Index(i)}
		target.TriggerActions[i] = jsTa.toGo()
	}
	//println("Go TAS: ", target )
	return
}