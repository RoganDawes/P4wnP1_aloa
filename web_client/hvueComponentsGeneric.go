// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/mame82/hvue"
	"strings"
)

func InitComponentsGeneric() {

	hvue.NewComponent(
		"generic",
		hvue.Template(compGeneric),
	)
	hvue.NewComponent(
		"startup-settings",
		hvue.Template(compStartupSettings),
		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := &struct {
				*js.Object
				ShowTemplateSelect bool `js:"ShowTemplateSelect"`
			}{Object: O()}

			data.ShowTemplateSelect = false
			return data
		}),
		hvue.Method("selectMasterTemplate",
			func(vm *hvue.VM, name *js.Object) {
				promise := vm.Store.Call("dispatch", VUEX_ACTION_SET_STARTUP_MASTER_TEMPLATE_NAME, name)
				promise.Call("then", func(val interface{}) {
					vm.Store.Call("dispatch", VUEX_ACTION_GET_STARTUP_MASTER_TEMPLATE_NAME)
				})
			}),

	)

	hvue.NewComponent(
		"system",
		hvue.Template(compSystem),
		hvue.Method("shutdown",
			func(vm *hvue.VM) {
				vm.Store.Call("dispatch", VUEX_ACTION_SHUTDOWN)
			},
		),
		hvue.Method("reboot",
			func(vm *hvue.VM) {
				vm.Store.Call("dispatch", VUEX_ACTION_REBOOT)
			},
		),
	)
	hvue.NewComponent(
		"master-template",
		hvue.Template(compMasterTemplate),
		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := &struct {
				*js.Object
				//MasterTemplate *jsMasterTemplate `js:"MasterTemplate"`

				ShowTemplateSelectBluetooth bool `js:"ShowTemplateSelectBluetooth"`
				ShowTemplateSelectWiFi bool `js:"ShowTemplateSelectWiFi"`
				ShowTemplateSelectUSB bool `js:"ShowTemplateSelectUSB"`
				ShowTemplateSelectTriggerAction bool `js:"ShowTemplateSelectTriggerAction"`
				ShowTemplateSelectNetwork bool `js:"ShowTemplateSelectNetwork"`

				ShowStoreModal bool   `js:"showStoreModal"`
				ShowLoadModal bool   `js:"showLoadModal"`
				ShowDeployStoredModal bool   `js:"showDeployStoredModal"`


			}{Object: O()}

			//data.MasterTemplate = NewMasterTemplate()
			data.ShowTemplateSelectBluetooth = false
			data.ShowTemplateSelectWiFi = false
			data.ShowTemplateSelectUSB = false
			data.ShowTemplateSelectTriggerAction = false
			data.ShowTemplateSelectNetwork = false

			data.ShowStoreModal = false
			data.ShowLoadModal = false
			data.ShowDeployStoredModal = false


			return data
		}),
		hvue.PropObj("value", hvue.Required),
		hvue.ComputedWithGetSet(
			"MasterTemplate",
			func(vm *hvue.VM) interface{} {
				return vm.Get("value")
			},
			func(vm *hvue.VM, newValue *js.Object) {
				vm.Emit("input", newValue)
			},
		),
		hvue.Method("deploy",
			func(vm *hvue.VM, masterTemplate *jsMasterTemplate) {
				vm.Get("$store").Call("dispatch", VUEX_ACTION_DEPLOY_MASTER_TEMPLATE, masterTemplate)
			}),

		hvue.Method("store",
			func(vm *hvue.VM, name *js.Object) {
				sReq := NewRequestMasterTemplateStorage()
				sReq.TemplateName = name.String()
				sReq.Template = &jsMasterTemplate{
					Object: vm.Get("$store").Get("state").Get("CurrentMasterTemplate"),
				}
				println("Storing :", sReq)
				vm.Get("$store").Call("dispatch", VUEX_ACTION_STORE_MASTER_TEMPLATE, sReq)
				vm.Set("showStoreModal", false)
			}),
		hvue.Method("load",
			func(vm *hvue.VM, name *js.Object) {
				println("Loading :", name.String())
				vm.Get("$store").Call("dispatch", VUEX_ACTION_LOAD_MASTER_TEMPLATE, name)
			}),
		hvue.Method("deleteStored",
			func(vm *hvue.VM, name *js.Object) {
				println("Loading :", name.String())
				vm.Get("$store").Call("dispatch", VUEX_ACTION_DELETE_STORED_MASTER_TEMPLATE, name)
			}),
		hvue.Method("deployStored",
			func(vm *hvue.VM, name *js.Object) {
				println("Loading :", name.String())
				vm.Get("$store").Call("dispatch", VUEX_ACTION_DEPLOY_STORED_MASTER_TEMPLATE, name)
			}),
		hvue.Method("updateStoredSettingsList",
			func(vm *hvue.VM) {
				vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_STORED_MASTER_TEMPLATE_LIST)
			}),
		hvue.Mounted(func(vm *hvue.VM) {
			vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_STORED_MASTER_TEMPLATE_LIST)
		}),

	)
	hvue.NewComponent(
		"database",
		hvue.Template(compDatabase),
		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := &struct {
				*js.Object
				ShowLoad bool `js:"ShowLoad"`
				ShowStore bool `js:"ShowStore"`
			}{Object: O()}

			data.ShowLoad = false
			data.ShowStore = false
			return data
		}),
		hvue.Method("load",
			func(vm *hvue.VM, val *js.Object) {
				vm.Store.Call("dispatch", VUEX_ACTION_RESTORE_DB, val)
			}),
		hvue.Method("store",
			func(vm *hvue.VM, val *js.Object) {
				vm.Store.Call("dispatch", VUEX_ACTION_BACKUP_DB, val)
			}),
		hvue.Method("updateList",
			func(vm *hvue.VM) {
				vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_STORED_DB_BACKUP_LIST)
			}),
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
		<div class="col-12 col-xl">
			<master-template v-model="$store.state.CurrentMasterTemplate" />
		</div>

		<div class="col-12 col-xl">
		<div class="row gutter-sm">
			<div class="col-12 col-xl-6">
				<system />
			</div>

			<div class="col-12 col-xl-6">
				<database />
			</div>

			<div class="col-12">
				<startup-settings />
			</div>
		</div>
		</div>
	</div>
