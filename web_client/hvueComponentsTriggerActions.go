// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/mame82/hvue"
	"strconv"
)

func ExportDefaultTriggerActions() {
	// create test trigger

	/*
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

	*/
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

/*
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

*/





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

func generateSelectOptionsGPIOOutValue() *js.Object {
	tts := js.Global.Get("Array").New()
	type option struct {
		*js.Object
		Label string `js:"label"`
		Value GPIOOutValue `js:"value"`
	}

	for _, value := range availableGPIOOutValues {
		label := gpioOutValueNames[value]
		o := option{Object:O()}
		o.Value = value
		o.Label = label
		tts.Call("push", o)
	}
	return tts
}

func generateSelectOptionsGPIOInPullUpDown() *js.Object {
	tts := js.Global.Get("Array").New()
	type option struct {
		*js.Object
		Label string `js:"label"`
		Value GPIOInPullUpDown `js:"value"`
	}

	for _, value := range availableGPIOInPullUpDowns {
		label := gpioInPullUpDownNames[value]
		o := option{Object:O()}
		o.Value = value
		o.Label = label
		tts.Call("push", o)
	}
	return tts
}

func generateSelectOptionsGPIOInEdges() *js.Object {
	tts := js.Global.Get("Array").New()
	type option struct {
		*js.Object
		Label string `js:"label"`
		Value GPIOInEdge `js:"value"`
	}

	for _, value := range availableGPIOInEdges {
		label := gpioInEdgeNames[value]
		o := option{Object:O()}
		o.Value = value
		o.Label = label
		tts.Call("push", o)
	}
	return tts
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
				TriggerAction *jsTriggerAction `js:"TriggerAction"`
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
		hvue.Computed("pullupdown", func(vm *hvue.VM) interface{} {
			return generateSelectOptionsGPIOInPullUpDown()
		}),
		hvue.Computed("edge", func(vm *hvue.VM) interface{} {
			return generateSelectOptionsGPIOInEdges()
		}),
		hvue.ComputedWithGetSet(
			"triggerType",
			func(vm *hvue.VM) interface{} {
				return vm.Get("ta").Get("TriggerType")
			},
			func(vm *hvue.VM, newValue *js.Object) {
				tType := triggerType(newValue.Int())
				ta := &jsTriggerAction{Object: vm.Get("ta")}
				ta.ChangeTriggerType(tType)
			}),
		hvue.Method(
			"TriggerGroupReceiveSequenceAddValue",
			func(vm *hvue.VM, newVal *js.Object) {
				println("Force add", newVal)
				ta := &jsTriggerAction{Object: vm.Get("ta")}
				if !ta.IsTriggerGroupReceiveSequence() { return }

				// cast data Object to jsTriggerGroupReceiveSequence
				tgrs := &jsTriggerGroupReceiveSequence{Object:ta.TriggerData}
				strVal := newVal.String()
				if intVal,errconv := strconv.Atoi(strVal); errconv == nil {
					//append to Values
					tgrs.ValueSequence = append(tgrs.ValueSequence, intVal)
				}
			}),
		hvue.ComputedWithGetSet(
			"TriggerGroupReceiveSequenceValues",
			func(vm *hvue.VM) interface{} {
				ta := &jsTriggerAction{Object: vm.Get("ta")}
				if !ta.IsTriggerGroupReceiveSequence() { return []string{} }

				// cast data Object to jsTriggerGroupReceiveSequence
				tgrs := &jsTriggerGroupReceiveSequence{Object:ta.TriggerData}

				res := make([]string, len(tgrs.ValueSequence))
				for idx,intVal := range tgrs.ValueSequence {
					res[idx] = strconv.Itoa(intVal)
				}
				return res
			},
			func(vm *hvue.VM, newValue *js.Object) {
				ta := &jsTriggerAction{Object: vm.Get("ta")}
				if !ta.IsTriggerGroupReceiveSequence() { return }

				// cast data Object to jsTriggerGroupReceiveSequence
				tgrs := &jsTriggerGroupReceiveSequence{Object:ta.TriggerData}

				// clear old array
				tgrs.ValueSequence = []int{}

				// iterate over newValue, which is assumed to be an Array of strings
				for idx := 0; idx < newValue.Length(); idx++ {
					//fetch value
					strVal := newValue.Index(idx).String()
					// try to cast to int
					if intVal,errconv := strconv.Atoi(strVal); errconv == nil {
						//append to Values
						tgrs.ValueSequence = append(tgrs.ValueSequence, intVal)
					}
				}
			}),
		hvue.Computed("isTriggerGPIOIn", func(vm *hvue.VM) interface{} {
			return (&jsTriggerAction{Object: vm.Get("ta")}).IsTriggerGPIOIn()
		}),
		hvue.Computed("isTriggerGroupReceive", func(vm *hvue.VM) interface{} {
			return (&jsTriggerAction{Object: vm.Get("ta")}).IsTriggerGroupReceive()
		}),
		hvue.Computed("isTriggerGroupReceiveSequence", func(vm *hvue.VM) interface{} {
			return (&jsTriggerAction{Object: vm.Get("ta")}).IsTriggerGroupReceiveSequence()
		}),
	)
	hvue.NewComponent(
		"action",
		hvue.Props("ta"),
		hvue.Template(templateAction),
		hvue.Computed("actiontypes", func(vm *hvue.VM) interface{} {
			return generateSelectOptionsAction()
		}),
		hvue.Computed("gpiooutvalues", func(vm *hvue.VM) interface{} {
			return generateSelectOptionsGPIOOutValue()
		}),
		hvue.ComputedWithGetSet(
			"actionType",
			func(vm *hvue.VM) interface{} {
				return vm.Get("ta").Get("ActionType")
			},
			func(vm *hvue.VM, newValue *js.Object) {
				aType := actionType(newValue.Int())
				ta := &jsTriggerAction{Object: vm.Get("ta")}
				ta.ChangeActionType(aType)
			}),
		hvue.Computed("isActionLog", func(vm *hvue.VM) interface{} {
			return (&jsTriggerAction{Object: vm.Get("ta")}).IsActionLog()
		}),
		hvue.Computed("isActionHidScript", func(vm *hvue.VM) interface{} {
			return (&jsTriggerAction{Object: vm.Get("ta")}).IsActionHidScript()
		}),
		hvue.Computed("isActionDeploySettingsTemplate", func(vm *hvue.VM) interface{} {
			return (&jsTriggerAction{Object: vm.Get("ta")}).IsActionDeploySettingsTemplate()
		}),
		hvue.Computed("isActionBashScript", func(vm *hvue.VM) interface{} {
			return (&jsTriggerAction{Object: vm.Get("ta")}).IsActionBashScript()
		}),
		hvue.Computed("isActionGPIOOut", func(vm *hvue.VM) interface{} {
			return (&jsTriggerAction{Object: vm.Get("ta")}).IsActionGPIOOut()
		}),
		hvue.Computed("isActionGroupSend", func(vm *hvue.VM) interface{} {
			return (&jsTriggerAction{Object: vm.Get("ta")}).IsActionGroupSend()
		}),
	)
}

