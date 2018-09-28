// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/mame82/hvue"
)

func ExportDefaultTriggerActions() {
	// create test trigger

	// Trigger to run startup script
	triggerData := &jsTriggerServiceStarted{Object:O()}
	trigger := &jsTriggerAction_ServiceStarted{Object:O()}
	trigger.ServiceStarted = triggerData
	actionData := &jsActionStartBashScript{Object:O()}
	actionData.ScriptPath = "/usr/local/P4wnP1/scripts/servicestart.sh"
	action := &jsTriggerAction_BashScript{Object:O()}
	action.BashScript = actionData
	svcUpRunScript := &jsTriggerAction{Object:O()}
	svcUpRunScript.OneShot = false
	svcUpRunScript.Id = 0
	svcUpRunScript.Trigger = trigger.Object
	svcUpRunScript.Action = action.Object

	js.Global.Set("testta", svcUpRunScript)

	// Try to cast back (shouldn't work because of the interfaces
	copyobj := &jsTriggerAction{Object:js.Global.Get("testta")}
	js.Global.Set("copyobj", copyobj)
	println("copyobj", copyobj)
	println("copyobjtrigger", copyobj.Trigger) //<--- this wouldn't work

	if isJsTriggerAction_ServiceStarted(copyobj.Trigger) {
		println("is service started trigger")
	}
	if isJsTriggerAction_UsbGadgetConnected(copyobj.Trigger) {
		println("is USB gadget connected trigger")
	}
	if isJsTriggerAction_BashScript(copyobj.Action) {
		println("is BashScript action")
	}
	if isJsTriggerAction_HidScript(copyobj.Trigger) {
		println("is HIDScript action")
	}


	/*
	serviceUpRunScript := &pb.TriggerAction{
		Trigger: &pb.TriggerAction_ServiceStarted{
			ServiceStarted: &pb.TriggerServiceStarted{},
		},
		Action: &pb.TriggerAction_BashScript{
			BashScript: &pb.ActionStartBashScript{
				ScriptPath: "/usr/local/P4wnP1/scripts/servicestart.sh", // ToDo: use real script path once ready
			},
		},
	}
	a[0] = serviceUpRunScript

	logServiceStart := &pb.TriggerAction{
		Trigger: &pb.TriggerAction_ServiceStarted{
			ServiceStarted: &pb.TriggerServiceStarted{},
		},
		Action: &pb.TriggerAction_Log{
			Log: &pb.ActionLog{},
		},
	}
	a[1]= logServiceStart

	logDHCPLease := &pb.TriggerAction{
		Trigger: &pb.TriggerAction_DhcpLeaseGranted{
			DhcpLeaseGranted: &pb.TriggerDHCPLeaseGranted{},
		},
		Action: &pb.TriggerAction_Log{
			Log: &pb.ActionLog{},
		},
	}
	a[2] = logDHCPLease

	logUSBGadgetConnected := &pb.TriggerAction{
		Trigger: &pb.TriggerAction_UsbGadgetConnected{
			UsbGadgetConnected: &pb.TriggerUSBGadgetConnected{},
		},
		Action: &pb.TriggerAction_Log{
			Log: &pb.ActionLog{},
		},
	}
	a[3] = logUSBGadgetConnected

	logUSBGadgetDisconnected := &pb.TriggerAction{
		Trigger: &pb.TriggerAction_UsbGadgetDisconnected{
			UsbGadgetDisconnected: &pb.TriggerUSBGadgetDisconnected{},
		},
		Action: &pb.TriggerAction_Log{
			Log: &pb.ActionLog{},
		},
	}
	a[4] = logUSBGadgetDisconnected

	logWifiAp := &pb.TriggerAction{
		Trigger: &pb.TriggerAction_WifiAPStarted{
			WifiAPStarted: &pb.TriggerWifiAPStarted{},
		},
		Action: &pb.TriggerAction_Log{
			Log: &pb.ActionLog{},
		},
	}
	a[5] = logWifiAp

	logWifiSta := &pb.TriggerAction{
		Trigger: &pb.TriggerAction_WifiConnectedAsSta{
			WifiConnectedAsSta: &pb.TriggerWifiConnectedAsSta{},
		},
		Action: &pb.TriggerAction_Log{
			Log: &pb.ActionLog{},
		},
	}
	a[6] = logWifiSta

	logSSHLogin := &pb.TriggerAction{
		Trigger: &pb.TriggerAction_SshLogin{
			SshLogin: &pb.TriggerSSHLogin{},
		},
		Action: &pb.TriggerAction_Log{
			Log: &pb.ActionLog{},
		},
	}
	a[7] = logSSHLogin
	*/

}