</q-page>
`

const compStartupSettings = `
<q-card>
	<q-card-title>
		启动设置
	</q-card-title>

	<q-card-main>
		<div class="row gutter-sm">
			<q-item tag="div" class="col-12">
				<select-string-from-array :values="$store.state.StoredMasterTemplateList" v-model="ShowTemplateSelect" title="Select Master Template used on startup" @load="selectMasterTemplate($event)"></select-string-from-array>
				<q-item-side icon="whatshot" color primary />
				<q-item-main>
					<q-item-tile label>启动主模板</q-item-tile>
					<q-item-tile sublabel>服务启动时加载的模板</q-item-tile>
					<q-item-tile>
						<div class="row no-wrap">
							<div class="fit">
								<q-input v-model="$store.state.CurrentStartupMasterTemplateName" color="primary" inverted readonly clearable></q-input>
							</div>
							<div><q-btn icon="more" color="primary" @click="ShowTemplateSelect=true" flat /></div>
						</div>
					</q-item-tile>
				</q-item-main>
			</q-item>

		</div>
	</q-card-main>
</q-card>
`

const compMasterTemplate = `
<q-card>
	<q-card-title>
		主模板编辑器
	</q-card-title>

<!--	{{ $data }} -->

	<q-card-main>
		<select-string-from-array :values="$store.state.StoredMasterTemplateList" v-model="showLoadModal" title="加载已保存的主模板" @load="load($event)" @delete="deleteStored($event)" with-delete></select-string-from-array>
		<select-string-from-array :values="$store.state.StoredMasterTemplateList" v-model="showDeployStoredModal" title="应用已保存的主模板" @load="deployStored($event)" @delete="deleteStored($event)" with-delete></select-string-from-array>
		<modal-string-input v-model="showStoreModal" title="保存当前主模板" @save="store($event)"></modal-string-input>


		<div class="row gutter-sm">

			<div class="col-12">
				<div class="row gutter-sm">
					<div class="col-6 col-sm"><q-btn class="fit" color="primary" @click="deploy(MasterTemplate)" label="应用" icon="launch"></q-btn></div>
					<div class="col-6 col-sm"><q-btn class="fit" color="primary" @click="showDeployStoredModal=true" label="应用已保存" icon="settings_backup_restore"></q-btn></div>
					<div class="col-6 col-sm"><q-btn class="fit" color="secondary" @click="showStoreModal=true" label="保存" icon="cloud_upload"></q-btn></div>
					<div class="col-6 col-sm"><q-btn class="fit" color="warning" @click="showLoadModal=true" label="load 加载已保存" icon="cloud_download"></q-btn></div>

				</div>
			</div>


			<!-- TriggerActions template -->
			
			<q-item tag="div" class="col-12">
				<select-string-from-array :values="$store.state.StoredTriggerActionSetsList"  v-model="ShowTemplateSelectTriggerAction" title="选择触发器动作模板" @load="MasterTemplate.TemplateNameTriggerActions=$event"></select-string-from-array>
				<q-item-side icon="whatshot" color primary />
				<q-item-main>
					<q-item-tile label>触发器动作模板</q-item-tile>
