// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/mame82/hvue"
	"strings"
)

type jsMasterTemplate struct {
	*js.Object
	TemplateNameBluetooth      string   `js:"TemplateNameBluetooth"`
	TemplateNameUSB            string   `js:"TemplateNameUSB"`
	TemplateNameWiFi           string   `js:"TemplateNameWiFi"`
	TemplateNameTriggerActions string   `js:"TemplateNameTriggerActions"`
	TemplateNamesNetwork       []string `js:"TemplateNamesNetwork"`
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

func InitComponentsGeneric() {

	hvue.NewComponent(
		"generic",
		hvue.Template(compGeneric),
	)
	hvue.NewComponent(
		"startup-settings",
		hvue.Template(compStartupSettings),
	)
	hvue.NewComponent(
		"system",
		hvue.Template(compSystem),
	)
	hvue.NewComponent(
		"master-template",
		hvue.Template(compMasterTemplate),
		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := &struct {
				*js.Object
				MasterTemplate *jsMasterTemplate `js:"MasterTemplate"`

				ShowTemplateSelectBluetooth bool `js:"ShowTemplateSelectBluetooth"`
				ShowTemplateSelectWiFi bool `js:"ShowTemplateSelectWiFi"`
				ShowTemplateSelectUSB bool `js:"ShowTemplateSelectUSB"`
				ShowTemplateSelectTriggerAction bool `js:"ShowTemplateSelectTriggerAction"`
				ShowTemplateSelectNetwork bool `js:"ShowTemplateSelectNetwork"`

			}{Object: O()}

			data.MasterTemplate = NewMasterTemplate()
			data.ShowTemplateSelectBluetooth = false
			data.ShowTemplateSelectWiFi = false
			data.ShowTemplateSelectUSB = false
			data.ShowTemplateSelectTriggerAction = false
			data.ShowTemplateSelectNetwork = false

			return data
		}),
	)
	hvue.NewComponent(
		"database",
		hvue.Template(compDatabase),
	)



	hvue.NewComponent(
		"select-network-templates",
		hvue.Template(templateSelectNetworkTemplatesModal),
		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := struct {
				*js.Object
				CurrentSelection []string `js:"CurrentSelection"`
			}{Object:O()}
			data.CurrentSelection = []string{}

			return &data
		}),
		/*
		hvue.Mounted(func(vm *hvue.VM) {
			vm.Set("CurrentSelection", vm.Get("values").Index(0))
			//			println("Index 0 on mount", vm.Get("values").Index(0))
		}),
		*/
		hvue.Computed("options", func(vm *hvue.VM) interface{} {
			inVals := vm.Get("values")
			options := js.Global.Get("Array").New()

			for i:=0; i < inVals.Length(); i++ {
				val := inVals.Index(i).String()
				entry := struct{
					*js.Object
					Label string `js:"label"`
					Value string `js:"value"`
				}{Object:O()}
				entry.Label = val
				entry.Value = val

				options.Call("push", entry)
			}

			return options
		}),

		hvue.ComputedWithGetSet(
			"visible",
			func(vm *hvue.VM) interface{} {
				return vm.Get("show")
			},
			func(vm *hvue.VM, newValue *js.Object) {
				vm.Call("$emit", "show", newValue)
			},
		),
		hvue.ComputedWithGetSet("selection",
			func(vm *hvue.VM) interface{} {
				return vm.Get("value")
			},
			func(vm *hvue.VM, newValue *js.Object) {
				vm.Call("$emit", "input", newValue)
			}),
		hvue.Computed("available", func(vm *hvue.VM) interface{} {


			sel := vm.Get("selection")
			selection := make(map[string]bool, 0)
			selectionPrefix := make(map[string]bool, 0)
			for i:=0; i<sel.Length(); i++ {
				name := sel.Index(i).String()
				selection[name] = true
				selectionPrefix[strings.Split(name,"_")[0]] = true
			}

			vals := vm.Get("values")
			values := make([]string, vals.Length())
			for i:=0; i<vals.Length(); i++ {
				values[i] = vals.Index(i).String()
			}

			var res []string

			for _,v := range values {
				prefix := strings.Split(v,"_")[0]
				_,inPrefix := selectionPrefix[prefix]
				_,inSelected := selection[v]

				if !inPrefix || inSelected {
					res = append(res, v)
				}
			}

			return res
		}),
		hvue.Method(
			"onLoadPressed",
			func(vm *hvue.VM) {
				vm.Call("$emit", "input", vm.Get("selection"))
			},
		),
		hvue.Method(
			"onDeletePressed",
			func(vm *hvue.VM, name *js.Object) {
				vm.Call("$emit", "delete", name)
			},
		),
		hvue.PropObj(
			"values",
			hvue.Required,
		),
		hvue.PropObj(
			"show",
			hvue.Required,
			hvue.Types(hvue.PBoolean),
		),
		hvue.PropObj(
			"with-delete",
			hvue.Types(hvue.PBoolean),
			hvue.Default(false),
		),
		hvue.PropObj(
			"title",
			hvue.Types(hvue.PString),
		),
		hvue.PropObj(
			"value",
		),
	)

	//return o.NewComponent()
}

