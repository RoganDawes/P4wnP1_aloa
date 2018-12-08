// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/mame82/hvue"
)


type CompUSBSettingsData struct {
	*js.Object

	GadgetSettings *jsGadgetSettings `js:"gadgetSettings"`
	DeployPending bool               `js:"deployPending"`
	CdcEcmDetails bool               `js:"cdcEcmDetails"`
	RndisDetails bool                `js:"rndisDetails"`
	ShowStoreModal bool   `js:"showStoreModal"`
	ShowLoadModal bool   `js:"showLoadModal"`
	ShowDeployStoredModal bool   `js:"showDeployStoredModal"`
	ShowUmsModal bool   `js:"ShowUmsModal"`

}

//This becomes a method of the Vue Component and encapsulates dispatching of a Vuex action
func (c *CompUSBSettingsData) UpdateFromDeployedGadgetSettings(vm *hvue.VM) {
	vm.Get("$store").Call("dispatch", VUEX_ACTION_UPDATE_CURRENT_USB_SETTINGS)
}

//This becomes a method of the Vue Component and encapsulates dispatching of a Vuex action
func (c *CompUSBSettingsData) ApplyGadgetSettings(vm *hvue.VM) {
	vm.Get("$store").Call("dispatch", VUEX_ACTION_DEPLOY_CURRENT_USB_SETTINGS)
}

func InitCompUSBSettings() {
	hvue.NewComponent(
		"usb-settings",
		hvue.Template(compUSBSettingsTemplate),
		hvue.DataFunc(newCompUSBSettingsData),
		hvue.MethodsOf(&CompUSBSettingsData{}), // Add the methods of CompUSBSettingsData to the Vue Component instance
		hvue.Computed(
			"currentGadgetSettings",
			func(vm *hvue.VM) interface{} {
				return vm.Get("$store").Get("state").Get("currentGadgetSettings")
			}),
		hvue.Computed("deploying",
			func(vm *hvue.VM) interface{} {
				return vm.Get("$store").Get("state").Get("deployingGadgetSettings")
			}),
		hvue.Mounted(func(vm *hvue.VM) {
			vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_CURRENT_USB_SETTINGS)
		}),
		hvue.Method("store",
			func(vm *hvue.VM, name *js.Object) {
				sReq := NewUSBRequestSettingsStorage()
				sReq.TemplateName = name.String()
				sReq.Settings = &jsGadgetSettings{
					Object: vm.Get("$store").Get("state").Get("currentGadgetSettings"),
				}
				println("Storing :", sReq)
				vm.Get("$store").Call("dispatch", VUEX_ACTION_STORE_USB_SETTINGS, sReq)
				vm.Set("showStoreModal", false)
			}),
		hvue.Method("load",
			func(vm *hvue.VM, name *js.Object) {
				println("Loading :", name.String())
				vm.Get("$store").Call("dispatch", VUEX_ACTION_LOAD_USB_SETTINGS, name)
			}),
		hvue.Method("deleteStored",
			func(vm *hvue.VM, name *js.Object) {
				println("Loading :", name.String())
				vm.Get("$store").Call("dispatch", VUEX_ACTION_DELETE_STORED_USB_SETTINGS, name)
			}),
		hvue.Method("deployStored",
			func(vm *hvue.VM, name *js.Object) {
				println("Loading :", name.String())
				vm.Get("$store").Call("dispatch", VUEX_ACTION_DEPLOY_STORED_USB_SETTINGS, name)
			}),
		hvue.Method("updateStoredSettingsList",
			func(vm *hvue.VM) {
				vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_STORED_USB_SETTINGS_LIST)
			}),

	)

	hvue.NewComponent("ums-settings",
		hvue.Template(compUSBUmsSettingsTemplate),
		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := &struct {
				*js.Object
				ShowImageSelect bool `js:"ShowImageSelect"`
			}{Object: O()}
			data.ShowImageSelect = false
			return data
		}),
		hvue.PropObj("value"),
		hvue.PropObj(
			"show",
			hvue.Required,
			hvue.Types(hvue.PBoolean),
		),
		hvue.ComputedWithGetSet(
			"visible",
			func(vm *hvue.VM) interface{} {
				return vm.Get("show")
			},
			func(vm *hvue.VM, newValue *js.Object) {
				vm.Call("$emit", "show", newValue)
			},
		),
		hvue.Method("updateFileLists",
			func(vm *hvue.VM) {
				vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_UMS_IMAGE_CDROM_LIST)
				vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_UMS_IMAGE_FLASHDRIVE_LIST)
			}),
		hvue.Mounted(func(vm *hvue.VM) {
			vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_UMS_IMAGE_CDROM_LIST)
			vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_UMS_IMAGE_FLASHDRIVE_LIST)
		}),
	)
}

