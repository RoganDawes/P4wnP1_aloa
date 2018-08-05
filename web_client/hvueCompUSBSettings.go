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
}

//This becomes a method of the Vue Component and encapsulates dispatching of a Vuex action
func (c *CompUSBSettingsData) UpdateFromDeployedGadgetSettings(vm *hvue.VM) {
	vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_GADGET_SETTINGS_FROM_DEPLOYED)
}

//This becomes a method of the Vue Component and encapsulates dispatching of a Vuex action
func (c *CompUSBSettingsData) ApplyGadgetSettings(vm *hvue.VM) {
	vm.Store.Call("dispatch", VUEX_ACTION_DEPLOY_CURRENT_GADGET_SETTINGS)
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
				return vm.Store.Get("state").Get("currentGadgetSettings")
			}),

	)
}

func newCompUSBSettingsData(vm *hvue.VM) interface{} {

	cc := &CompUSBSettingsData{
		Object: js.Global.Get("Object").New(),
	}
	cc.GadgetSettings = NewUSBGadgetSettings()

	cc.DeployPending = false
	cc.RndisDetails = false
	cc.CdcEcmDetails = false

	return cc
}



const (
	compUSBSettingsTemplate = `
<div>
	<table cellspacing="1">
		<tr>
			<td>USB gadget settings</td>
			<td><button @click="ApplyGadgetSettings" :disabled="deployPending">Apply</button>
			<button @click="UpdateFromDeployedGadgetSettings">Deployed</button></td>
		</tr>
		<tr>
			<td>Gadget enabled</td>
			<td><toggle-switch v-model="currentGadgetSettings.Enabled"></toggle-switch></td>
		</tr>
		<tr>
			<td>Vendor ID</td>
			<td><input v-model="currentGadgetSettings.Vid"/></td>
		</tr>
		<tr>
			<td>Product ID</td>
			<td><input v-model="currentGadgetSettings.Pid"/></td> 
		</tr>
		<tr>
			<td>Manufacturer Name</td>
			<td><input v-model="currentGadgetSettings.Manufacturer"/></td>
		</tr>
		<tr>
			<td>Product Name</td>
			<td><input v-model="currentGadgetSettings.Product"/></td>
		</tr>
		<tr>
			<td>Serial number</td>
			<td><input v-model="currentGadgetSettings.Serial"/></td>
		</tr>
		<tr>
			<td>CDC ECM</td>
			<td>
				<toggle-switch v-model="currentGadgetSettings.Use_CDC_ECM"></toggle-switch>
				<a @click="cdcEcmDetails = !cdcEcmDetails" :class="{ 'toggle-collapse-closed': cdcEcmDetails, 'toggle-collapse-opened': !cdcEcmDetails } ">	</a>
			</td>
		</tr>
		<tr v-if="cdcEcmDetails">
			<td colspan="2"><ethernet-addresses v-bind:settings="currentGadgetSettings.CdcEcmSettings" @hostAddrChange="currentGadgetSettings.CdcEcmSettings.HostAddr=$event" @devAddrChange="currentGadgetSettings.CdcEcmSettings.DevAddr=$event"></ethernet-addresses></td>
		</tr>
		<tr>
			<td>RNDIS</td>
			<td>
				<toggle-switch v-model="currentGadgetSettings.Use_RNDIS"></toggle-switch>
				<a @click="rndisDetails = !rndisDetails" :class="{ 'toggle-collapse-closed': rndisDetails, 'toggle-collapse-opened': !rndisDetails } "></a>
			</td>
		</tr>

		<tr v-if="rndisDetails">
			<td colspan="2"><ethernet-addresses v-bind:settings="currentGadgetSettings.RndisSettings" @hostAddrChange="currentGadgetSettings.RndisSettings.HostAddr=$event" @devAddrChange="currentGadgetSettings.RndisSettings.DevAddr=$event"></ethernet-addresses></td>
		</tr>

		<tr>
			<td>HID Keyboard</td>
			<td><toggle-switch v-model="currentGadgetSettings.Use_HID_KEYBOARD"></toggle-switch></td>
		</tr>
		<tr>
			<td>HID Mouse</td>
			<td><toggle-switch v-model="currentGadgetSettings.Use_HID_MOUSE"></toggle-switch></td>
		</tr>
		<tr>
			<td>HID Raw</td>
			<td><toggle-switch v-model="currentGadgetSettings.Use_HID_RAW"></toggle-switch></td>
		</tr>
		<tr>
			<td>Serial</td>
			<td><toggle-switch v-model="currentGadgetSettings.Use_SERIAL"></toggle-switch></td>
		</tr>
		<tr>
			<td>Mass Storage</td>
			<td><toggle-switch v-model="currentGadgetSettings.Use_UMS"></toggle-switch></td>
		</tr>
	</table>
</div>
`
)