/*
- Startup settings (templates)
- Reboot & shutdown
- States (CPU load, memory usage etc.)
- DB backup & Restore
 */

const compGeneric = `
<q-page padding>
	<div class="row gutter-sm">
		<div class="col-12 col-lg">
			<master-template />
		</div>

		<div class="col-12 col-lg">
			<startup-settings />
		</div>

		<div class="col-12 col-lg">
			<system />
		</div>

		<div class="col-12 col-lg">
			<database />
		</div>

	</div>
</q-page>
`

const compStartupSettings = `
<q-card>
	<q-card-title>
		Startup Settings
	</q-card-title>

	<q-card-main>
		<div class="row gutter-sm">

		</div>
	</q-card-main>
</q-card>
`

const compMasterTemplate = `
<q-card>
	<q-card-title>
		Master Template
	</q-card-title>

	{{ $data }}

	<q-card-main>
		

		<div class="row gutter-sm">

			<q-item tag="label">
<select-string-from-array :values="$store.state.StoredUSBSettingsList"  v-model="ShowTemplateSelectUSB" title="Select USB template" @load="MasterTemplate.TemplateNameUSB=$event"></select-string-from-array>
				<q-item-main>
					<q-item-tile label>USB Template</q-item-tile>
					<q-item-tile sublabel>If not empty, the selected USB settings are deployed along with the master template</q-item-tile>
					<q-item-tile>
<q-input @click="ShowTemplateSelectUSB=true" v-model="MasterTemplate.TemplateNameUSB" color="primary" inverted readonly :after="[{icon: 'more_horiz', handler(){}}]"></q-input>
					</q-item-tile>
				</q-item-main>
			</q-item>

			<q-item tag="label">
<select-network-templates 
	:values="$store.state.StoredEthernetInterfaceSettingsList" 
	:show="ShowTemplateSelectNetwork" 
	@show="ShowTemplateSelectNetwork=$event"
	title="Select Network templates" 
	v-model="MasterTemplate.TemplateNamesNetwork" 
/>
				<q-item-main>
					<q-item-tile label>Networks templates</q-item-tile>
					<q-item-tile sublabel>If not empty, the selected network templates are deployed along with the master template. Only one template could be selected per interface.</q-item-tile>
					<q-item-tile>
<q-chips-input v-model="MasterTemplate.TemplateNamesNetwork"  color="primary" inverted :after="[{icon: 'add', handler(){ShowTemplateSelectNetwork=true}}]" />
					</q-item-tile>
				</q-item-main>
			</q-item>



		</div>
	</q-card-main>
</q-card>
`

const compSystem = `
<q-card>
	<q-card-title>
		System
	</q-card-title>

	<q-card-main>
		<div class="row gutter-sm">
			<div class="col"> <q-btn class="fit" color="warning" label="reboot" icon="refresh" /> </div>
			<div class="col"> <q-btn class="fit" color="negative" label="shutdown" icon="power_settings_new" /> </div>
		</div>
	</q-card-main>
</q-card>
`

const compDatabase = `
<q-card>
	<q-card-title>
		Database
	</q-card-title>

	<q-card-main>
		<div class="row gutter-sm">
			<div class="col"> <q-btn class="fit" color="primary" label="backup" icon="cloud_upload" /> </div>
			<div class="col"> <q-btn class="fit" color="negative" label="restore" icon="cloud_download" /> </div>
		</div>
	</q-card-main>
</q-card>
`

const templateSelectNetworkTemplatesModal = `
	<q-modal v-model="visible">
		<q-modal-layout>
			<q-toolbar slot="header">
				<q-toolbar-title>
					{{ title }}
    				<span slot="subtitle">
      					Only one template could be selected per interface
    				</span>
				</q-toolbar-title>
			</q-toolbar>

			<q-list>

				<q-item link tag="label" v-for="name in available" :key="name">
					<q-item-side>
						<q-checkbox v-model="selection" :val="name"/>
					</q-item-side>
					<q-item-main>
						<q-item-tile label>{{ name }}</q-item-tile>
					</q-item-main>
					<q-item-side v-if="withDelete" right>
						<q-btn icon="delete" color="negative" @click="onDeletePressed(name)" round flat />
					</q-item-side>
				</q-item>


			</q-list>

			<q-list slot="footer">
				<q-item tag="label">
					<q-item-main>
						<q-item-tile>
<!--
							<q-btn color="primary" v-show="CurrentSelection != undefined" label="ok" @click="onLoadPressed(); visible=false"/>							
-->
							<q-btn color="secondary" v-close-overlay label="close" />
						</q-item-tile>
					</q-item-main>
				</q-item>
			</q-list>

		</q-modal-layout>
	</q-modal>
`