type jsIsTriggerAction_Trigger interface {
	isTriggerAction_Trigger()
}
type jsIsTriggerAction_Action interface{ isTriggerAction_Action() }

type jsTriggerAction struct {
	*js.Object

	Id      uint32 `js:"Id"`
	OneShot bool `js:"OnShot"`

	Trigger *js.Object `js:"Trigger"`
	Action *js.Object `js:"Action"`
}

// TriggerAction_ServiceStarted is assignable to Trigger
type jsTriggerAction_ServiceStarted struct {
	*js.Object
	ServiceStarted *jsTriggerServiceStarted `js:"ServiceStarted"`
}

func isJsTriggerAction_ServiceStarted(src *js.Object) bool {
	test := jsTriggerAction_ServiceStarted{Object:src}
	if test.ServiceStarted.Object == js.Undefined { return false }
	return true
}

// TriggerAction_UsbGadgetConnected is assignable to Trigger
type jsTriggerAction_UsbGadgetConnected struct {
	*js.Object
	UsbGadgetConnected *jsTriggerUSBGadgetConnected `js:"UsbGadgetConnected"`
}

func isJsTriggerAction_UsbGadgetConnected(src *js.Object) bool {
	test := jsTriggerAction_UsbGadgetConnected{Object:src}
	if test.UsbGadgetConnected.Object == js.Undefined { return false }
	return true
}

// TriggerAction_UsbGadgetDisconnected is assignable to Trigger
type jsTriggerAction_UsbGadgetDisconnected struct {
	*js.Object
	UsbGadgetDisconnected *jsTriggerUSBGadgetDisconnected `js:"UsbGadgetDisconnected"`
}
func isJsTriggerAction_UsbGadgetDisconnected(src *js.Object) bool {
	test := jsTriggerAction_UsbGadgetDisconnected{Object:src}
	if test.UsbGadgetDisconnected.Object == js.Undefined { return false }
	return true
}

// TriggerAction_WifiAPStarted is assignable to Trigger
type jsTriggerAction_WifiAPStarted struct {
	*js.Object
	WifiAPStarted *jsTriggerWifiAPStarted `js:"WifiAPStarted"`
}
func iJsTriggerAction_WifiAPStarted(src *js.Object) bool {
	test := jsTriggerAction_WifiAPStarted{Object:src}
	if test.WifiAPStarted.Object == js.Undefined { return false }
	return true
}

// TriggerAction_WifiConnectedAsSta is assignable to Trigger
type jsTriggerAction_WifiConnectedAsSta struct {
	*js.Object
	WifiConnectedAsSta *jsTriggerWifiConnectedAsSta `js:"WifiConnectedAsSta"`
}
func isJsTriggerAction_WifiConnectedAsSta(src *js.Object) bool {
	test := jsTriggerAction_WifiConnectedAsSta{Object:src}
	if test.WifiConnectedAsSta.Object == js.Undefined { return false }
	return true
}

// TriggerAction_SshLogin is assignable to Trigger
type jsTriggerAction_SshLogin struct {
	*js.Object
	SshLogin *jsTriggerSSHLogin `js:"SshLogin"`
}
func isJsTriggerAction_SshLogin(src *js.Object) bool {
	test := jsTriggerAction_SshLogin{Object:src}
	if test.SshLogin.Object == js.Undefined { return false }
	return true
}

