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


func InitComponentsTriggerActions() {
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

//
const templateTriggerAction = `
<q-card class="fit">
<!-- {{ ta }} -->
	<q-card-title>TriggerAction (ID {{ ta.Id }})</q-card-title>
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
						<q-input v-model="ta.ActionData.Value" color="secondary" type="number" decimals="0" inverted :disable="!ta.IsActive"></q-input>
					</q-item-tile>
				</q-item-main>
			</q-item>




			<q-item tag="label" v-if="isActionDeploySettingsTemplate">
				<q-item-main>
					<q-item-tile label>Type</q-item-tile>
					<q-item-tile sublabel>Name of the stored settings template to load</q-item-tile>
					<q-item-tile>
						<q-select v-model="ta.ActionData.Type" :options="templatetypes" color="secondary" inverted :disable="!ta.IsActive"></q-select>
					</q-item-tile>
				</q-item-main>
			</q-item>

			<q-item tag="label" v-if="isActionDeploySettingsTemplate">
				<q-item-main>
					<q-item-tile label>Template name</q-item-tile>
					<q-item-tile sublabel>Name of the stored settings template to load</q-item-tile>
					<q-item-tile>
						<q-input v-model="ta.ActionData.TemplateName" color="secondary" inverted :disable="!ta.IsActive"></q-input>
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