//
const templateTriggerAction = `
<q-card class="fit">
<!-- {{ ta }} -->
	<q-card-title>TriggereAction (ID {{ ta.Id }})</q-card-title>
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
						<q-select v-model="triggerType" :options="triggertypes" inverted :disable="!ta.IsActive"></q-select>
					</q-item-tile>
				</q-item-main>
			</q-item>

			<q-item tag="label" v-if="isTriggerGroupReceive || isTriggerGroupReceiveSequence">
				<q-item-main>
					<q-item-tile label>Group name</q-item-tile>
					<q-item-tile sublabel>Only values send for this group name are regarded</q-item-tile>
					<q-item-tile>
						<q-input v-model="ta.TriggerData.GroupName" inverted :disable="!ta.IsActive"></q-input>
					</q-item-tile>
				</q-item-main>
			</q-item>
			<q-item tag="label" v-if="isTriggerGroupReceive">
				<q-item-main>
					<q-item-tile label>Value</q-item-tile>
					<q-item-tile sublabel>The numeric value which has to be received to activate the trigger</q-item-tile>
					<q-item-tile>
						<q-input v-model="ta.TriggerData.Value" type="number" decimals="0" inverted :disable="!ta.IsActive"></q-input>
					</q-item-tile>
				</q-item-main>
			</q-item>
			<q-item tag="label" v-if="isTriggerGroupReceiveSequence">
				<q-item-main>
					<q-item-tile label>Values</q-item-tile>
					<q-item-tile sublabel>The numeric value which has to be received to activate the trigger</q-item-tile>
					<q-item-tile>
						<q-chips-input v-model="TriggerGroupReceiveSequenceValues" @duplicate="TriggerGroupReceiveSequenceAddValue($event)" decimals="0" inverted :disable="!ta.IsActive"></q-chips-input>
					</q-item-tile>
				</q-item-main>
			</q-item>
			<q-item tag="label" link :disabled="!ta.IsActive" v-if="isTriggerGroupReceiveSequence">
				<q-item-side>
					<q-toggle v-model="ta.TriggerData.IgnoreOutOfOrder" :disable="!ta.IsActive"></q-toggle>
				</q-item-side>
				<q-item-main>
					<q-item-tile label>Ignore out-of-order values</q-item-tile>
					<q-item-tile sublabel>If enabled, the trigger fires even if other values arrive in between the ones of the sequence. If disabled, no other values are allowed in between the sequence</q-item-tile>
				</q-item-main>
			</q-item>


			<q-item tag="label" v-if="isTriggerGPIOIn">
				<q-item-main>
					<q-item-tile label>GPIO Number</q-item-tile>
					<q-item-tile sublabel>The number of the GPIO to monitor</q-item-tile>
					<q-item-tile>
						<q-input v-model="ta.TriggerData.GpioNum" type="number" decimals="0" inverted :disable="!ta.IsActive"></q-input>
					</q-item-tile>
				</q-item-main>
			</q-item>
			<q-item tag="label" v-if="isTriggerGPIOIn">
				<q-item-main>
					<q-item-tile label>Pull resistor</q-item-tile>
					<q-item-tile sublabel>Chose if internal Pull-up/down resistor should be used</q-item-tile>
					<q-item-tile>
						<q-select v-model="ta.TriggerData.PullUpDown" :options="pullupdown" inverted :disable="!ta.IsActive"></q-select>
					</q-item-tile>
				</q-item-main>
			</q-item>
			<q-item tag="label" v-if="isTriggerGPIOIn">
				<q-item-main>
					<q-item-tile label>Edge</q-item-tile>
					<q-item-tile sublabel>What edge (level change) has to occur to fire the trigger</q-item-tile>
					<q-item-tile>
						<q-select v-model="ta.TriggerData.Edge" :options="edge" inverted :disable="!ta.IsActive"></q-select>
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
						<q-input v-model="ta.ActionData.ScriptName" color="secondary" inverted :disable="!ta.IsActive"></q-input>
					</q-item-tile>
				</q-item-main>
			</q-item>

			<q-item tag="label" v-if="isActionHidScript">
				<q-item-main>
					<q-item-tile label>Script name</q-item-tile>
					<q-item-tile sublabel>Name of a stored HIDScript</q-item-tile>
					<q-item-tile>
						<q-input v-model="ta.ActionData.ScriptName" color="secondary" inverted :disable="!ta.IsActive"></q-input>
					</q-item-tile>
				</q-item-main>
			</q-item>

			<q-item tag="label" v-if="isActionGPIOOut">
				<q-item-main>
					<q-item-tile label>GPIO Number</q-item-tile>
					<q-item-tile sublabel>The number of the GPIO to output on</q-item-tile>
					<q-item-tile>
						<q-input v-model="ta.ActionData.GpioNum" type="number" decimals="0" color="secondary" inverted :disable="!ta.IsActive"></q-input>
					</q-item-tile>
				</q-item-main>
			</q-item>
			<q-item tag="label" v-if="isActionGPIOOut">
				<q-item-main>
					<q-item-tile label>Output</q-item-tile>
					<q-item-tile sublabel>Output low/high on the given GPIO or toggle the output</q-item-tile>
					<q-item-tile>
						<q-select v-model="ta.ActionData.Value" :options="gpiooutvalues" color="secondary" inverted :disable="!ta.IsActive"></q-select>
					</q-item-tile>
				</q-item-main>
			</q-item>


			<q-item tag="label" v-if="isActionGroupSend">
				<q-item-main>
					<q-item-tile label>Group name</q-item-tile>
					<q-item-tile sublabel>The name of the group to send to (has to match respective listeners)</q-item-tile>
					<q-item-tile>
						<q-input v-model="ta.ActionData.GroupName" color="secondary" inverted :disable="!ta.IsActive"></q-input>
					</q-item-tile>
				</q-item-main>
			</q-item>
			<q-item tag="label" v-if="isActionGroupSend">
				<q-item-main>
					<q-item-tile label>Value</q-item-tile>
					<q-item-tile sublabel>The numeric value which is sent to the group channel</q-item-tile>
					<q-item-tile>
						<q-input v-model="ta.ActionData.Value" color="secondary" type="number" decimals="0" inverted :disable="!ta.IsActive"></q-input>
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