// TriggerAction_DhcpLeaseGranted is assignable to Trigger
type jsTriggerAction_DhcpLeaseGranted struct {
	*js.Object
	DhcpLeaseGranted *jsTriggerDHCPLeaseGranted `js:"DhcpLeaseGranted"`
}
func isJsTriggerAction_DhcpLeaseGranted(src *js.Object) bool {
	test := jsTriggerAction_DhcpLeaseGranted{Object:src}
	if test.DhcpLeaseGranted.Object == js.Undefined { return false }
	return true
}

// TriggerAction_BashScript is assignable to Action
type jsTriggerAction_BashScript struct {
	*js.Object
	BashScript *jsActionStartBashScript `js:"BashScript"`
}
func isJsTriggerAction_BashScript(src *js.Object) bool {
	test := jsTriggerAction_BashScript{Object:src}
	if test.BashScript.Object == js.Undefined { return false }
	return true
}

// TriggerAction_HidScript is assignable to Action
type jsTriggerAction_HidScript struct {
	*js.Object
	HidScript *jsActionStartHIDScript `js:"HidScript"`
}
func isJsTriggerAction_HidScript(src *js.Object) bool {
	test := jsTriggerAction_HidScript{Object:src}
	if test.HidScript.Object == js.Undefined { return false }
	return true
}

// TriggerAction_DeploySettingsTemplate is assignable to Action
type jsTriggerAction_DeploySettingsTemplate struct {
	*js.Object
	DeploySettingsTemplate *jsActionDeploySettingsTemplate `js:"DeploySettingsTemplate"`
}
func isJsTriggerAction_DeploySettingsTemplate(src *js.Object) bool {
	test := jsTriggerAction_DeploySettingsTemplate{Object:src}
	if test.DeploySettingsTemplate.Object == js.Undefined { return false }
	return true
}

// TriggerAction_Log is assignable to Action
type jsTriggerAction_Log struct {
	*js.Object
	Log *jsActionLog `js:"Log"`
}
func isJsTriggerAction_Log(src *js.Object) bool {
	test := jsTriggerAction_Log{Object:src}
	if test.Log.Object == js.Undefined { return false }
	return true
}

func (*jsTriggerAction_ServiceStarted) isTriggerAction_Trigger()        {}
func (*jsTriggerAction_UsbGadgetConnected) isTriggerAction_Trigger()    {}
func (*jsTriggerAction_UsbGadgetDisconnected) isTriggerAction_Trigger() {}
func (*jsTriggerAction_WifiAPStarted) isTriggerAction_Trigger()         {}
func (*jsTriggerAction_WifiConnectedAsSta) isTriggerAction_Trigger()    {}
func (*jsTriggerAction_SshLogin) isTriggerAction_Trigger()              {}
func (*jsTriggerAction_DhcpLeaseGranted) isTriggerAction_Trigger()      {}
func (*jsTriggerAction_BashScript) isTriggerAction_Action()             {}
func (*jsTriggerAction_HidScript) isTriggerAction_Action()              {}
func (*jsTriggerAction_DeploySettingsTemplate) isTriggerAction_Action() {}
func (*jsTriggerAction_Log) isTriggerAction_Action()                    {}

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
	ResInterface string `js:"ResInterface"`
	ResClientIP  string `js:"ResClientIP"`
	ResClientMac string `js:"ResClientMac"`
}

