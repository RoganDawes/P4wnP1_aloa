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

func generateSelectOptionsGroupReceiveMultiType() *js.Object {
	tts := js.Global.Get("Array").New()
	type option struct {
		*js.Object
		Label string `js:"label"`
		Value GroupReceiveMultiType `js:"value"`
	}

	for _, value := range availableGroupReceiveMulti {
		label := groupReceiveMultiNames[value]
		o := option{Object:O()}
		o.Value = value
		o.Label = label
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

func generateSelectOptionsGPIONames(vm *hvue.VM) *js.Object {
	tts := js.Global.Get("Array").New()
	type option struct {
		*js.Object
		Label string `js:"label"`
		Value string `js:"value"`
	}

	gpioNames := vm.Store.Get("state").Get("GpioNamesList")
	for i := 0; i < gpioNames.Length(); i++ {
		gpioName := gpioNames.Index(i).String()
		o := option{Object:O()}
		o.Value = gpioName
		o.Label = gpioName
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

/*
type TriggerActionCompData struct {
	*js.Object
	Edit bool `js:"Edit"`
}
*/


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
		hvue.Method("editTa",
			func(vm *hvue.VM, taID *js.Object) {
				vm.Get("$refs").Index(taID.Int()).Index(0).Call("setEditMode", true)
			}),
		hvue.Method("addTA",
			func(vm *hvue.VM) {
				promise := vm.Get("$store").Call("dispatch", VUEX_ACTION_ADD_NEW_TRIGGER_ACTION)
				promise.Call("then",
					func(value *js.Object) {
						// set the trigger action into edit mode
						vm.Call("editTa", value)
					},
					func(reason *js.Object) {
						println("add TriggerAction failed", reason)
					},
				)

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
		hvue.Method("deleteStored",
			func(vm *hvue.VM, storedTASName *js.Object) {
				//vm.Get("$q").Call("notify", "Add '" + storedTASName.String() +"' to current TAS")
				vm.Get("$store").Call("dispatch", VUEX_ACTION_DELETE_STORED_TRIGGER_ACTION_SET, storedTASName)
			}),
		hvue.Method("updateStoredTriggerActionSetsList",
			func(vm *hvue.VM) {
				vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_STORED_TRIGGER_ACTION_SETS_LIST)
			}),
		hvue.Mounted(func(vm *hvue.VM) {
			vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_CURRENT_TRIGGER_ACTIONS_FROM_SERVER)
			vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_GPIO_NAMES_LIST)

			js.Global.Set("tam",vm)

		}),
	)



	hvue.NewComponent(
		"TriggerAction",
		hvue.Template(templateTriggerAction),
		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := &struct {
				*js.Object
				Edit bool `js:"Edit"`
			}{Object:O()}
			data.Edit = false
			return data
		}),
		hvue.PropObj("ta"),
		hvue.PropObj("edit",
			hvue.Types(hvue.PBoolean),
		),
		hvue.Method("setEditMode",
			func(vm *hvue.VM, enabled bool) {
				vm.Data.Set("Edit", enabled)
			}),
		hvue.Mounted(func(vm *hvue.VM) {
			vm.Set("Edit", vm.Get("edit"))
		}),
	)

	hvue.NewComponent(
		"TriggerActionOverview",
		hvue.Template(templateTriggerActionOverview),
		hvue.PropObj("ta"),
		hvue.PropObj("edit",
			hvue.Types(hvue.PBoolean),
			),
/*
		hvue.Mounted(func(vm *hvue.VM) {
			data := TriggerActionCompData{Object: vm.Data}
			data.Edit = vm.Get("edit").Bool()
		}),
*/
		/*
		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := &TriggerActionCompData{Object: O()}
			data.Edit = false

			return data
		}),
		*/
		hvue.ComputedWithGetSet("EditMode",
			func(vm *hvue.VM) interface{} {
				/*
				data := TriggerActionCompData{Object: vm.Data}
				return data.Edit
				*/
				return vm.Get("edit")
			},
			func(vm *hvue.VM, newValue *js.Object) {
/*
				data := TriggerActionCompData{Object: vm.Data}
				data.Edit = newValue.Bool()
*/
				// Emit event for editmode change
				vm.Emit("edit", newValue)
			}),
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
			case ta.IsTriggerGroupReceiveMulti():
				t := jsTriggerGroupReceiveMulti{Object: ta.TriggerData}
				strTrigger += " ("
				strTrigger += t.GroupName + ": "
				switch t.Type {
				case GroupReceiveMultiType_SEQUENCE:
					strTrigger += "sequence of"
				case GroupReceiveMultiType_EXACT_SEQUENCE:
					strTrigger += "exact sequence of"
				case GroupReceiveMultiType_OR:
					strTrigger += "one of"
				case GroupReceiveMultiType_AND:
					strTrigger += "all from"
				}
				strTrigger += " ["
				for idx,val  := range t.Values {
					if idx != 0 {
						strTrigger += ", "
					}
					strTrigger += strconv.Itoa(int(val))
				}
				strTrigger += "]"
				strTrigger += ")"
			case ta.IsTriggerGPIOIn():
				t := jsTriggerGPIOIn{Object: ta.TriggerData}
				strTrigger += " ("
				strTrigger += t.GpioName
				strTrigger += ": " + gpioInEdgeNames[t.Edge]
				strTrigger += ", resistor: " + gpioInPullUpDownNames[t.PullUpDown]
				strTrigger += ", debounce: " + strconv.Itoa(int(t.DebounceMillis)) + "ms"
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
				strAction += a.GpioName
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

		hvue.Method(
			"updateTA",
			func(vm *hvue.VM) {
				ta_obj := vm.Get("ta")
				println("update ta: ", ta_obj)
/*
				//Replace the whole TriggerActionSet of server with the current one from vuex store
				// ToDo: This has to be changed to update a single action (inconssistnecy with multiple clients, all TA IDs change, overhead of transferring a whole set) -> has to be implemented like deleteTA logic
				currentTas := vm.Get("$store").Get("state").Get("triggerActionList") //Current TriggerActionSet of ViewModel
				vm.Get("$store").Call("dispatch", VUEX_ACTION_DEPLOY_TRIGGER_ACTION_SET_REPLACE, currentTas)
*/
				updateTas := NewTriggerActionSet()
				updateTas.UpdateEntry(&jsTriggerAction{Object: ta_obj})
				vm.Get("$store").Call("dispatch", VUEX_ACTION_UPDATE_TRIGGER_ACTIONS, updateTas)
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
		hvue.Computed("groupReceiveMultiSelect", func(vm *hvue.VM) interface{} {
			return generateSelectOptionsGroupReceiveMultiType()
		}),
		hvue.Computed("edge", func(vm *hvue.VM) interface{} {
			return generateSelectOptionsGPIOInEdges()
		}),
		hvue.Computed("gpioname", func(vm *hvue.VM) interface{} {
			return generateSelectOptionsGPIONames(vm)
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
			"TriggerGroupReceiveMultiAddValue",
			func(vm *hvue.VM, newVal *js.Object) {
				println("Force add", newVal)
				ta := &jsTriggerAction{Object: vm.Get("ta")}
				if !ta.IsTriggerGroupReceiveMulti() { return }

				// cast data Object to jsTriggerGroupReceiveMulti
				tgrs := &jsTriggerGroupReceiveMulti{Object: ta.TriggerData}
				strVal := newVal.String()
				if intVal,errconv := strconv.Atoi(strVal); errconv == nil {
					//append to Values
					tgrs.Values = append(tgrs.Values, int32(intVal))
				}
			}),
		hvue.ComputedWithGetSet(
			"TriggerGroupReceiveMultiValues",
			func(vm *hvue.VM) interface{} {
				ta := &jsTriggerAction{Object: vm.Get("ta")}
				if !ta.IsTriggerGroupReceiveMulti() { return []string{} }

				// cast data Object to jsTriggerGroupReceiveMulti
				tgrs := &jsTriggerGroupReceiveMulti{Object: ta.TriggerData}

				res := make([]string, len(tgrs.Values))
				for idx,intVal := range tgrs.Values {
					res[idx] = strconv.Itoa(int(intVal))
				}
				return res
			},
			func(vm *hvue.VM, newValue *js.Object) {
				ta := &jsTriggerAction{Object: vm.Get("ta")}
				if !ta.IsTriggerGroupReceiveMulti() { return }

				// cast data Object to jsTriggerGroupReceiveMulti
				tgrs := &jsTriggerGroupReceiveMulti{Object: ta.TriggerData}

				// clear old array
				tgrs.Values = []int32{}

				// iterate over newValue, which is assumed to be an Array of strings
				for idx := 0; idx < newValue.Length(); idx++ {
					//fetch value
					strVal := newValue.Index(idx).String()
					// try to cast to int
					if intVal,errconv := strconv.Atoi(strVal); errconv == nil {
						//append to Values
						tgrs.Values = append(tgrs.Values, int32(intVal))
					}
				}
			}),
		hvue.Computed("isTriggerGPIOIn", func(vm *hvue.VM) interface{} {
			return (&jsTriggerAction{Object: vm.Get("ta")}).IsTriggerGPIOIn()
		}),
		hvue.Computed("isTriggerGroupReceive", func(vm *hvue.VM) interface{} {
			return (&jsTriggerAction{Object: vm.Get("ta")}).IsTriggerGroupReceive()
		}),
		hvue.Computed("isTriggerGroupReceiveMulti", func(vm *hvue.VM) interface{} {
			return (&jsTriggerAction{Object: vm.Get("ta")}).IsTriggerGroupReceiveMulti()
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
					return vm.Store.Get("state").Get("StoredMasterTemplateList")
				case TemplateTypeBluetooth:
					return vm.Store.Get("state").Get("StoredBluetoothSettingsList")
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
					vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_STORED_MASTER_TEMPLATE_LIST)
				case TemplateTypeBluetooth:
					vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_STORED_BLUETOOTH_SETTINGS_LIST)
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
		hvue.Computed("gpioname", func(vm *hvue.VM) interface{} {
			return generateSelectOptionsGPIONames(vm)
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
	<TriggerActionOverview :ta="ta" :edit="Edit" @edit="Edit=$event"></TriggerActionOverview>
</div>
`
const templateTriggerActionOverview = `
<div>
<q-modal v-model="EditMode" no-route-dismiss no-esc-dismiss no-backdrop-dismiss>
	<TriggerActionEdit :ta="ta">
		<span slot="actions">
			<q-btn color="primary" @click="updateTA(); EditMode=false" label="更新" />
			<q-btn color="secondary" @click="cancelUpdateTA(); EditMode=false" label="取消" />
		</span>
	</TriggerActionEdit>
	
</q-modal>

<q-card tag="label" :color="computedColor" :text-color="ta.IsActive ? '': 'light'" :disabled="ta.Immutable" :dark="ta.Immutable">
	<q-card-title>
		{{ ta.Immutable ? "immutable, " : "" }}
		{{ ta.IsActive ? "enabled" : "disabled" }}
		触发器动作 (ID {{ ta.Id }})
	
		<span slot="subtitle">
			<q-icon name="input"></q-icon> 
			{{ strTrigger }}
			<br><q-icon name="launch"></q-icon>
			{{ strAction }}{{ta.OneShot ? " 仅运行一次" : "" }}	
		</span>

		<div slot="right" v-if="!ta.Immutable">
			<q-btn color="primary" icon="edit" @click="EditMode=true" flat></q-btn>
			<q-btn color="negative" icon="delete" @click="deleteTA" flat></q-btn>
		</div>
	</q-card-title>
</q-card>
</div>
`

const templateTriggerActionEdit = `
<q-card class="fit">
	<q-card-title>
		触发器动作
		<span slot="subtitle">ID {{ ta.Id }}</span>
		<!-- <q-btn slot="right" icon="more_vert" flat></q-btn> -->
	</q-card-title>
	<q-list>
			<q-item tag="label" link>
				<q-item-side>
					<q-toggle v-model="ta.IsActive"></q-toggle>
				</q-item-side>
				<q-item-main>
					<q-item-tile label>已启用</q-item-tile>
					<q-item-tile sublabel>如果未启用, 则忽略触发器动作</q-item-tile>
				</q-item-main>
			</q-item>

			<q-item tag="label" link :disabled="!ta.IsActive">
				<q-item-side>
					<q-toggle v-model="ta.OneShot" :disable="!ta.IsActive"></q-toggle>
				</q-item-side>
				<q-item-main>
					<q-item-tile label>仅执行一次</q-item-tile>
					<q-item-tile sublabel>每次发生相应事件时触发器都会触发， 如果启用"一次性"，则仅触发一次</q-item-tile>
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
					<q-item-tile label>触发器</q-item-tile>
					<q-item-tile sublabel>选择必须发生的事件以启动所选操作</q-item-tile>
					<q-item-tile>
						<q-select v-model="triggerType" :options="triggertypes" inverted :disable="!ta.IsActive"></q-select>
					</q-item-tile>
				</q-item-main>
			</q-item>

			<q-item tag="label" v-if="isTriggerGroupReceive || isTriggerGroupReceiveMulti">
				<q-item-main>
					<q-item-tile label>触发器组名称</q-item-tile>
					<q-item-tile sublabel>仅考虑为此组名发送的值</q-item-tile>
					<q-item-tile>
						<q-input v-model="ta.TriggerData.GroupName" inverted :disable="!ta.IsActive"></q-input>
					</q-item-tile>
				</q-item-main>
			</q-item>
			<q-item tag="label" v-if="isTriggerGroupReceive">
				<q-item-main>
					<q-item-tile label>触发值</q-item-tile>
					<q-item-tile sublabel>必须接收以激活触发器的数值</q-item-tile>
					<q-item-tile>
						<q-input v-model="ta.TriggerData.Value" type="number" inverted :disable="!ta.IsActive"></q-input>
					</q-item-tile>
				</q-item-main>
			</q-item>
			<q-item tag="label" v-if="isTriggerGroupReceiveMulti">
				<q-item-main>
					<q-item-tile label>触发值</q-item-tile>
					<q-item-tile sublabel>必须接收的数值才能激活触发器</q-item-tile>
					<q-item-tile>
						<q-chips-input v-model="TriggerGroupReceiveMultiValues" @duplicate="TriggerGroupReceiveMultiAddValue($event)" type="number" decimals="0" inverted :disable="!ta.IsActive"></q-chips-input>
					</q-item-tile>
				</q-item-main>
			</q-item>
<!--
			<q-item tag="label" link :disabled="!ta.IsActive" v-if="isTriggerGroupReceiveMulti">
				<q-item-side>
					<q-toggle v-model="ta.TriggerData.IgnoreOutOfOrder" :disable="!ta.IsActive"></q-toggle>
				</q-item-side>
				<q-item-main>
					<q-item-tile label>忽略无序值</q-item-tile>
					<q-item-tile sublabel>如果启用，序列可能会被其他值中断，如果禁用，他们必须准确到达</q-item-tile>
				</q-item-main>
			</q-item>
-->
			<q-item tag="label" v-if="isTriggerGroupReceiveMulti">
				<q-item-main>
					<q-item-tile label>类型</q-item-tile>
					<q-item-tile sublabel>选择应如何检查值(逻辑OR，逻辑AND，序列或精确序列)</q-item-tile>
					<q-item-tile>
						<q-select v-model="ta.TriggerData.Type" :options="groupReceiveMultiSelect" inverted :disable="!ta.IsActive"></q-select>
					</q-item-tile>
				</q-item-main>
			</q-item>



			<q-item tag="label" v-if="isTriggerGPIOIn">
				<q-item-main>
					<q-item-tile label>GPIO编号</q-item-tile>
					<q-item-tile sublabel>要监听的GPIO的编号</q-item-tile>
					<q-item-tile>
						<q-select v-model="ta.TriggerData.GpioName" :options="gpioname" inverted :disable="!ta.IsActive"></q-select>
					</q-item-tile>
				</q-item-main>
			</q-item>
			<q-item tag="label" v-if="isTriggerGPIOIn">
				<q-item-main>
					<q-item-tile label>电阻</q-item-tile>
					<q-item-tile sublabel>选择是否应使用内部上拉/下拉电阻</q-item-tile>
					<q-item-tile>
						<q-select v-model="ta.TriggerData.PullUpDown" :options="pullupdown" inverted :disable="!ta.IsActive"></q-select>
					</q-item-tile>
				</q-item-main>
			</q-item>
			<q-item tag="label" v-if="isTriggerGPIOIn">
				<q-item-main>
					<q-item-tile label>阈值</q-item-tile>
					<q-item-tile sublabel>触发触发器必须发生什么电平变化</q-item-tile>
					<q-item-tile>
						<q-select v-model="ta.TriggerData.Edge" :options="edge" inverted :disable="!ta.IsActive"></q-select>
					</q-item-tile>
				</q-item-main>
			</q-item>
			<q-item tag="label" v-if="isTriggerGPIOIn">
				<q-item-main>
					<q-item-tile label>去抖持续时间</q-item-tile>
					<q-item-tile sublabel>此持续时间中的连续边缘事件将被忽略</q-item-tile>
					<q-item-tile>
						<q-input v-model="ta.TriggerData.DebounceMillis" type="number" suffix="ms" inverted :disable="!ta.IsActive"></q-input>
					</q-item-tile>
				</q-item-main>
			</q-item>

		</q-list>
`

const templateAction = `
		<q-list class="fit" no-border link :disabled="!ta.IsActive">
			<q-item tag="label">
				<q-item-main>
					<q-item-tile label>动作</q-item-tile>
					<q-item-tile sublabel>选择触发器触发时应该启动的动作</q-item-tile>
					<q-item-tile>
						<q-select v-model="actionType" :options="actiontypes" color="secondary" inverted :disable="!ta.IsActive"></q-select>
					</q-item-tile>
				</q-item-main>
			</q-item>

			<q-item tag="label" v-if="isActionBashScript">
<select-string-from-array :values="$store.state.StoredBashScriptsList" v-model="ShowSelectBashScriptModal" title="Select BASH script" @load="ta.ActionData.ScriptName=$event"></select-string-from-array>
				<q-item-main>
					<q-item-tile label>脚本路径</q-item-tile>
					<q-item-tile sublabel>应该发布的Bash脚本的路径</q-item-tile>
					<q-item-tile>
<q-input @click="updateStoredBashScriptsList();ShowSelectBashScriptModal=true" v-model="ta.ActionData.ScriptName" color="secondary" inverted readonly :after="[{icon: 'more_horiz', handler(){}}]" :disable="!ta.IsActive"></q-input>
					</q-item-tile>
				</q-item-main>
			</q-item>

			<q-item tag="label" v-if="isActionHidScript">
<select-string-from-array :values="$store.state.StoredHIDScriptsList" v-model="ShowSelectHIDScriptModal" title="Select HIDScript" @load="ta.ActionData.ScriptName=$event"></select-string-from-array>

				<q-item-main>
					<q-item-tile label>脚本名称</q-item-tile>
					<q-item-tile sublabel>存储的HID脚本的名称</q-item-tile>
					<q-item-tile>
<q-input @click="updateStoredHIDScriptsList();ShowSelectHIDScriptModal=true" v-model="ta.ActionData.ScriptName" color="secondary" inverted readonly :after="[{icon: 'more_horiz', handler(){}}]" :disable="!ta.IsActive"></q-input>
					</q-item-tile>
				</q-item-main>
				
				
			</q-item>

			<q-item tag="label" v-if="isActionGPIOOut">
				<q-item-main>
					<q-item-tile label>GPIO编号</q-item-tile>
					<q-item-tile sublabel>要输出的GPIO的编号</q-item-tile>
					<q-item-tile>
						<q-select v-model="ta.ActionData.GpioName" :options="gpioname" color="secondary" inverted :disable="!ta.IsActive"></q-select>
					</q-item-tile>
				</q-item-main>
			</q-item>
			<q-item tag="label" v-if="isActionGPIOOut">
				<q-item-main>
					<q-item-tile label>输出</q-item-tile>
					<q-item-tile sublabel>在给定的GPIO上输出低/高或切换输出</q-item-tile>
					<q-item-tile>
						<q-select v-model="ta.ActionData.Value" :options="gpiooutvalues" color="secondary" inverted :disable="!ta.IsActive"></q-select>
					</q-item-tile>
				</q-item-main>
			</q-item>


			<q-item tag="label" v-if="isActionGroupSend">
				<q-item-main>
					<q-item-tile label>组名</q-item-tile>
					<q-item-tile sublabel>要发送到的组的名称(必须匹配相应的侦听器)</q-item-tile>
					<q-item-tile>
						<q-input v-model="ta.ActionData.GroupName" color="secondary" inverted :disable="!ta.IsActive"></q-input>
					</q-item-tile>
				</q-item-main>
			</q-item>
			<q-item tag="label" v-if="isActionGroupSend">
				<q-item-main>
					<q-item-tile label>通道值</q-item-tile>
					<q-item-tile sublabel>发送到组通道的数值</q-item-tile>
					<q-item-tile>
						<q-input v-model="ta.ActionData.Value" color="secondary"  type="number" inverted :disable="!ta.IsActive"></q-input>
					</q-item-tile>
				</q-item-main>
			</q-item>




			<q-item tag="label" v-if="isActionDeploySettingsTemplate">
				<q-item-main>
					<q-item-tile label>模板类型</q-item-tile>
					<q-item-tile sublabel>选择要加载的模板类型</q-item-tile>
					<q-item-tile>
						<q-select v-model="ta.ActionData.Type" :options="templatetypes" color="secondary" @input="ta.ActionData.TemplateName=''" inverted :disable="!ta.IsActive"></q-select>
					</q-item-tile>
				</q-item-main>
			</q-item>

			<q-item tag="label" v-if="isActionDeploySettingsTemplate">
<select-string-from-array :values="typedTemplateList" v-model="ShowSelectTemplateModal" title="Select template" @load="ta.ActionData.TemplateName=$event"></select-string-from-array>
				<q-item-main>
					<q-item-tile label>模板名</q-item-tile>
					<q-item-tile sublabel>要加载的存储设置模板的名称</q-item-tile>
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
	<select-string-from-array :values="$store.state.StoredTriggerActionSetsList" v-model="showReplaceTASModal" title="Replace current Trigger Actions with stored set" @load="replaceCurrentTAS($event)" @delete="deleteStored($event)" with-delete></select-string-from-array>
	<select-string-from-array :values="$store.state.StoredTriggerActionSetsList" v-model="showAddTASModal" title="Add stored set to current Trigger Actions" @load="addToCurrentTAS($event)" @delete="deleteStored($event)" with-delete></select-string-from-array>

	<div class="row gutter-sm">
		<div class="col-12">
			<q-card>
				<q-card-title>
					触发器动作管理器
				</q-card-title>

				<q-card-main>
					<div class="row gutter-sm">
	    				<div class="col-6 col-sm"><q-btn class="fit" color="primary" label="添加一个" @click="addTA" icon="add_box" /></div>
    					<div class="col-6 col-sm"><q-btn class="fit" color="secondary" label="保存" @click="showStoreTASModal=true" icon="cloud_upload" /></div>
    					<div class="col-6 col-sm"><q-btn class="fit" color="warning" label="加载并替换" @click="updateStoredTriggerActionSetsList(); showReplaceTASModal=true" icon="cloud_download" /></div>
    					<div class="col-6 col-sm"><q-btn class="fit" color="warning" label="加载并添加" @click="updateStoredTriggerActionSetsList(); showAddTASModal=true" icon="add_to_photos" /></div>
					</div>
  				</q-card-main>


			</q-card>
		</div>

		<div class="col-12 col-lg-6" v-for="ta in $store.getters.triggerActions"> 
			<TriggerAction :ref="ta.Id" :key="ta.Id" :ta="ta" :edit="false"></TriggerAction>
		</div>
	</div>
</q-page>	

`