<!--
					<q-item-tile sublabel>如果不为空，则选定的触发器动作与主模板一起部署</q-item-tile>
-->
					<q-item-tile>
						<div class="row no-wrap">
							<div class="fit">
								<q-input v-model="MasterTemplate.TemplateNameTriggerActions" color="primary" inverted readonly clearable></q-input>
							</div>
							<div><q-btn icon="more" color="primary" @click="ShowTemplateSelectTriggerAction=true" flat /></div>
							<div><q-btn v-if="MasterTemplate.TemplateNameTriggerActions.length > 0" icon="clear" color="primary" @click="MasterTemplate.TemplateNameTriggerActions=''" flat /></div>
						</div>
					</q-item-tile>
				</q-item-main>
			</q-item>


			<!-- USB template -->
			<q-item tag="div" class="col-12">
				<select-string-from-array :values="$store.state.StoredUSBSettingsList"  v-model="ShowTemplateSelectUSB" title="选择USB模板" @load="MasterTemplate.TemplateNameUSB=$event"></select-string-from-array>
				<q-item-side icon="usb" color primary />
				<q-item-main>
					<q-item-tile label>USB模板</q-item-tile>
<!--
					<q-item-tile sublabel>如果不为空，则会将所选USB设置与主模板一起部署</q-item-tile>
-->
					<q-item-tile>
						<div class="row no-wrap">
							<div class="fit">
								<q-input v-model="MasterTemplate.TemplateNameUSB" color="primary" inverted readonly clearable></q-input>
							</div>
							<div><q-btn icon="more" color="primary" @click="ShowTemplateSelectUSB=true" flat /></div>
							<div><q-btn v-if="MasterTemplate.TemplateNameUSB.length > 0" icon="clear" color="primary" @click="MasterTemplate.TemplateNameUSB=''" flat /></div>
						</div>

					</q-item-tile>
				</q-item-main>
			</q-item>

			<!-- WiFi template -->
			<q-item tag="div" class="col-12">
				<select-string-from-array :values="$store.state.StoredWifiSettingsList"  v-model="ShowTemplateSelectWiFi" title="选择WiFi模板" @load="MasterTemplate.TemplateNameWiFi=$event"></select-string-from-array>
				<q-item-side icon="wifi" color primary />
				<q-item-main>
					<q-item-tile label>WiFi模板</q-item-tile>
<!--
					<q-item-tile sublabel>如果不为空，则将选择的WiFi设置与主模板一起部署</q-item-tile>
