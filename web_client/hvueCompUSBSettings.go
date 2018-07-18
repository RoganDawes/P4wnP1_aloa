package main

import (
	"github.com/gopherjs/gopherjs/js"
	pb "../proto/gopherjs"
	"time"
	"context"
	"google.golang.org/grpc/status"
	"github.com/HuckRidgeSW/hvue"
)


type VGadgetSettings struct {
	*js.Object
	Enabled          bool  `js:"Enabled"`
	Vid              string  `js:"Vid"`
	Pid              string  `js:"Pid"`
	Manufacturer     string `js:"Manufacturer"`
	Product          string `js:"Product"`
	Serial           string `js:"Serial"`
	Use_CDC_ECM      bool `js:"Use_CDC_ECM"`
	Use_RNDIS        bool `js:"Use_RNDIS"`
	Use_HID_KEYBOARD bool `js:"Use_HID_KEYBOARD"`
	Use_HID_MOUSE    bool `js:"Use_HID_MOUSE"`
	Use_HID_RAW      bool `js:"Use_HID_RAW"`
	Use_UMS          bool `js:"Use_UMS"`
	Use_SERIAL       bool `js:"Use_SERIAL"`
	RndisSettings    *VGadgetSettingsEthernet `js:"RndisSettings"`
	CdcEcmSettings   *VGadgetSettingsEthernet `js:"CdcEcmSettings"`
	UmsSettings      *VGadgetSettingsUMS `js:"UmsSettings"`
}

type VGadgetSettingsEthernet struct {
	*js.Object
	HostAddr string `js:"HostAddr"`
	DevAddr  string `js:"DevAddr"`
}


type VGadgetSettingsUMS struct {
	*js.Object
	Cdrom bool `js:"Cdrom"`
	File  string `js:"File"`
}

func (vGS VGadgetSettings) toGS() (gs *pb.GadgetSettings) {
	return &pb.GadgetSettings{
		Serial: vGS.Serial,
		Use_SERIAL: vGS.Use_SERIAL,
		Use_UMS: vGS.Use_UMS,
		Use_HID_RAW: vGS.Use_HID_RAW,
		Use_HID_MOUSE: vGS.Use_HID_MOUSE,
		Use_HID_KEYBOARD: vGS.Use_HID_KEYBOARD,
		Use_RNDIS: vGS.Use_RNDIS,
		Use_CDC_ECM: vGS.Use_CDC_ECM,
		Product: vGS.Product,
		Manufacturer: vGS.Manufacturer,
		Vid: vGS.Vid,
		Pid: vGS.Pid,
		Enabled: vGS.Enabled,
		UmsSettings: &pb.GadgetSettingsUMS{
			Cdrom: vGS.UmsSettings.Cdrom,
			File: vGS.UmsSettings.File,
		},
		CdcEcmSettings: &pb.GadgetSettingsEthernet{
			DevAddr: vGS.CdcEcmSettings.DevAddr,
			HostAddr: vGS.CdcEcmSettings.HostAddr,
		},
		RndisSettings: &pb.GadgetSettingsEthernet{
			DevAddr: vGS.RndisSettings.DevAddr,
			HostAddr: vGS.RndisSettings.HostAddr,
		},
	}
}

func (vGS *VGadgetSettings) fromGS(gs *pb.GadgetSettings) {
	println(gs)

	vGS.Enabled = gs.Enabled
	vGS.Vid = gs.Vid
	vGS.Pid = gs.Pid
	vGS.Manufacturer = gs.Manufacturer
	vGS.Product = gs.Product
	vGS.Serial = gs.Serial
	vGS.Use_CDC_ECM = gs.Use_CDC_ECM
	vGS.Use_RNDIS = gs.Use_RNDIS
	vGS.Use_HID_KEYBOARD = gs.Use_HID_KEYBOARD
	vGS.Use_HID_MOUSE = gs.Use_HID_MOUSE
	vGS.Use_HID_RAW = gs.Use_HID_RAW
	vGS.Use_UMS = gs.Use_UMS
	vGS.Use_SERIAL = gs.Use_SERIAL

	vGS.RndisSettings = &VGadgetSettingsEthernet{
		Object: js.Global.Get("Object").New(),
	}
	if gs.RndisSettings != nil {
		vGS.RndisSettings.HostAddr = gs.RndisSettings.HostAddr
		vGS.RndisSettings.DevAddr = gs.RndisSettings.DevAddr
	}

	vGS.CdcEcmSettings = &VGadgetSettingsEthernet{
		Object: js.Global.Get("Object").New(),
	}
	if gs.CdcEcmSettings != nil {
		vGS.CdcEcmSettings.HostAddr = gs.CdcEcmSettings.HostAddr
		vGS.CdcEcmSettings.DevAddr = gs.CdcEcmSettings.DevAddr
	}

	vGS.UmsSettings = &VGadgetSettingsUMS{
		Object: js.Global.Get("Object").New(),
	}
	if gs.UmsSettings != nil {
		vGS.UmsSettings.File = gs.UmsSettings.File
		vGS.UmsSettings.Cdrom = gs.UmsSettings.Cdrom
	}
}

// Note: internalize wouldn't work on this, as the nested structs don't translate back
type CompUSBSettingsData struct {
	*js.Object

	GadgetSettings *VGadgetSettings `js:"gadgetSettings"`
	DeployPending bool `js:"deployPending"`
	CdcEcmDetails bool `js:"cdcEcmDetails"`
	RndisDetails bool `js:"rndisDetails"`
}


