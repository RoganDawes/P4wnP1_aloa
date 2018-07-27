package main

import (
	"github.com/gopherjs/gopherjs/js"
	pb "../proto/gopherjs"
	"time"
	"context"
	"google.golang.org/grpc/status"
	"github.com/mame82/hvue"
)


type CompUSBSettingsData struct {
	*js.Object

	GadgetSettings *jsGadgetSettings `js:"gadgetSettings"`
	DeployPending bool               `js:"deployPending"`
	CdcEcmDetails bool               `js:"cdcEcmDetails"`
	RndisDetails bool                `js:"rndisDetails"`
}

//ToDo: Reimplement with Action on global state
func (c *CompUSBSettingsData) UpdateFromDeployedGadgetSettings(vm *hvue.VM) {
	vm.Store.Call("commit", "setCurrentGadgetSettingsFromDeployed")
}

//ToDo: Reimplement with actions on global state
func (c *CompUSBSettingsData) ApplyGadgetSettings(vm *hvue.VM) {
	//println("Trying to deploy GadgetSettings: " + fmt.Sprintf("%+v",c.GadgetSettings.toGS()))
	println("Trying to deploy GadgetSettings...")
	//gs:=c.GadgetSettings.toGS()
	gs := jsGadgetSettings{Object: vm.Store.Get("state").Get("currentGadgetSettings")}.toGS()

	go func() {
		c.DeployPending = true
		defer func() {c.DeployPending = false}()

		ctx,cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		//Set gadget settings
		_, err := Client.Client.SetGadgetSettings(ctx, gs)
		if err != nil {
			js.Global.Call("alert", "Error setting given gadget settings: " + status.Convert(err).Message())
			println(err)
			c.UpdateFromDeployedGadgetSettings(vm)
			return
		}
		println("New GadgetSettings have been set")



		//deploy the settings
		deployedGs,err := Client.Client.DeployGadgetSetting(ctx, &pb.Empty{})
		if err != nil {
			js.Global.Call("alert", "Error deploying gadget settings: " + status.Convert(err).Message())
			println(err)
			c.UpdateFromDeployedGadgetSettings(vm)
			return
		}
		println("New GadgetSettings have been deployed")

		js.Global.Call("alert", "New USB gadget settings deployed ")

		newGs := &jsGadgetSettings{
			Object: js.Global.Get("Object").New(),
		}
		newGs.fromGS(deployedGs)
		c.GadgetSettings = newGs
	}()
}

func InitCompUSBSettings() {

	hvue.NewComponent(
		"usb-settings",
		hvue.Template(compUSBSettingsTemplate),
		hvue.DataFunc(newCompUSBSettingsData),
		hvue.MethodsOf(&CompUSBSettingsData{}),
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

	cc.UpdateFromDeployedGadgetSettings(vm)
	cc.DeployPending = false
	cc.RndisDetails = false
	cc.CdcEcmDetails = false

	return cc
}



const (
	compUSBSettingsTemplate = `
<div>
	<table>
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
			<td><toggle-switch v-model="currentGadgetSettings.Use_CDC_ECM"></toggle-switch></td>
		</tr>
		<tr v-if="currentGadgetSettings.Use_CDC_ECM">
			<td></td>
			<td><ethernet-addresses v-bind:settings="currentGadgetSettings.CdcEcmSettings" @hostAddrChange="currentGadgetSettings.CdcEcmSettings.HostAddr=$event" @devAddrChange="currentGadgetSettings.CdcEcmSettings.DevAddr=$event"></ethernet-addresses></td>
		</tr>
		<tr>
			<td>RNDIS</td>
			<td><toggle-switch v-model="currentGadgetSettings.Use_RNDIS"></toggle-switch></td>
			<td><input type="checkbox" v-if="currentGadgetSettings.Use_RNDIS" v-model="rndisDetails"></td>
		</tr>
		<tr v-if="rndisDetails">
			<td></td>
			<td><ethernet-addresses v-bind:settings="currentGadgetSettings.RndisSettings" @hostAddrChange="currentGadgetSettings.RndisSettings.HostAddr=$event" @devAddrChange="currentGadgetSettings.RndisSettings.DevAddr=$event"></ethernet-addresses></td>
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

