package main

import "github.com/gopherjs/gopherjs/js"

type triggerType int
type actionType int
const (
	TriggerServiceStarted = triggerType(0)
	TriggerUsbGadgetConnected = triggerType(1)
	TriggerUsbGadgetDisconnected = triggerType(2)
	TriggerWifiAPStarted = triggerType(3)
	TriggerWifiConnectedAsSta = triggerType(4)
	TriggerSshLogin = triggerType(5)
	TriggerDhcpLeaseGranted = triggerType(6)
	TriggerGPIOIn = triggerType(7)
	TriggerGroupReceive = triggerType(8)
	TriggerGroupReceiveSequence = triggerType(9)

	ActionLog = actionType(0)
	ActionHidScript = actionType(1)
	ActionDeploySettingsTemplate = actionType(2)
	ActionBashScript = actionType(3)
	ActionGPIOOut = actionType(4)
	ActionGroupSend = actionType(5)
)
var triggerNames = map[triggerType]string{
	TriggerServiceStarted: "Service started",
	TriggerUsbGadgetConnected: "USB Gadget connected to host",
	TriggerUsbGadgetDisconnected: "USB Gadget disconnected from host",
	TriggerWifiAPStarted: "WiFi Access Point is up",
	TriggerWifiConnectedAsSta: "Connected to existing WiFi",
	TriggerSshLogin: "User login via SSH",
	TriggerDhcpLeaseGranted: "Client received DHCP lease",
	TriggerGPIOIn: "GPIO Pin input",
	TriggerGroupReceive: "Group channel received value",
	TriggerGroupReceiveSequence: "Group channel received sequence",
}
var actionNames = map[actionType]string{
	ActionLog: "Log to internal console",
	ActionHidScript: "Start a HIDScript",
	ActionDeploySettingsTemplate: "Load and deploy the given settings",
	ActionBashScript: "Run the given bash script",
	ActionGPIOOut: "GPIO Pin output",
	ActionGroupSend: "Send value to group channel",
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
	TriggerGroupReceiveSequence,
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

func (ta *jsTriggerAction) ChangeActionType(newAt actionType) {
	var data *js.Object
	switch newAt {
	case ActionLog:
		d := &jsActionLog{Object:O()}
		data = d.Object
	case ActionHidScript:
		d := &jsActionStartHIDScript{Object:O()}
		d.ScriptName = "somescript"
		data = d.Object
	case ActionDeploySettingsTemplate:
		d := &jsActionDeploySettingsTemplate{Object:O()}
		d.TemplateName = "somescript"
		d.Type = "Template type"
		data = d.Object
	case ActionBashScript:
		d := &jsActionStartBashScript{Object:O()}
		d.ScriptName = "/path/to/some/script"
		data = d.Object
	case ActionGPIOOut:
		d := &jsActionGPIOOut{Object:O()}
		d.GpioNum = 1
		d.Value = GPIOOutValueHigh
		data = d.Object
	case ActionGroupSend:
		d := &jsActionGroupSend{Object:O()}
		d.GroupName = "Channel1"
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
		d.ResLoginUser = "root"
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
		d.GpioNum = 2
		d.PullUpDown = GPIOInPullUp
		data = d.Object
	case TriggerGroupReceive:
		d := &jsTriggerGroupReceive{Object:O()}
		d.GroupName = "Channel1"
		d.Value = 0
		data = d.Object
	case TriggerGroupReceiveSequence:
		d := &jsTriggerGroupReceiveSequence{Object:O()}
		d.GroupName = "Channel1"
		d.IgnoreOutOfOrder = false
		d.ValueSequence = []int{1,1}
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
func (ta *jsTriggerAction) IsTriggerGroupReceiveSequence() bool {
	return ta.TriggerType == TriggerGroupReceiveSequence
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
	ResLoginUser string `js:"ResLoginUser"`
}
type jsTriggerDHCPLeaseGranted struct {
	*js.Object
}
type jsTriggerGroupReceive struct {
	*js.Object
	GroupName string `js:"GroupName"`
	Value int `js:"Value"`
}
type jsTriggerGroupReceiveSequence struct {
	*js.Object
	GroupName string `js:"GroupName"`
	IgnoreOutOfOrder bool `js:"IgnoreOutOfOrder"`
	ValueSequence []int `js:"ValueSequence"`
}
type jsTriggerGPIOIn struct {
	*js.Object
	GpioNum int `js:"GpioNum"`
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
	Type         string `js:"Type"`
}
type jsActionGPIOOut struct {
	*js.Object
	GpioNum int `js:"GpioNum"`
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
	Value int `js:"Value"`
}
type jsActionLog struct {
	*js.Object
}