func (c *CompUSBSettingsData) UpdateToDeployedGadgetSettings(vm *hvue.VM) {
	//gs := vue.GetVM(c).Get("gadgetSettings")
	println("Trying to fetch deployed GadgetSettings")

	go func() {
		ctx,cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		deployedGs, err := pb.NewP4WNP1Client(serverAddr).GetDeployedGadgetSetting(ctx, &pb.Empty{})
		if err != nil { println(err); return }

		newGs := &VGadgetSettings{
			Object: js.Global.Get("Object").New(),
		}
		newGs.fromGS(deployedGs)
		c.GadgetSettings = newGs
	}()
}

func (c *CompUSBSettingsData) ApplyGadgetSettings(vm *hvue.VM) {
	//println("Trying to deploy GadgetSettings: " + fmt.Sprintf("%+v",c.GadgetSettings.toGS()))
	println("Trying to deploy GadgetSettings...")
	gs:=c.GadgetSettings.toGS()
	go func() {
		c.DeployPending = true
		defer func() {c.DeployPending = false}()

		ctx,cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		client := pb.NewP4WNP1Client(serverAddr)

		//Set gadget settings
		_, err := client.SetGadgetSettings(ctx, gs)
		if err != nil {
			js.Global.Call("alert", "Error setting given gadget settings: " + status.Convert(err).Message())
			println(err)
			c.UpdateToDeployedGadgetSettings(vm)
			return
		}
		println("New GadgetSettings have been set")



		//deploy the settings
		deployedGs,err := client.DeployGadgetSetting(ctx, &pb.Empty{})
		if err != nil {
			js.Global.Call("alert", "Error deploying gadget settings: " + status.Convert(err).Message())
			println(err)
			c.UpdateToDeployedGadgetSettings(vm)
			return
		}
		println("New GadgetSettings have been deployed")

		js.Global.Call("alert", "New USB gadget settings deployed ")

		newGs := &VGadgetSettings{
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
	)
}

func newCompUSBSettingsData(vm *hvue.VM) interface{} {

	cc := &CompUSBSettingsData{
		Object: js.Global.Get("Object").New(),
	}

	cc.GadgetSettings = &VGadgetSettings{
		Object: js.Global.Get("Object").New(),
	}
	cc.GadgetSettings.fromGS(&pb.GadgetSettings{}) //start with empty settings, but create nested structs


	cc.UpdateToDeployedGadgetSettings(vm)
	cc.DeployPending = false
	cc.RndisDetails = false
	cc.CdcEcmDetails = false

	return cc
}



const (
	compUSBSettingsTemplate = `
	<div>
	<table>
		<tr> <td>USB gadget settings</td><td><button @click="ApplyGadgetSettings" :disabled="deployPending">Apply</button></td> </tr>

		<tr>
			<td>Gadget enabled</td>
			<td>
			<toggle-switch v-model="gadgetSettings.Enabled"></toggle-switch>
			</td>
		</tr>

		
		<tr> <td>Vendor ID</td><td><input v-model="gadgetSettings.Vid"/></td> </tr>
		<tr> <td>Product ID</td><td><input v-model="gadgetSettings.Pid"/></td> </tr>
		<tr> <td>Manufacturer Name</td><td><input v-model="gadgetSettings.Manufacturer"/></td> </tr>
		<tr> <td>Product Name</td><td><input v-model="gadgetSettings.Product"/></td> </tr>
		<tr> <td>Serial number</td><td><input v-model="gadgetSettings.Serial"/></td> </tr>

		<tr>
			<td>CDC ECM</td>
			<td>
			<toggle-switch v-model="gadgetSettings.Use_CDC_ECM"></toggle-switch>
			</td>
		</tr>
		<tr v-if="gadgetSettings.Use_CDC_ECM">
		<td></td>
		<td>
			<ethernet-addresses v-bind:settings="gadgetSettings.CdcEcmSettings" @hostAddrChange="gadgetSettings.CdcEcmSettings.HostAddr=$event" @devAddrChange="gadgetSettings.CdcEcmSettings.DevAddr=$event"></ethernet-addresses>
		</td>
		</tr>

		<tr>
			<td>RNDIS</td>
			<td>

			<toggle-switch v-model="gadgetSettings.Use_RNDIS"></toggle-switch>
			<td></td>
			<input type="checkbox" v-if="gadgetSettings.Use_RNDIS" v-model="rndisDetails">

			</td>
		</tr>

		<tr v-if="rndisDetails">
		<td></td>
		<td>
			<ethernet-addresses v-bind:settings="gadgetSettings.RndisSettings" @hostAddrChange="gadgetSettings.RndisSettings.HostAddr=$event" @devAddrChange="gadgetSettings.RndisSettings.DevAddr=$event"></ethernet-addresses>
		</td>
		</tr>

		<tr>
			<td>HID Keyboard</td>
			<td>

			<toggle-switch v-model="gadgetSettings.Use_HID_KEYBOARD"></toggle-switch>

		
			</td>
		</tr>
		<tr>
			<td>HID Mouse</td>
			<td>
			<toggle-switch v-model="gadgetSettings.Use_HID_MOUSE"></toggle-switch>

			</td>
		</tr>
		<tr>
			<td>HID Raw</td>
			<td>
			<toggle-switch v-model="gadgetSettings.Use_HID_RAW"></toggle-switch>
			</td>
		</tr>
		<tr>
			<td>Serial</td>
			<td>
			<toggle-switch v-model="gadgetSettings.Use_SERIAL"></toggle-switch>
			</td>
		</tr>
		<tr>
			<td>Mass Storage</td>
			<td>
			<toggle-switch v-model="gadgetSettings.Use_UMS"></toggle-switch>
			</td>
		</tr>

	</table>
	</div>
`
)

