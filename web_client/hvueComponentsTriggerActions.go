// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/mame82/hvue"
	"strconv"
)

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

func generateSelectOptionsGPIONum() *js.Object {
	tts := js.Global.Get("Array").New()
	type option struct {
		*js.Object
		Label string `js:"label"`
		Value GPIONum `js:"value"`
	}

	for _, value := range availableGPIONums {
		label := gpioNumNames[value]
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

func generateSelectOptionsTemplateTypes() *js.Object {
	tts := js.Global.Get("Array").New()
	type option struct {
		*js.Object
		Label string `js:"label"`
		Value TemplateType `js:"value"`
	}

	for _, value := range availableTemplateTypes {
		label := templateTypeNames[value]
		o := option{Object:O()}
		o.Value = value
		o.Label = label
		tts.Call("push", o)
	}
	return tts
}

type TriggerActionCompData struct {
	*js.Object
	EditMode bool `js:"EditMode"`
}



func InitComponentsTriggerActions() {
	hvue.NewComponent(
		"triggeraction-manager",
		hvue.Template(templateTriggerActionManager),
		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := struct {
				*js.Object
				ShowReplaceTASModal bool   `js:"showReplaceTASModal"`
				ShowAddTASModal bool   `js:"showAddTASModal"`
				ShowStoreTASModal bool   `js:"showStoreTASModal"`
				TemplateName   string `js:"templateName"`
			}{Object: O()}
			data.ShowReplaceTASModal = false
			data.ShowAddTASModal = false
			data.ShowStoreTASModal = false
			data.TemplateName = ""
			return &data
		}),
		hvue.Method("addTA",
			func(vm *hvue.VM) {
				vm.Get("$store").Call("dispatch", VUEX_ACTION_ADD_NEW_TRIGGER_ACTION)
			}),
		hvue.Method("storeTAS",
			func(vm *hvue.VM, name *js.Object) {
				tas_obj := vm.Get("$store").Get("state").Get("triggerActionList")
				current_tas := jsTriggerActionSet{Object:tas_obj}.toGo()
				store_tas := NewTriggerActionSet()
				store_tas.Name = name.String()
				for _,ta := range current_tas.TriggerActions {

					if ta.IsActive && !ta.Immutable {
						jsTa := &jsTriggerAction{Object:O()}
						jsTa.fromGo(ta)
						store_tas.UpdateEntry(jsTa)
					}
				}

				vm.Get("$store").Call("dispatch", VUEX_ACTION_STORE_TRIGGER_ACTION_SET, store_tas)
			}),
		hvue.Method("replaceCurrentTAS",
			func(vm *hvue.VM, storedTASName *js.Object) {
				//vm.Get("$q").Call("notify", "Replacing TAS with '" + storedTASName.String() +"'")
				vm.Get("$store").Call("dispatch", VUEX_ACTION_DEPLOY_STORED_TRIGGER_ACTION_SET_REPLACE, storedTASName)
			}),
		hvue.Method("addToCurrentTAS",
			func(vm *hvue.VM, storedTASName *js.Object) {
				//vm.Get("$q").Call("notify", "Add '" + storedTASName.String() +"' to current TAS")
				vm.Get("$store").Call("dispatch", VUEX_ACTION_DEPLOY_STORED_TRIGGER_ACTION_SET_ADD, storedTASName)
			}),
		hvue.Method("updateStoredTriggerActionSetsList",
			func(vm *hvue.VM) {
				vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_STORED_TRIGGER_ACTION_SETS_LIST)
			}),
		hvue.Mounted(func(vm *hvue.VM) {
			vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_CURRENT_TRIGGER_ACTIONS_FROM_SERVER)
		}),
	)



	hvue.NewComponent(
		"TriggerAction",
		hvue.Template(templateTriggerAction),
		hvue.PropObj("ta"),
	)

	hvue.NewComponent(
		"TriggerActionOverview",
		hvue.Template(templateTriggerActionOverview),
		hvue.PropObj("ta"),
		hvue.Computed("computedColor", func(vm *hvue.VM) interface{} {
			ta := &jsTriggerAction{Object: vm.Get("ta")}
			switch {
			case ta.Immutable:
				return "dark"
			/*
			case !ta.IsActive:
				return "light"
			*/
			default:
				return ""
			}

		}),
		hvue.Computed("strTrigger", func(vm *hvue.VM) interface{} {
			ta := &jsTriggerAction{Object: vm.Get("ta")}
			strTrigger := triggerNames[ta.TriggerType]

			switch {
			case ta.IsTriggerGroupReceive():
				t := jsTriggerGroupReceive{Object: ta.TriggerData}
				strTrigger += " ("
				strTrigger += t.GroupName
				strTrigger += ": " + strconv.Itoa(int(t.Value))
				strTrigger += ")"
			case ta.IsTriggerGroupReceiveSequence():
				t := jsTriggerGroupReceiveSequence{Object: ta.TriggerData}
				strTrigger += " ("
				strTrigger += t.GroupName
				strTrigger += ": ["
				for idx,val  := range t.ValueSequence {
					if idx != 0 {
						strTrigger += ", "
					}
					strTrigger += strconv.Itoa(int(val))
				}
				strTrigger += "]"
				if !t.IgnoreOutOfOrder {
					strTrigger += ", out-of-order allowed"
				}
				strTrigger += ")"
			case ta.IsTriggerGPIOIn():
				t := jsTriggerGPIOIn{Object: ta.TriggerData}
				strTrigger += " ("
				strTrigger += gpioNumNames[t.GpioNum]
				strTrigger += ": " + gpioInEdgeNames[t.Edge]
				strTrigger += ", resistor: " + gpioInPullUpDownNames[t.PullUpDown]
				strTrigger += ")"
			}

			return strTrigger
		}),
		hvue.Computed("strAction", func(vm *hvue.VM) interface{} {
			ta := &jsTriggerAction{Object: vm.Get("ta")}
			strAction := actionNames[ta.ActionType]
			switch {
			case ta.IsActionGroupSend():
				tgs := jsActionGroupSend{Object: ta.ActionData}
				strAction += " ("
				strAction += tgs.GroupName
				strAction += ": " + strconv.Itoa(int(tgs.Value))
				strAction += ")"
			case ta.IsActionGPIOOut():
				a := jsActionGPIOOut{Object: ta.ActionData}
				strAction += " ("
				strAction += gpioNumNames[a.GpioNum]
				strAction += ": " + gpioOutValueNames[a.Value]
				strAction += ")"
			case ta.IsActionBashScript():
				a := jsActionStartBashScript{Object: ta.ActionData}
				strAction += " ('"
				strAction += a.ScriptName
				strAction += "')"
			case ta.IsActionDeploySettingsTemplate():
				a := jsActionDeploySettingsTemplate{Object: ta.ActionData}
				strAction += " ("
				strAction += templateTypeNames[a.Type]
				strAction += ": '" + a.TemplateName
				strAction += "')"
			case ta.IsActionHidScript():
				a := jsActionStartHIDScript{Object: ta.ActionData}
				strAction += " ('"
				strAction += a.ScriptName
				strAction += "')"
			}
			return strAction
		}),
		hvue.Mounted(func(vm *hvue.VM) {
			data := TriggerActionCompData{Object: vm.Data}
			data.EditMode = vm.Get("overview").Bool()
		}),
		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := &TriggerActionCompData{Object: O()}
			data.EditMode = false

			return data
		}),
		hvue.Method(
			"enableEditMode",
			func(vm *hvue.VM) {
				data := TriggerActionCompData{Object: vm.Data}
				data.EditMode = true
			}),
		hvue.Method(
			"updateTA",
			func(vm *hvue.VM) {
				println("update ta: ", vm.Get("ta"))
				//Replace the whole TriggerActionSet of server with the current one from vuex store
				// ToDo: This has to be changed to update a single action (inconssistnecy with multiple clients, all TA IDs change, overhead of transferring a whole set) -> has to be implemented like deleteTA logic
				currentTas := vm.Get("$store").Get("state").Get("triggerActionList") //Current TriggerActionSet of ViewModel
				vm.Get("$store").Call("dispatch", VUEX_ACTION_DEPLOY_TRIGGER_ACTION_SET_REPLACE, currentTas)
			}),
		hvue.Method(
			"cancelUpdateTA",
			func(vm *hvue.VM) {
				println("cancel update ta: ", vm.Get("ta"))
				//Reload the whole TriggerActionSet from server and overwrite the current one of vuex store
				vm.Get("$store").Call("dispatch", VUEX_ACTION_UPDATE_CURRENT_TRIGGER_ACTIONS_FROM_SERVER)
			}),
		hvue.Method(
			"deleteTA",
			func(vm *hvue.VM) {
				ta_obj := vm.Get("ta")
				println("delete ta: ", ta_obj)

				delTas := NewTriggerActionSet()
				delTas.UpdateEntry(&jsTriggerAction{Object: ta_obj})
				vm.Get("$store").Call("dispatch", VUEX_ACTION_REMOVE_TRIGGER_ACTIONS, delTas)
			}),
	)

	hvue.NewComponent(
		"TriggerActionEdit",
		hvue.Template(templateTriggerActionEdit),
		hvue.PropObj("ta"),
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
		hvue.Computed("gpionum", func(vm *hvue.VM) interface{} {
			return generateSelectOptionsGPIONum()
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
					tgrs.ValueSequence = append(tgrs.ValueSequence, int32(intVal))
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
					res[idx] = strconv.Itoa(int(intVal))
				}
				return res
			},
			func(vm *hvue.VM, newValue *js.Object) {
				ta := &jsTriggerAction{Object: vm.Get("ta")}
				if !ta.IsTriggerGroupReceiveSequence() { return }

				// cast data Object to jsTriggerGroupReceiveSequence
				tgrs := &jsTriggerGroupReceiveSequence{Object:ta.TriggerData}

				// clear old array
				tgrs.ValueSequence = []int32{}

				// iterate over newValue, which is assumed to be an Array of strings
				for idx := 0; idx < newValue.Length(); idx++ {
					//fetch value
					strVal := newValue.Index(idx).String()
					// try to cast to int
					if intVal,errconv := strconv.Atoi(strVal); errconv == nil {
						//append to Values
						tgrs.ValueSequence = append(tgrs.ValueSequence, int32(intVal))
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
		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := struct {
				*js.Object
				ShowSelectHIDScriptModal bool   `js:"ShowSelectHIDScriptModal"`
				ShowSelectBashScriptModal bool   `js:"ShowSelectBashScriptModal"`
				ShowSelectTemplateModal bool   `js:"ShowSelectTemplateModal"`
			}{Object: O()}
			data.ShowSelectHIDScriptModal = false
			data.ShowSelectBashScriptModal = false
			data.ShowSelectTemplateModal = false
			return &data
		}),
		hvue.Method("updateStoredHIDScriptsList",
			func(vm *hvue.VM) {
				vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_STORED_HID_SCRIPTS_LIST)
			}),
		hvue.Method("updateStoredBashScriptsList",
			func(vm *hvue.VM) {
				vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_STORED_BASH_SCRIPTS_LIST)
			}),
		hvue.Computed(
			"typedTemplateList",
			func(vm *hvue.VM) interface{} {
				// template type: ta.ActionData.Type
				ta:=&jsTriggerAction{Object: vm.Get("ta")}
				if !ta.IsActionDeploySettingsTemplate() {
					return []string{}
				}
				aData := &jsActionDeploySettingsTemplate{Object: ta.ActionData}
				switch aData.Type {
				case TemplateTypeFullSettings:
					//ToDo: Implement
				case TemplateTypeBluetooth:
					//ToDo: Implement
				case TemplateTypeUSB:
					//return USB list
					return vm.Store.Get("state").Get("StoredUSBSettingsList")
				case TemplateTypeTriggerActions:
					//return TriggerAction list
					return vm.Store.Get("state").Get("StoredTriggerActionSetsList")
				case TemplateTypeWifi:
					//return WiFi settings list
					return vm.Store.Get("state").Get("StoredWifiSettingsList")
				case TemplateTypeNetwork:
					//return ethernet interface settings list
					return vm.Store.Get("state").Get("StoredEthernetInterfaceSettingsList")
				}
				return []string{} //empty list
			}),
		hvue.Method(
			"actionTemplateTypeUpdate",
			func(vm *hvue.VM) interface{} {
				// template type: ta.ActionData.Type
				ta:=&jsTriggerAction{Object: vm.Get("ta")}
				if !ta.IsActionDeploySettingsTemplate() {
					return []string{}
				}
				aData := &jsActionDeploySettingsTemplate{Object: ta.ActionData}
				switch aData.Type {
				case TemplateTypeFullSettings:
					//ToDo: Implement
				case TemplateTypeBluetooth:
					//ToDo: Implement
				case TemplateTypeUSB:
					//update USB list
					vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_STORED_USB_SETTINGS_LIST)
				case TemplateTypeTriggerActions:
					//update TriggerAction list
					vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_STORED_TRIGGER_ACTION_SETS_LIST)
				case TemplateTypeWifi:
					//update WiFi settings template list
					vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_STORED_WIFI_SETTINGS_LIST)
				case TemplateTypeNetwork:
					//update ethernet interface settings template list
					vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_STORED_ETHERNET_INTERFACE_SETTINGS_LIST)
				}
				return []string{} //empty list
			}),
		hvue.Computed("actiontypes", func(vm *hvue.VM) interface{} {
			return generateSelectOptionsAction()
		}),
		hvue.Computed("gpiooutvalues", func(vm *hvue.VM) interface{} {
			return generateSelectOptionsGPIOOutValue()
		}),
		hvue.Computed("gpionum", func(vm *hvue.VM) interface{} {
			return generateSelectOptionsGPIONum()
		}),
		hvue.Computed("templatetypes", func(vm *hvue.VM) interface{} {
			return generateSelectOptionsTemplateTypes()
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

const templateTriggerAction = `
<div>
<!-- {{ ta }} -->
	<TriggerActionOverview :ta="ta"></TriggerActionOverview>
</div>
`
const templateTriggerActionOverview = `
<div>
<q-modal v-model="EditMode" no-route-dismiss no-esc-dismiss no-backdrop-dismiss>
	<TriggerActionEdit :ta="ta">
		<span slot="actions">
			<q-btn color="primary" @click="updateTA(); EditMode=false" label="update" />
			<q-btn color="secondary" @click="cancelUpdateTA(); EditMode=false" label="cancel" />
		</span>
	</TriggerActionEdit>
	
</q-modal>

<q-card tag="label" :color="computedColor" :text-color="ta.IsActive ? '': 'light'" :disabled="ta.Immutable" :dark="ta.Immutable">
	<q-card-title>
		{{ ta.Immutable ? "immutable, " : "" }}
		{{ ta.IsActive ? "enabled" : "disabled" }}
		TriggerAction (ID {{ ta.Id }})
	
		<span slot="subtitle">
			<q-icon name="input"></q-icon> 
			{{ strTrigger }}
			<br><q-icon name="launch"></q-icon>
			{{ strAction }}{{ta.OneShot ? " only once" : "" }}	
		</span>

		<div slot="right" v-if="!ta.Immutable">
			<q-btn color="primary" icon="edit" @click="enableEditMode" flat></q-btn>
			<q-btn color="negative" icon="delete" @click="deleteTA" flat></q-btn>
		</div>
	</q-card-title>
</q-card>
</div>
`

const templateTriggerActionEdit = `
<q-card class="fit">
	<q-card-title>
		TriggerAction
		<span slot="subtitle">ID {{ ta.Id }}</span>
		<!-- <q-btn slot="right" icon="more_vert" flat></q-btn> -->
	</q-card-title>
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
					<q-item-tile sublabel>The trigger fires every time the respective event occurs. If "one shot" is enabled it fires only once.</q-item-tile>
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

	<q-card-actions>
		<slot name="actions"></slot>
	</q-card-actions>
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
					<q-item-tile sublabel>The numeric value sequence which has to be received to activate the trigger</q-item-tile>
					<q-item-tile>
						<q-chips-input v-model="TriggerGroupReceiveSequenceValues" @duplicate="TriggerGroupReceiveSequenceAddValue($event)" type="number" decimals="0" inverted :disable="!ta.IsActive"></q-chips-input>
					</q-item-tile>
				</q-item-main>
			</q-item>
			<q-item tag="label" link :disabled="!ta.IsActive" v-if="isTriggerGroupReceiveSequence">
				<q-item-side>
					<q-toggle v-model="ta.TriggerData.IgnoreOutOfOrder" :disable="!ta.IsActive"></q-toggle>
				</q-item-side>
				<q-item-main>
					<q-item-tile label>Ignore out-of-order values</q-item-tile>
					<q-item-tile sublabel>If enabled the sequence may be interrupted by other values. If disabled they have to arrive in exact order.</q-item-tile>
				</q-item-main>
			</q-item>


			<q-item tag="label" v-if="isTriggerGPIOIn">
				<q-item-main>
					<q-item-tile label>GPIO Number</q-item-tile>
					<q-item-tile sublabel>The number of the GPIO to monitor</q-item-tile>
					<q-item-tile>
						<q-select v-model="ta.TriggerData.GpioNum" :options="gpionum" inverted :disable="!ta.IsActive"></q-select>
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
<select-string-from-array :values="$store.state.StoredBashScriptsList" v-model="ShowSelectBashScriptModal" title="Select BASH script" @load="ta.ActionData.ScriptName=$event"></select-string-from-array>
				<q-item-main>
					<q-item-tile label>Script path</q-item-tile>
					<q-item-tile sublabel>Path to the BashScript which should be issued</q-item-tile>
					<q-item-tile>
<q-input @click="updateStoredBashScriptsList();ShowSelectBashScriptModal=true" v-model="ta.ActionData.ScriptName" color="secondary" inverted readonly :after="[{icon: 'more_horiz', handler(){}}]" :disable="!ta.IsActive"></q-input>
					</q-item-tile>
				</q-item-main>
			</q-item>

			<q-item tag="label" v-if="isActionHidScript">
<select-string-from-array :values="$store.state.StoredHIDScriptsList" v-model="ShowSelectHIDScriptModal" title="Select HIDScript" @load="ta.ActionData.ScriptName=$event"></select-string-from-array>

				<q-item-main>
					<q-item-tile label>Script name</q-item-tile>
					<q-item-tile sublabel>Name of a stored HIDScript</q-item-tile>
					<q-item-tile>
<q-input @click="updateStoredHIDScriptsList();ShowSelectHIDScriptModal=true" v-model="ta.ActionData.ScriptName" color="secondary" inverted readonly :after="[{icon: 'more_horiz', handler(){}}]" :disable="!ta.IsActive"></q-input>
					</q-item-tile>
				</q-item-main>
				
				
			</q-item>

			<q-item tag="label" v-if="isActionGPIOOut">
				<q-item-main>
					<q-item-tile label>GPIO Number</q-item-tile>
					<q-item-tile sublabel>The number of the GPIO to output on</q-item-tile>
					<q-item-tile>
						<q-select v-model="ta.ActionData.GpioNum" :options="gpionum" color="secondary" inverted :disable="!ta.IsActive"></q-select>
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
						<q-input v-model="ta.ActionData.Value" color="secondary"  type="number" decimals="0" inverted :disable="!ta.IsActive"></q-input>
					</q-item-tile>
				</q-item-main>
			</q-item>




			<q-item tag="label" v-if="isActionDeploySettingsTemplate">
				<q-item-main>
					<q-item-tile label>Type</q-item-tile>
					<q-item-tile sublabel>Name of the stored settings template to load</q-item-tile>
					<q-item-tile>
						<q-select v-model="ta.ActionData.Type" :options="templatetypes" color="secondary" @input="ta.ActionData.TemplateName=''" inverted :disable="!ta.IsActive"></q-select>
					</q-item-tile>
				</q-item-main>
			</q-item>

			<q-item tag="label" v-if="isActionDeploySettingsTemplate">
<select-string-from-array :values="typedTemplateList" v-model="ShowSelectTemplateModal" title="Select template" @load="ta.ActionData.TemplateName=$event"></select-string-from-array>
				<q-item-main>
					<q-item-tile label>Template name</q-item-tile>
					<q-item-tile sublabel>Name of the stored settings template to load</q-item-tile>
					<q-item-tile>
<q-input @click="actionTemplateTypeUpdate(); ShowSelectTemplateModal=true" v-model="ta.ActionData.TemplateName" color="secondary" inverted readonly :after="[{icon: 'more_horiz', handler(){}}]" :disable="!ta.IsActive"></q-input>
					</q-item-tile>
				</q-item-main>
			</q-item>


		</q-list>
`



const templateTriggerActionManager = `
<q-page padding>
	<modal-string-input v-model="showStoreTASModal" title="Store selected TriggerActions" @save="storeTAS($event)"></modal-string-input>
	<select-string-from-array :values="$store.state.StoredTriggerActionSetsList" v-model="showReplaceTASModal" title="Replace current Trigger Actions with stored set" @load="replaceCurrentTAS($event)"></select-string-from-array>
	<select-string-from-array :values="$store.state.StoredTriggerActionSetsList" v-model="showAddTASModal" title="Add stored set to current Trigger Actions" @load="addToCurrentTAS($event)"></select-string-from-array>

	<div class="row gutter-sm">
		<div class="col-12">
			<q-card>
				<q-card-title>
					TriggerAction Manager
				</q-card-title>

<!--
				<q-card-actions>
    				<q-btn label="add TriggerAction" @click="addTA" icon="note_add" />
    				<q-btn label="store template" @click="showStoreTASModal=true" icon="save" />
    				<q-btn label="load template" @click="updateStoredTriggerActionSetsList(); showReplaceTASModal=true" icon="settings_backup_restore" />
    				<q-btn label="insert template" @click="updateStoredTriggerActionSetsList(); showAddTASModal=true" icon="add_to_photos" />
  				</q-card-actions>
-->
				<q-card-main>
					<div class="row gutter-sm">
	    				<div class="col-12 col-sm"><q-btn class="fit" color="primary" label="add one" @click="addTA" icon="note_add" /></div>
    					<div class="col-12 col-sm"><q-btn class="fit" color="secondary" label="store" @click="showStoreTASModal=true" icon="save" /></div>
    					<div class="col-12 col-sm"><q-btn class="fit" color="warning" label="load & replace" @click="updateStoredTriggerActionSetsList(); showReplaceTASModal=true" icon="settings_backup_restore" /></div>
    					<div class="col-12 col-sm"><q-btn class="fit" color="warning" label="load & add" @click="updateStoredTriggerActionSetsList(); showAddTASModal=true" icon="add_to_photos" /></div>
					</div>
  				</q-card-main>


			</q-card>
		</div>

		<div class="col-12 col-lg-6" v-for="ta in $store.getters.triggerActions"> 
			<TriggerAction :key="ta.Id" :ta="ta" overview></TriggerAction>
		</div>
	</div>
</q-page>	

`