-->
					<q-item-tile>
						<div class="row no-wrap">
							<div class="fit">
								<q-input v-model="MasterTemplate.TemplateNameWiFi" color="primary" inverted readonly clearable></q-input>
							</div>
							<div><q-btn icon="more" color="primary" @click="ShowTemplateSelectWiFi=true" flat /></div>
							<div><q-btn v-if="MasterTemplate.TemplateNameWiFi.length > 0" icon="clear" color="primary" @click="MasterTemplate.TemplateNameWiFi=''" flat /></div>
						</div>
					</q-item-tile>
				</q-item-main>
			</q-item>

			<!-- Bluetooth template -->
			<q-item tag="div" class="col-12">
				<select-string-from-array :values="$store.state.StoredBluetoothSettingsList"  v-model="ShowTemplateSelectBluetooth" title="选择蓝牙模板" @load="MasterTemplate.TemplateNameBluetooth=$event"></select-string-from-array>
				<q-item-side icon="bluetooth" color primary />
				<q-item-main>
					<q-item-tile label>蓝牙模板</q-item-tile>
<!--
					<q-item-tile sublabel>如果不为空，则会将选定的蓝牙设置与主模板一起部署</q-item-tile>
-->
					<q-item-tile>
						<div class="row no-wrap">
							<div class="fit">
								<q-input v-model="MasterTemplate.TemplateNameBluetooth" color="primary" inverted readonly clearable></q-input>
							</div>
							<div><q-btn icon="more" color="primary" @click="ShowTemplateSelectBluetooth=true" flat /></div>
							<div><q-btn v-if="MasterTemplate.TemplateNameBluetooth.length > 0" icon="clear" color="primary" @click="MasterTemplate.TemplateNameBluetooth=''" flat /></div>
						</div>
					</q-item-tile>
				</q-item-main>
			</q-item>

			<q-item tag="div" class="col-12">
				<select-network-templates 
					:values="$store.state.StoredEthernetInterfaceSettingsList" 
					:show="ShowTemplateSelectNetwork" 
					@show="ShowTemplateSelectNetwork=$event"
					title="选择网络模板" 
					v-model="MasterTemplate.TemplateNamesNetwork" 
				/>
				<q-item-side icon="settings_ethernet" color primary />
				<q-item-main>
					<q-item-tile label>网络模板</q-item-tile>
					<q-item-tile sublabel>每个接口只能选择一个模板</q-item-tile>
<!--
					<q-item-tile sublabel>If not empty, the selected network templates are deployed along with the master template. Only one template could be selected per interface.</q-item-tile>
-->
					<q-item-tile>
						<div class="row no-wrap">
							<div class="fit">
<!--
								<q-chips-input v-model="MasterTemplate.TemplateNamesNetwork"  color="primary" inverted clearable />
-->
								<q-chips-input v-model="MasterTemplate.TemplateNamesNetwork"  color="primary" inverted readonly />
							</div>
							<div><q-btn icon="more" color="primary" @click="ShowTemplateSelectNetwork=true" flat /></div>
							<div><q-btn v-if="MasterTemplate.TemplateNamesNetwork.length > 0" icon="clear" color="primary" @click="MasterTemplate.TemplateNamesNetwork=[]" flat /></div>
						</div>

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
		系统
	</q-card-title>

	<q-card-main>
		<div class="row gutter-sm">
			<div class="col"> <q-btn class="fit" color="warning" label="重启" icon="refresh" @click="reboot" /> </div>
			<div class="col"> <q-btn class="fit" color="negative" label="关机" icon="power_settings_new" @click="shutdown"/> </div>
		</div>
	</q-card-main>
</q-card>
`

const compDatabase = `
<q-card>

	<select-string-from-array v-model="ShowLoad" :values="$store.state.DBBackupList" title="选择数据库备份" @load="load($event)"></select-string-from-array>
	<modal-string-input v-model="ShowStore" title="保存当前主模板" @save="store($event)"></modal-string-input>

	<q-card-title>
		数据库
	</q-card-title>

	<q-card-main>
		<div class="row gutter-sm">
			<div class="col"> <q-btn class="fit" color="primary" label="备份" icon="cloud_upload" @click="ShowStore=true" /> </div>
			<div class="col"> <q-btn class="fit" color="negative" label="还原" icon="cloud_download" @click="updateList();ShowLoad=true" /> </div>
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
      					每个接口只能选择一个模板
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