type jsActionStartBashScript struct {
	*js.Object
	ScriptPath string `js:"ScriptPath"`
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
type jsActionLog struct {
	*js.Object
}



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

	ActionLog = actionType(0)
	ActionHidScript = actionType(1)
	ActionDeploySettingsTemplate = actionType(2)
	ActionBashScript = actionType(3)
)
var triggerNames = map[triggerType]string{
	TriggerServiceStarted: "Service started",
	TriggerUsbGadgetConnected: "USB Gadget connected to host",
	TriggerUsbGadgetDisconnected: "USB Gadget disconnected from host",
	TriggerWifiAPStarted: "WiFi Access Point is up",
	TriggerWifiConnectedAsSta: "Connected to existing WiFi",
	TriggerSshLogin: "User login via SSH",
	TriggerDhcpLeaseGranted: "Client received DHCP lease",
}
var actionNames = map[actionType]string{
	ActionLog: "Log to internal console",
	ActionHidScript: "Start a HIDScript",
	ActionDeploySettingsTemplate: "Load and deploy the given settings",
	ActionBashScript: "Run the given bash script",
}
var availableTriggers = []triggerType{
	TriggerServiceStarted,
	TriggerUsbGadgetConnected,
	TriggerUsbGadgetDisconnected,
	TriggerWifiAPStarted,
	TriggerWifiConnectedAsSta,
	TriggerDhcpLeaseGranted,
}
var availableActions = []actionType {
	ActionLog,
	ActionBashScript,
}
func generateSelectOptionsTrigger() *js.Object {
	tts := js.Global.Get("Array").New()
	type option struct {
		*js.Object
		Label string `js:"label"`
		Value triggerType `js:"value"`
	}
	for _,triggerVal := range availableTriggers {
		triggerLabel := triggerNames[triggerVal]
		o := option{Object:O()}
		o.Value = triggerVal
		o.Label = triggerLabel
		tts.Call("push", o)
	}
	return tts
}

func generateSelectOptionsAction() *js.Object {
	tts := js.Global.Get("Array").New()
	type option struct {
		*js.Object
		Label string `js:"label"`
		Value actionType `js:"value"`
	}
	for _, actionVal := range availableActions {
		actionLabel := actionNames[actionVal]
		o := option{Object:O()}
		o.Value = actionVal
		o.Label = actionLabel
		tts.Call("push", o)
	}
	return tts
}

type jsVMTriggerAction struct {
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

func NewTriggerAction() *jsVMTriggerAction {
	ta := &jsVMTriggerAction{Object:O()}
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

func InitComponentsTriggerActions() {
	// ToDo: delete test
	ExportDefaultTriggerActions()

	hvue.NewComponent(
		"triggeraction-manager",
		hvue.Template(templateTriggerActionManager),
		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := struct {
				*js.Object
				TriggerAction *jsVMTriggerAction `js:"TriggerAction"`
			}{Object: O()}
			data.TriggerAction = NewTriggerAction()
			return &data
		}),
		hvue.Mounted(func(vm *hvue.VM) {
			vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_TRIGGER_ACTIONS)
		}),
	)

	hvue.NewComponent(
		"triggeraction",
		hvue.Template(templateTriggerAction),
		hvue.Props("ta"),
	)
	hvue.NewComponent(
		"trigger",
		hvue.Props("ta"),
		hvue.Template(templateTrigger),
		hvue.Computed("triggertypes", func(vm *hvue.VM) interface{} {
			return generateSelectOptionsTrigger()
		}),
	)
	hvue.NewComponent(
		"action",
		hvue.Props("ta"),
		hvue.Template(templateAction),
		hvue.Computed("actiontypes", func(vm *hvue.VM) interface{} {
			return generateSelectOptionsAction()
		}),
		hvue.ComputedWithGetSet(
			"actionType",
			func(vm *hvue.VM) interface{} {
				return vm.Get("ta").Get("ActionType")
			},
			func(vm *hvue.VM, newValue *js.Object) {
				// set ActionData accordingly
				var data *js.Object
				switch aType := actionType(newValue.Int()); aType {
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
					d.ScriptPath = "/path/to/some/script"
					data = d.Object
				default:
					println("Unknown action type")
					data = O()
				}
				vm.Get("ta").Set("ActionData", data)

				// set type itself (after data)
				vm.Get("ta").Set("ActionType", newValue)
			}),
		hvue.Computed("isActionLog", func(vm *hvue.VM) interface{} {
			if at := actionType(vm.Get("actionType").Int()); at == ActionLog {
				return true
			} else {
				return false
			}
		}),
		hvue.Computed("isActionHidScript", func(vm *hvue.VM) interface{} {
			if at := actionType(vm.Get("actionType").Int()); at == ActionHidScript {
				return true
			} else {
				return false
			}
		}),
		hvue.Computed("isActionDeploySettingsTemplate", func(vm *hvue.VM) interface{} {
			if at := actionType(vm.Get("actionType").Int()); at == ActionDeploySettingsTemplate {
				return true
			} else {
				return false
			}
		}),
		hvue.Computed("isActionBashScript", func(vm *hvue.VM) interface{} {
			if at := actionType(vm.Get("actionType").Int()); at == ActionBashScript {
				return true
			} else {
				return false
			}
		}),
	)
}

