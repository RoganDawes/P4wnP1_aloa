// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/HuckRidgeSW/hvue"
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
	vm.Get("$store").Call("dispatch", VUEX_ACTION_UPDATE_GADGET_SETTINGS_FROM_DEPLOYED)
}

//This becomes a method of the Vue Component and encapsulates dispatching of a Vuex action
func (c *CompUSBSettingsData) ApplyGadgetSettings(vm *hvue.VM) {
	vm.Get("$store").Call("dispatch", VUEX_ACTION_DEPLOY_CURRENT_GADGET_SETTINGS)
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
<q-page class="row justify gutter-sm">
<div>
	<q-btn :loading="deploying" color="primary" @click="ApplyGadgetSettings" label="deploy"></q-btn>
	<q-btn color="secondary" @click="UpdateFromDeployedGadgetSettings" label="reset"></q-btn>
	<br><br>
	
	<q-list link>
		<q-list-header>Generic Gadget Settings</q-list-header>
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

 		<q-item-separator />

		<q-list-header>Gadget functions</q-list-header>
		<q-item tag="label">
			<q-item-side>
				<q-toggle v-model="currentGadgetSettings.Use_CDC_ECM"></q-toggle>
			</q-item-side>
			<q-item-main>
				<q-item-tile label>CDC ECM</q-item-tile>
				<q-item-tile sublabel>Ethernet over USB for Linux, Unix and OSX</q-item-tile>
			</q-item-main>
		</q-item>
		<q-item tag="label">
			<q-item-side>
				<q-toggle v-model="currentGadgetSettings.Use_RNDIS"></q-toggle>
			</q-item-side>
			<q-item-main>
				<q-item-tile label>RNDIS</q-item-tile>
				<q-item-tile sublabel>Ethernet over USB for Windows (and some Linux kernels)</q-item-tile>
			</q-item-main>
		</q-item>
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
	<q-list>	
</div>
	
	
	
<!--
	<table cellspacing="1">
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
-->
</q-page>
`
)
