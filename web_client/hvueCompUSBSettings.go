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

	return data
}



const (
	compUSBSettingsTemplate = `
<q-page padding>
	<select-string-from-array :values="$store.state.StoredUSBSettingsList" v-model="showLoadModal" title="Load USB gadget settings" @load="load($event)"></select-string-from-array>
	<select-string-from-array :values="$store.state.StoredUSBSettingsList" v-model="showDeployStoredModal" title="Deploy stored USB gadget settings" @load="deployStored($event)"></select-string-from-array>
	<modal-string-input v-model="showStoreModal" title="Store current USB gadget Settings" @save="store($event)"></modal-string-input>


	<div class="row gutter-sm">
		<div class="col-12 col-sm-6 col-md-5 col-xl-4">
		<q-card class="full-height">
			<q-card-title>
				Generic USB Gadget Settings
			</q-card-title>
			<q-card-actions>
				<q-btn :loading="deploying" color="primary" @click="ApplyGadgetSettings" label="deploy"></q-btn>
				<q-btn color="secondary" @click="UpdateFromDeployedGadgetSettings" label="reset"></q-btn>
				<q-btn color="primary" @click="showStoreModal=true" label="store"></q-btn>
				<q-btn color="primary" @click="updateStoredSettingsList(); showLoadModal=true" label="load stored"></q-btn>
				<q-btn color="primary" @click="updateStoredSettingsList(); showDeployStoredModal=true" label="deploy stored"></q-btn>


			</q-card-actions>

			<q-alert v-show="deploying" type="warning">If you're connected via Ethernet over USB, you will loose connection during deployment (deadline exceeded error)"</q-alert>

			<q-list link>
				<q-item-separator />

				<q-item tag="label">
					<q-item-side>
						<q-toggle v-model="currentGadgetSettings.Enabled"></q-toggle>
					</q-item-side>
					<q-item-main>
						<q-item-tile label>Enabled</q-item-tile>
						<q-item-tile sublabel>Enable/Disable USB gadget (if enabled, at least one function has to be turned on)</q-item-tile>
					</q-item-main>
				</q-item>
				<q-item tag="label">
					<q-item-main>
						<q-item-tile label>Vendor ID</q-item-tile>
						<q-item-tile sublabel>Example: 0x1d6b</q-item-tile>
						<q-item-tile>
							<q-input v-model="currentGadgetSettings.Vid" inverted></q-input>
						</q-item-tile>
					</q-item-main>
				</q-item>
				<q-item tag="label">
					<q-item-main>
						<q-item-tile label>Product ID</q-item-tile>
						<q-item-tile sublabel>Example: 0x1337</q-item-tile>
						<q-item-tile>
							<q-input v-model="currentGadgetSettings.Pid" inverted></q-input>
						</q-item-tile>
					</q-item-main>
				</q-item>
				<q-item tag="label">
					<q-item-main>
						<q-item-tile label>Manufacturer Name</q-item-tile>
						<q-item-tile sublabel></q-item-tile>
						<q-item-tile>
							<q-input v-model="currentGadgetSettings.Manufacturer" inverted></q-input>
						</q-item-tile>
					</q-item-main>
				</q-item>
				<q-item tag="label">
					<q-item-main>
						<q-item-tile label>Product Name</q-item-tile>
						<q-item-tile sublabel></q-item-tile>
						<q-item-tile>
							<q-input v-model="currentGadgetSettings.Product" inverted></q-input>
						</q-item-tile>
					</q-item-main>
				</q-item>
				<q-item tag="label">
					<q-item-main>
						<q-item-tile label>Serial Number</q-item-tile>
						<q-item-tile sublabel></q-item-tile>
						<q-item-tile>
							<q-input v-model="currentGadgetSettings.Serial" inverted></q-input>
						</q-item-tile>
					</q-item-main>
				</q-item>
			</q-list>
		</q-card>
		</div>
		<div class="col-12 col-sm-6 col-md-5 col-xl-4">
		<q-card class="full-height">
			<q-card-title>
				USB Gadget Functions
			</q-card-title>
			<q-list link>

				<q-item-separator />

				<q-item tag="label">
					<q-item-side>
						<q-toggle v-model="currentGadgetSettings.Use_CDC_ECM"></q-toggle>
					</q-item-side>
					<q-item-main>
						<q-item-tile label>CDC ECM</q-item-tile>
						<q-item-tile sublabel>Ethernet over USB for Linux, Unix and OSX</q-item-tile>
					</q-item-main>
				</q-item>


				<q-collapsible icon="settings_ethernet" label="MAC addresses for CDC ECM" v-show="currentGadgetSettings.Use_CDC_ECM" indent>
					<q-item tag="label" indent>
						<q-item-main>
							<q-item-tile label>Host Address</q-item-tile>
							<q-item-tile sublabel>MAC of USB adapter on remote host (format: AA:BB:CC:DD:EE:FF)</q-item-tile>
							<q-item-tile>
								<q-input v-model="currentGadgetSettings.CdcEcmSettings.HostAddr" inverted></q-input>
							</q-item-tile>
						</q-item-main>
					</q-item>
					<q-item tag="label" indent>
						<q-item-main>
							<q-item-tile label>Device Address</q-item-tile>
							<q-item-tile sublabel>MAC address on P4wnP1's end (format: AA:BB:CC:DD:EE:FF)</q-item-tile>
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
						<q-item-tile sublabel>Ethernet over USB for Windows (and some Linux kernels)</q-item-tile>
					</q-item-main>
				</q-item>

				<q-collapsible icon="settings_ethernet" label="MAC addresses for RNDIS" v-show="currentGadgetSettings.Use_RNDIS" indent>
					<q-item tag="label" ident>
						<q-item-main>
							<q-item-tile label>Host Address</q-item-tile>
							<q-item-tile sublabel>MAC of USB adapter on remote host - could get overwritten by host (format: AA:BB:CC:DD:EE:FF)</q-item-tile>
							<q-item-tile>
								<q-input v-model="currentGadgetSettings.RndisSettings.HostAddr" inverted></q-input>
							</q-item-tile>
						</q-item-main>
					</q-item>
					<q-item tag="label" ident>
						<q-item-main>
							<q-item-tile label>Device Address</q-item-tile>
							<q-item-tile sublabel>MAC address on P4wnP1's end (format: AA:BB:CC:DD:EE:FF)</q-item-tile>
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
						<q-item-tile label>Keyboard</q-item-tile>
						<q-item-tile sublabel>HID Keyboard functionality (needed for HID Script)</q-item-tile>
					</q-item-main>
				</q-item>
				<q-item tag="label">
					<q-item-side>
						<q-toggle v-model="currentGadgetSettings.Use_HID_MOUSE"></q-toggle>
					</q-item-side>
					<q-item-main>
						<q-item-tile label>Mouse</q-item-tile>
						<q-item-tile sublabel>HID Mouse functionality (needed for HID Script)</q-item-tile>
					</q-item-main>
				</q-item>
				<q-item tag="label">
					<q-item-side>
						<q-toggle v-model="currentGadgetSettings.Use_HID_RAW"></q-toggle>
					</q-item-side>
					<q-item-main>
						<q-item-tile label>Custom HID device</q-item-tile>
						<q-item-tile sublabel>Raw HID device function, used for covert channel</q-item-tile>
					</q-item-main>
				</q-item>
				<q-item tag="label">
					<q-item-side>
						<q-toggle v-model="currentGadgetSettings.Use_Serial"></q-toggle>
					</q-item-side>
					<q-item-main>
						<q-item-tile label>Serial Interface</q-item-tile>
						<q-item-tile sublabel>Provides a serial port over USB</q-item-tile>
					</q-item-main>
				</q-item>
				<q-item tag="label">
					<q-item-side>
						<q-toggle v-model="currentGadgetSettings.Use_UMS"></q-toggle>
					</q-item-side>
					<q-item-main>
						<q-item-tile label>Mass Storage</q-item-tile>
						<q-item-tile sublabel>Emulates USB flash drive or CD-ROM</q-item-tile>
					</q-item-main>
				</q-item>
			</q-list>	
		</q-card>
		</div>
	
	</div>
</q-page>

`
)