//
const templateTriggerAction = `
<q-card class="fit">
<!-- {{ ta }} -->
	<q-card-title>Triggered Action (ID {{ ta.Id }})</q-card-title>
	<q-list>
			<q-item tag="label" link>
				<q-item-side>
					<q-toggle v-model="ta.IsActive"></q-toggle>
				</q-item-side>
				<q-item-main>
					<q-item-tile label>Enabled</q-item-tile>
					<q-item-tile sublabel>If not enabled, the triggered action is ignored</q-item-tile>
				</q-item-main>
			</q-item>

			<q-item tag="label" link :disabled="!ta.IsActive">
				<q-item-side>
					<q-toggle v-model="ta.OneShot" :disable="!ta.IsActive"></q-toggle>
				</q-item-side>
				<q-item-main>
					<q-item-tile label>One shot</q-item-tile>
					<q-item-tile sublabel>The trigger fires every time the respective event occurs. If "one shot" is enabled it fores only once.</q-item-tile>
				</q-item-main>
			</q-item>
	</q-list>

						<div class="row items-stretch">
							<div class="col-12 col-md-6"">
								<trigger :ta="ta"></trigger>
							</div>
	
							<div class="col-12 col-md-6">
								<action :ta="ta"></action>
							</div>
						</div>


</q-card>
`
const templateTrigger = `
		<q-list class="fit" no-border link :disabled="!ta.IsActive">
			<q-item tag="label">
				<q-item-main>
					<q-item-tile label>Trigger</q-item-tile>
					<q-item-tile sublabel>Chose the event which has to occur to start the selected action</q-item-tile>
					<q-item-tile>
						<q-select v-model="ta.TriggerType" :options="triggertypes" inverted :disable="!ta.IsActive"></q-select>
					</q-item-tile>
				</q-item-main>
			</q-item>
		</q-list>
`

const templateAction = `
		<q-list class="fit" no-border link :disabled="!ta.IsActive">
			<q-item tag="label">
				<q-item-main>
					<q-item-tile label>Action</q-item-tile>
					<q-item-tile sublabel>Chose the action which should be started when the trigger fired</q-item-tile>
					<q-item-tile>
						<q-select v-model="actionType" :options="actiontypes" color="secondary" inverted :disable="!ta.IsActive"></q-select>
					</q-item-tile>
				</q-item-main>
			</q-item>

			<q-item tag="label" v-if="isActionBashScript">
				<q-item-main>
					<q-item-tile label>Script path</q-item-tile>
					<q-item-tile sublabel>Path to the BashScript which should be issued</q-item-tile>
					<q-item-tile>
						<q-input v-model="ta.ActionData.ScriptPath" :options="actiontypes" color="secondary" inverted :disable="!ta.IsActive"></q-input>
					</q-item-tile>
				</q-item-main>
			</q-item>
		</q-list>
`



const templateTriggerActionManager = `
<q-page padding>
<div class="row gutter-sm">
	<div class="col-12 col-xl-6" v-for="ta in $store.getters.triggerActions"> 
		<triggeraction :key="ta.Id" :ta="ta"></triggeraction>
	</div>
</div>
</q-page>	

`
