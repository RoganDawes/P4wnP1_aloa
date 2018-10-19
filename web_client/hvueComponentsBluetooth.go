// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/mame82/hvue"
	pb "github.com/mame82/P4wnP1_go/proto/gopherjs"
)

type jsBluetoothControllerSettings struct {
	*js.Object
	Powered                 bool `js:"Powered"`
	Connectable             bool `js:"Connectable"`
	FastConnectable         bool `js:"FastConnectable"`
	Discoverable            bool `js:"Discoverable"`
	Bondable                bool `js:"Bondable"`
	LinkLevelSecurity       bool `js:"LinkLevelSecurity"`
	SecureSimplePairing     bool `js:"SecureSimplePairing"`
	BrEdr                   bool `js:"BrEdr"`
	HighSpeed               bool `js:"HighSpeed"`
	LowEnergy               bool `js:"LowEnergy"`
	Advertising             bool `js:"Advertising"`
	SecureConnections       bool `js:"SecureConnections"`
	DebugKeys               bool `js:"DebugKeys"`
	Privacy                 bool `js:"Privacy"`
	ControllerConfiguration bool `js:"ControllerConfiguration"`
	StaticAddress           bool `js:"StaticAddress"`
}

func (target *jsBluetoothControllerSettings) fromGo(src *pb.BluetoothControllerSettings) {
	target.Powered = src.Powered
	target.Connectable = src.Connectable
	target.FastConnectable = src.FastConnectable
	target.Discoverable = src.Discoverable
	target.Bondable = src.Bondable
	target.LinkLevelSecurity = src.LinkLevelSecurity
	target.SecureSimplePairing = src.SecureSimplePairing
	target.BrEdr = src.BrEdr
	target.HighSpeed = src.HighSpeed
	target.LowEnergy = src.LowEnergy
	target.Advertising = src.Advertising
	target.SecureConnections = src.SecureConnections
	target.DebugKeys = src.DebugKeys
	target.Privacy = src.Privacy
	target.ControllerConfiguration = src.ControllerConfiguration
	target.StaticAddress = src.StaticAddress
}

func (src *jsBluetoothControllerSettings) toGo() (target *pb.BluetoothControllerSettings) {
	target = &pb.BluetoothControllerSettings{}
	target.Powered = src.Powered
	target.Connectable = src.Connectable
	target.FastConnectable = src.FastConnectable
	target.Discoverable = src.Discoverable
	target.Bondable = src.Bondable
	target.LinkLevelSecurity = src.LinkLevelSecurity
	target.SecureSimplePairing = src.SecureSimplePairing
	target.BrEdr = src.BrEdr
	target.HighSpeed = src.HighSpeed
	target.LowEnergy = src.LowEnergy
	target.Advertising = src.Advertising
	target.SecureConnections = src.SecureConnections
	target.DebugKeys = src.DebugKeys
	target.Privacy = src.Privacy
	target.ControllerConfiguration = src.ControllerConfiguration
	target.StaticAddress = src.StaticAddress
	return
}

func NewBluetoothControllerSettings() (res *jsBluetoothControllerSettings) {
	res = &jsBluetoothControllerSettings{Object: O()}
	res.Powered = false
	res.Connectable = false
	res.FastConnectable = false
	res.Discoverable = false
	res.Bondable = false
	res.LinkLevelSecurity = false
	res.SecureSimplePairing = false
	res.BrEdr = false
	res.HighSpeed = false
	res.LowEnergy = false
	res.Advertising = false
	res.SecureConnections = false
	res.DebugKeys = false
	res.Privacy = false
	res.ControllerConfiguration = false
	res.StaticAddress = false
	return
}

type jsBluetoothControllerInformation struct {
	*js.Object
	IsAvailable       bool                           `js:"IsAvailable"`
	Address           []byte                         `js:"Address"`
	BluetoothVersion  byte                           `js:"BluetoothVersion"`
	Manufacturer      uint16                         `js:"Manufacturer"`
	SupportedSettings *jsBluetoothControllerSettings `js:"SupportedSettings"`
	CurrentSettings   *jsBluetoothControllerSettings `js:"CurrentSettings"`
	ClassOfDevice     []byte                         `js:"ClassOfDevice"` // 3, till clear how to parse
	Name              string                         `js:"Name"`          //[249]byte, 0x00 terminated
	ShortName         string                         `js:"ShortName"`     //[11]byte, 0x00 terminated
}