func newCompUSBSettingsData(vm *hvue.VM) interface{} {

	data := &CompUSBSettingsData{
		Object: js.Global.Get("Object").New(),
	}
	data.GadgetSettings = NewUSBGadgetSettings()

	data.ShowStoreModal = false
	data.ShowLoadModal = false
	data.ShowDeployStoredModal = false

	data.DeployPending = false
	data.RndisDetails = false
	data.CdcEcmDetails = false

	data.ShowUmsModal = false

	return data
}



const (
	compUSBUmsSettingsTemplate = `
	<q-modal v-model="visible">
		<q-modal-layout>
			<q-toolbar slot="header">
				<q-toolbar-title>
					大容量USB存储
				</q-toolbar-title>
			</q-toolbar>

			<q-list>
				<q-item tag="label">
					<q-item-side>
						<q-toggle v-model="value.Cdrom" @input="value.File=''"></q-toggle>
					</q-item-side>
					<q-item-main>
						<q-item-tile label>CD-Rom</q-item-tile>
						<q-item-tile sublabel>如果启用，则模拟CD-ROM驱动器而不是可写的闪存驱动器</q-item-tile>
					</q-item-main>
				</q-item>


			<q-item tag="div" class="col-12">
				<select-string-from-array :values="value.Cdrom ? $store.state.UmsImageListCdrom : $store.state.UmsImageListFlashdrive"  v-model="ShowImageSelect" title="Select image" @load="value.File=$event"></select-string-from-array>
				<q-item-side icon="archive" color primary />
				<q-item-main>
					<q-item-tile label>要使用的镜像文件</q-item-tile>
<!--
					<q-item-tile sublabel>如果不为空，则选定的触发器动作与主模板一起部署</q-item-tile>
-->
					<q-item-tile>
						<div class="row no-wrap">
							<div class="fit">
								<q-input v-model="value.File" color="primary" inverted readonly clearable></q-input>
							</div>
							<div><q-btn icon="more" color="primary" @click="updateFileLists();ShowImageSelect=true" flat /></div>
							<div><q-btn v-if="value.File.length > 0" icon="clear" color="primary" @click="value.File=''" flat /></div>
						</div>
					</q-item-tile>
				</q-item-main>
			</q-item>



			</q-list>


			<q-list slot="footer">
				<q-item tag="label">
					<q-item-main>
						<q-item-tile>
							<q-btn color="secondary" v-close-overlay label="关闭" />
						</q-item-tile>
					</q-item-main>
				</q-item>
			</q-list>

		</q-modal-layout>
	</q-modal>

`
	compUSBSettingsTemplate = `
<q-page padding>
	<ums-settings :show="ShowUmsModal" @show="ShowUmsModal=$event" v-model="currentGadgetSettings.UmsSettings" />

	<select-string-from-array :values="$store.state.StoredUSBSettingsList" v-model="showLoadModal" title="加载USB工具设置" @load="load($event)" @delete="deleteStored($event)" with-delete></select-string-from-array>
	<select-string-from-array :values="$store.state.StoredUSBSettingsList" v-model="showDeployStoredModal" title="应用已保存的USB工具设置" @load="deployStored($event)" @delete="deleteStored($event)" with-delete></select-string-from-array>
	<modal-string-input v-model="showStoreModal" title="保存当前USB工具设置" @save="store($event)"></modal-string-input>


	<div class="row gutter-sm">
		<div class="col-12">
			<q-card>
				<q-card-title>
					USB工具设置
				</q-card-title>

				<q-card-main>
					<div class="row gutter-sm">

						<div class="col-6 col-sm"><q-btn class="fit" :loading="deploying" color="primary" @click="ApplyGadgetSettings" label="应用" icon="launch"></q-btn></div>
						<div class="col-6 col-sm"><q-btn class="fit" color="primary" @click="updateStoredSettingsList(); showDeployStoredModal=true" label="应用已保存" icon="settings_backup_restore"></q-btn></div>
						<div class="col-6 col-sm"><q-btn class="fit" color="secondary" @click="UpdateFromDeployedGadgetSettings" label="重置" icon="autorenew"></q-btn></div>
						<div class="col-6 col-sm"><q-btn class="fit" color="secondary" @click="showStoreModal=true" label="保存" icon="cloud_upload"></q-btn></div>
						<div class="col-12 col-sm"><q-btn class="fit" color="warning" @click="updateStoredSettingsList(); showLoadModal=true" label="加载保存" icon="cloud_download"></q-btn></div>

					</div>
  				</q-card-main>


			</q-card>
		</div>


		<div class="col-12 col-lg">
		<q-card class="full-height">
			<q-alert v-show="deploying" type="warning">如果您通过USB通过以太网连接，则在部署期间将断开连接(截止日期超出错误)"</q-alert>
			<q-list link>
				<q-item tag="label">
					<q-item-side>
						<q-toggle v-model="currentGadgetSettings.Enabled"></q-toggle>
					</q-item-side>
					<q-item-main>
						<q-item-tile label>已启用</q-item-tile>
						<q-item-tile sublabel>启用/禁用USB小工具(如果启用，则必须至少打开一个功能)</q-item-tile>
					</q-item-main>
				</q-item>
				<q-item tag="label">
					<q-item-main>
						<q-item-tile label>Vendor ID</q-item-tile>
						<q-item-tile sublabel>示例: 0x1d6b</q-item-tile>
						<q-item-tile>
							<q-input v-model="currentGadgetSettings.Vid" inverted></q-input>
						</q-item-tile>
					</q-item-main>
				</q-item>
				<q-item tag="label">
					<q-item-main>
						<q-item-tile label>Product ID</q-item-tile>
						<q-item-tile sublabel>示例: 0x1337</q-item-tile>
						<q-item-tile>
							<q-input v-model="currentGadgetSettings.Pid" inverted></q-input>
						</q-item-tile>
					</q-item-main>
				</q-item>
				<q-item tag="label">
					<q-item-main>
						<q-item-tile label>制造商名称</q-item-tile>
						<q-item-tile sublabel></q-item-tile>
						<q-item-tile>
							<q-input v-model="currentGadgetSettings.Manufacturer" inverted></q-input>
						</q-item-tile>
					</q-item-main>
				</q-item>
				<q-item tag="label">
					<q-item-main>
						<q-item-tile label>产品名称</q-item-tile>
						<q-item-tile sublabel></q-item-tile>
						<q-item-tile>
							<q-input v-model="currentGadgetSettings.Product" inverted></q-input>
						</q-item-tile>
					</q-item-main>
				</q-item>
				<q-item tag="label">
					<q-item-main>
						<q-item-tile label>序列号</q-item-tile>
						<q-item-tile sublabel></q-item-tile>
						<q-item-tile>
							<q-input v-model="currentGadgetSettings.Serial" inverted></q-input>
						</q-item-tile>
					</q-item-main>
				</q-item>
			</q-list>
		</q-card>
		</div>
		<div class="col-12 col-lg">
		<q-card class="full-height">
			<q-list link>
				<q-item tag="label">
					<q-item-side>
						<q-toggle v-model="currentGadgetSettings.Use_CDC_ECM"></q-toggle>
					</q-item-side>
					<q-item-main>
						<q-item-tile label>CDC ECM</q-item-tile>
						<q-item-tile sublabel>适用于Linux，Unix和OSX的USB以太网</q-item-tile>
					</q-item-main>
				</q-item>


				<q-collapsible icon="settings_ethernet" label="MAC addresses for CDC ECM" v-show="currentGadgetSettings.Use_CDC_ECM" indent>
					<q-item tag="label" indent>
						<q-item-main>
							<q-item-tile label>主机地址</q-item-tile>
							<q-item-tile sublabel>远程主机上USB适配器的MAC地址(格式:AA:BB:CC:DD:EE:FF)</q-item-tile>
							<q-item-tile>
								<q-input v-model="currentGadgetSettings.CdcEcmSettings.HostAddr" inverted></q-input>
							</q-item-tile>
						</q-item-main>
					</q-item>
					<q-item tag="label" indent>
						<q-item-main>
							<q-item-tile label>设备地址</q-item-tile>
							<q-item-tile sublabel>P4wnP1端的MAC地址(格式:AA:BB:CC:DD:EE:FF)</q-item-tile>
							<q-item-tile>
								<q-input v-model="currentGadgetSettings.CdcEcmSettings.DevAddr" inverted></q-input>
							</q-item-tile>
						</q-item-main>
					</q-item>
				</q-collapsible>

				<q-item tag="label">
					<q-item-side>
						<q-toggle v-model="currentGadgetSettings.Use_RNDIS"></q-toggle>
					</q-item-side>
					<q-item-main>
						<q-item-tile label>RNDIS</q-item-tile>
						<q-item-tile sublabel>适用于Windows的以太网USB(以及一些Linux内核)</q-item-tile>
					</q-item-main>
				</q-item>

				<q-collapsible icon="settings_ethernet" label="RNDIS的MAC地址" v-show="currentGadgetSettings.Use_RNDIS" indent>
					<q-item tag="label" ident>
						<q-item-main>
							<q-item-tile label>主机地址</q-item-tile>
							<q-item-tile sublabel>远程主机上USB适配器的MAC地址-可能被主机覆盖(格式:AA:BB:CC:DD:EE:FF)</q-item-tile>
							<q-item-tile>
								<q-input v-model="currentGadgetSettings.RndisSettings.HostAddr" inverted></q-input>
							</q-item-tile>
						</q-item-main>
					</q-item>
					<q-item tag="label" ident>
						<q-item-main>
							<q-item-tile label>设备地址</q-item-tile>
							<q-item-tile sublabel>P4wnP1端的MAC地址(格式:AA:BB:CC:DD:EE:FF)</q-item-tile>
							<q-item-tile>
								<q-input v-model="currentGadgetSettings.RndisSettings.DevAddr" inverted></q-input>
							</q-item-tile>
						</q-item-main>
					</q-item>
				</q-collapsible>


				<q-item tag="label">
					<q-item-side>
						<q-toggle v-model="currentGadgetSettings.Use_HID_KEYBOARD"></q-toggle>
					</q-item-side>
					<q-item-main>
						<q-item-tile label>键盘</q-item-tile>
						<q-item-tile sublabel>HID键盘功能(HID脚本需要)</q-item-tile>
					</q-item-main>
				</q-item>
				<q-item tag="label">
					<q-item-side>
						<q-toggle v-model="currentGadgetSettings.Use_HID_MOUSE"></q-toggle>
					</q-item-side>
					<q-item-main>
						<q-item-tile label>鼠标</q-item-tile>
						<q-item-tile sublabel>HID鼠标功能(HID脚本需要)</q-item-tile>
					</q-item-main>
				</q-item>
				<q-item tag="label">
					<q-item-side>
						<q-toggle v-model="currentGadgetSettings.Use_HID_RAW"></q-toggle>
					</q-item-side>
					<q-item-main>
						<q-item-tile label>自定义HID设备</q-item-tile>
						<q-item-tile sublabel>原始HID设备功能,用于隐蔽通道</q-item-tile>
					</q-item-main>
				</q-item>
				<q-item tag="label">
					<q-item-side>
						<q-toggle v-model="currentGadgetSettings.Use_SERIAL"></q-toggle>
					</q-item-side>
					<q-item-main>
						<q-item-tile label>串行接口</q-item-tile>
						<q-item-tile sublabel>通过USB提供串行端口</q-item-tile>
					</q-item-main>
				</q-item>
				<q-item tag="label">
					<q-item-side>
						<q-toggle v-model="currentGadgetSettings.Use_UMS"></q-toggle>
					</q-item-side>
					<q-item-main>
						<q-item-tile label>大容量存储</q-item-tile>
						<q-item-tile sublabel>模拟USB闪存驱动器或CD-ROM</q-item-tile>
					</q-item-main>
					<q-item-side right v-if="currentGadgetSettings.Use_UMS">
						<div><q-btn icon="more" color="primary" flat @click="ShowUmsModal=true" /></div>
					</q-item-side>

				</q-item>
			</q-list>	
		</q-card>
		</div>
	
	</div>
</q-page>

`
)