func (src *jsBluetoothControllerInformation) toGo() (target *pb.BluetoothControllerInformation) {
	target = &pb.BluetoothControllerInformation{}
	target.IsAvailable = src.IsAvailable
	target.Address = src.Address
	target.BluetoothVersion = uint32(src.BluetoothVersion)
	target.Manufacturer = uint32(src.Manufacturer)
	target.SupportedSettings = src.SupportedSettings.toGo()
	target.CurrentSettings = src.CurrentSettings.toGo()
	target.Name = src.Name
	target.ShortName = src.ShortName
	return
}

func (target *jsBluetoothControllerInformation) fromGo(src *pb.BluetoothControllerInformation) {
	target.Address = src.Address
	target.IsAvailable = src.IsAvailable
	target.ClassOfDevice = src.ClassOfDevice
	target.BluetoothVersion = byte(src.BluetoothVersion)
	target.Manufacturer = uint16(src.Manufacturer)
	target.SupportedSettings = NewBluetoothControllerSettings()
	target.SupportedSettings.fromGo(src.SupportedSettings)
	target.CurrentSettings = NewBluetoothControllerSettings()
	target.CurrentSettings.fromGo(src.CurrentSettings)
	target.Name = src.Name
	target.ShortName = src.ShortName
}

func NewBluetoothControllerInformation() (res *jsBluetoothControllerInformation) {
	res = &jsBluetoothControllerInformation{Object: O()}
	res.IsAvailable = false
	res.ShortName = ""
	res.Name = ""
	res.Manufacturer = 0
	res.BluetoothVersion = 0
	res.Address = make([]byte, 6)
	res.ClassOfDevice = make([]byte, 3)
	res.SupportedSettings = NewBluetoothControllerSettings()
	res.CurrentSettings = NewBluetoothControllerSettings()
	return
}

func InitComponentsBluetooth() {
	hvue.NewComponent(
		"bluetooth",
		hvue.Template(templateBluetoothPage),
		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := &struct {
				*js.Object
		//		ControllerInfo *jsBluetoothControllerInformation `js:"ControllerInfo"`
			}{Object: O()}

		//	data.ControllerInfo = NewBluetoothControllerInformation()

			return data
		}),
		hvue.Computed("CurrentControllerInfo",
			func(vm *hvue.VM) interface{} {
				return vm.Get("$store").Get("state").Get("CurrentBluetoothControllerInformation")
			}),
		hvue.Computed("available", func(vm *hvue.VM) interface{} {
			return true
		}),
		hvue.Mounted(func(vm *hvue.VM) {
			vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_CURRENT_BLUETOOTH_CONTROLLER_INFORMATION)
		}),
	)
	hvue.NewComponent(
		"bluetooth-controller",
		hvue.Template(templateBluetoothController),
		hvue.PropObj("controllerInfo"),
		hvue.Computed("CurrentControllerInfo",
			func(vm *hvue.VM) interface{} {
				return vm.Get("$store").Get("state").Get("CurrentBluetoothControllerInformation")
			}),
	)
}

const templateBluetoothPage = `
<q-page>
<bluetooth-controller :controllerInfo="CurrentControllerInfo"></bluetooth-controller>
</q-page>
`
const templateBluetoothController = `
<q-card>
{{ controllerInfo }}

	<q-list link>
		<q-item tag="label">
			<q-item-side>
				<q-toggle v-model="controllerInfo.CurrentSettings.Powered" @input="$store.dispatch('deployCurrentBluetoothControllerInformation')"></q-toggle>
			</q-item-side>
			<q-item-main>
				<q-item-tile label>Enabled</q-item-tile>
				<q-item-tile sublabel>Power on/off Bluetooth controller</q-item-tile>
			</q-item-main>
		</q-item>

		<q-item tag="label">
			<q-item-main>
				<q-item-tile label>Name</q-item-tile>
				<q-item-tile sublabel>Visible name of the bluetooth device</q-item-tile>
				<q-item-tile>
					<q-input :value="controllerInfo.Name" @change="controllerInfo.Name = $event; $store.dispatch('deployCurrentBluetoothControllerInformation')" inverted></q-input>
				</q-item-tile>
			</q-item-main>
		</q-item>
	</q-list>

</q-card>
`
