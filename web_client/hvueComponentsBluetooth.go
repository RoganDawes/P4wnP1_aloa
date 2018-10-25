// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/mame82/hvue"
	pb "github.com/mame82/P4wnP1_go/proto/gopherjs"
)

type jsBluetoothRequestSettingsStorage struct {
	*js.Object
	TemplateName string `js:"TemplateName"`
	Settings     *jsBluetoothSettings `js:"Settings"`
}

func (rs *jsBluetoothRequestSettingsStorage) toGo() *pb.BluetoothRequestSettingsStorage {
	return &pb.BluetoothRequestSettingsStorage{
		TemplateName: rs.TemplateName,
		Settings: rs.Settings.toGo(),
	}
}

func (rs *jsBluetoothRequestSettingsStorage) fromGo(src *pb.BluetoothRequestSettingsStorage) {
	rs.TemplateName = src.TemplateName
	rs.Settings = NewBluetoothSettings()
	rs.Settings.fromGo(src.Settings)
}

func NewBluetoothRequestSettingsStorage() (res *jsBluetoothRequestSettingsStorage) {
	res = &jsBluetoothRequestSettingsStorage{Object:O()}
	res.TemplateName = ""
	res.Settings = NewBluetoothSettings()
	return res
}

func NewBluetoothRequestSettingsStorageFromArgs(as *jsBluetoothAgentSettings, ci *jsBluetoothControllerInformation, templateName string) (res *jsBluetoothRequestSettingsStorage) {
	res = &jsBluetoothRequestSettingsStorage{Object:O()}
	res.TemplateName = templateName
	res.Settings = NewBluetoothSettings()
	res.Settings.fromASandCI(as,ci)
	return res
}

type jsBluetoothSettings struct {
	*js.Object
	Ci *jsBluetoothControllerInformation `js:"Ci"`
	As *jsBluetoothAgentSettings `js:"As"`
}

func (target *jsBluetoothSettings) fromGo(src *pb.BluetoothSettings) {
	target.As = NewBluetoothAgentSettings()
	target.As.fromGo(src.As)
	target.Ci = NewBluetoothControllerInformation()
	target.Ci.fromGo(src.Ci)
}

func (target *jsBluetoothSettings) fromASandCI(as *jsBluetoothAgentSettings, ci *jsBluetoothControllerInformation) {
	target.As = as
	target.Ci = ci
}

func (src *jsBluetoothSettings) toGo() (target *pb.BluetoothSettings) {
	target = &pb.BluetoothSettings{
		Ci: src.Ci.toGo(),
		As: src.As.toGo(),
	}
	return target
}

func NewBluetoothSettings() (res *jsBluetoothSettings) {
	res = &jsBluetoothSettings{Object:O()}
	res.As = NewBluetoothAgentSettings()
	res.Ci = NewBluetoothControllerInformation()
	return
}

type jsBluetoothAgentSettings struct {
	*js.Object
	Pin string `js:"Pin"`
}

func (target *jsBluetoothAgentSettings) fromGo(src *pb.BluetoothAgentSettings) {
	target.Pin = src.Pin
}

func (src *jsBluetoothAgentSettings) toGo() (target *pb.BluetoothAgentSettings) {
	target = &pb.BluetoothAgentSettings{}
	target.Pin = src.Pin
	return
}

func NewBluetoothAgentSettings() (res *jsBluetoothAgentSettings) {
	res = &jsBluetoothAgentSettings{Object:O()}
	res.Pin = ""
	return
}

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

	ServiceNetworkServerNAP bool `js:"ServiceNetworkServerNAP"`
	ServiceNetworkServerPANU bool `js:"ServiceNetworkServerPANU"`
	ServiceNetworkServerGN bool `js:"ServiceNetworkServerGN"`
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

	target.ServiceNetworkServerGn = src.ServiceNetworkServerGN
	target.ServiceNetworkServerPanu = src.ServiceNetworkServerPANU
	target.ServiceNetworkServerNap = src.ServiceNetworkServerNAP
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
	target.ServiceNetworkServerGN = src.ServiceNetworkServerGn
	target.ServiceNetworkServerPANU = src.ServiceNetworkServerPanu
	target.ServiceNetworkServerNAP = src.ServiceNetworkServerNap

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

	res.ServiceNetworkServerGN = false
	res.ServiceNetworkServerPANU = false
	res.ServiceNetworkServerNAP = false

	return
}

func InitComponentsBluetooth() {
	hvue.NewComponent(
		"bluetooth",
		hvue.Template(templateBluetoothPage),
		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := struct {
				*js.Object
				ShowStoreModal bool   `js:"showStoreModal"`
				ShowDeployStoredModal bool   `js:"showDeployStoredModal"`
//				TemplateName   string `js:"templateName"`
			}{Object: O()}
			data.ShowStoreModal = false
			data.ShowDeployStoredModal = false
//			data.TemplateName = ""
			return &data
		}),
		hvue.Computed("CurrentControllerInfo",
			func(vm *hvue.VM) interface{} {
				return vm.Get("$store").Get("state").Get("CurrentBluetoothControllerInformation")
			}),
		hvue.Computed("available", func(vm *hvue.VM) interface{} {
			return true
		}),
		hvue.Computed("CurrentBluetoothAgentSettings",
			func(vm *hvue.VM) interface{} {
				return vm.Get("$store").Get("state").Get("CurrentBluetoothAgentSettings")
			}),
		hvue.Method("store",
			func(vm *hvue.VM, name *js.Object) {
				ci := &jsBluetoothControllerInformation{
					Object: vm.Get("$store").Get("state").Get("CurrentBluetoothControllerInformation"),
				}
				as := &jsBluetoothAgentSettings{
					Object: vm.Get("$store").Get("state").Get("CurrentBluetoothAgentSettings"),
				}
				sReq := NewBluetoothRequestSettingsStorageFromArgs(as,ci,name.String())
				println("Storing :", sReq)
				vm.Get("$store").Call("dispatch", VUEX_ACTION_STORE_BLUETOOTH_SETTINGS, sReq)
				vm.Set("showStoreModal", false)
			}),
		hvue.Method("deleteStored",
			func(vm *hvue.VM, name *js.Object) {
				println("Loading :", name.String())
				vm.Get("$store").Call("dispatch", VUEX_ACTION_DELETE_STORED_BLUETOOTH_SETTINGS, name)
			}),
		hvue.Method("deployStored",
			func(vm *hvue.VM, name *js.Object) {
				println("Loading :", name.String())
				vm.Get("$store").Call("dispatch", VUEX_ACTION_DEPLOY_STORED_BLUETOOTH_SETTINGS, name)
			}),
		hvue.Method("updateStoredSettingsList",
			func(vm *hvue.VM) {
				vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_STORED_BLUETOOTH_SETTINGS_LIST)
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
	hvue.NewComponent(
		"bluetooth-controller-network-services",
		hvue.Template(templateBluetoothControllerNetworkServices),
		hvue.PropObj("controllerInfo"),
		hvue.Computed("CurrentControllerInfo",
			func(vm *hvue.VM) interface{} {
				return vm.Get("$store").Get("state").Get("CurrentBluetoothControllerInformation")
			}),
	)
	hvue.NewComponent(
		"bluetooth-agent",
		hvue.Template(templateBluetoothAgent),
		hvue.PropObj("bluetoothAgent"),
		hvue.Mounted(func(vm *hvue.VM) {
			vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_CURRENT_BLUETOOTH_AGENT_SETTINGS)
		}),
	)
}

const templateBluetoothPage = `
<q-page padding>

	<select-string-from-array :values="$store.state.StoredBluetoothSettingsList" v-model="showDeployStoredModal" title="Deploy stored bluetooth settings" @load="deployStored($event)" @delete="deleteStored($event)" with-delete></select-string-from-array>
	<modal-string-input v-model="showStoreModal" title="Store bluetooth settings" @save="store($event)"></modal-string-input>


	<div class="row gutter-sm">

		<div class="col-12">
			<q-card>
				<q-card-title>
					Bluetooth Settings
				</q-card-title>

				<q-card-main>
					<div class="row gutter-sm">
						<div class="col-6 col-sm""><q-btn class="fit" color="primary" @click="updateStoredSettingsList(); showDeployStoredModal=true" label="deploy stored" icon="settings_backup_restore"></q-btn></div>
						<div class="col-6 col-sm""><q-btn class="fit" color="secondary" @click="showStoreModal=true" label="store" icon="cloud_upload"></q-btn></div>
					</div>
  				</q-card-main>
			</q-card>
		</div>


		<div class="col-12">
			{{ CurrentControllerInfo }}
		</div>

		<div class="col-12 col-lg">
			<bluetooth-controller :controllerInfo="CurrentControllerInfo"></bluetooth-controller>
		</div>
		<div class="col-12 col-lg">
			<bluetooth-controller-network-services :controllerInfo="CurrentControllerInfo"></bluetooth-controller-network-services>
		</div>
		<div class="col-12 col-lg">
			<bluetooth-agent :bluetoothAgent="CurrentBluetoothAgentSettings"></bluetooth-agent>
		</div>
	</div>
</q-page>
`
const templateBluetoothController = `
<q-card>
	<q-card-title>
		Generic Bluetooth Controller settings
	</q-card-title>

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

		<q-item tag="label">
			<q-item-side>
				<q-toggle v-model="controllerInfo.CurrentSettings.Connectable" @input="$store.dispatch('deployCurrentBluetoothControllerInformation')"></q-toggle>
			</q-item-side>
			<q-item-main>
				<q-item-tile label>Connectable</q-item-tile>
				<q-item-tile sublabel>Allow incoming connections</q-item-tile>
			</q-item-main>
		</q-item>

		<q-item tag="label">
			<q-item-side>
				<q-toggle v-model="controllerInfo.CurrentSettings.Discoverable" @input="$store.dispatch('deployCurrentBluetoothControllerInformation')"></q-toggle>
			</q-item-side>
			<q-item-main>
				<q-item-tile label>Discoverable</q-item-tile>
				<q-item-tile sublabel>P4wnP1 could be discovered by other devices if enabled (only if Connectable)</q-item-tile>
			</q-item-main>
		</q-item>

		<q-item tag="label">
			<q-item-side>
				<q-toggle v-model="controllerInfo.CurrentSettings.Bondable" @input="$store.dispatch('deployCurrentBluetoothControllerInformation')"></q-toggle>
			</q-item-side>
			<q-item-main>
				<q-item-tile label>Bondable</q-item-tile>
				<q-item-tile sublabel>Other devices could pair with P4wnP1</q-item-tile>
			</q-item-main>
		</q-item>

		<q-item tag="label">
			<q-item-side>
				<q-toggle v-model="controllerInfo.CurrentSettings.HighSpeed" @input="$store.dispatch('deployCurrentBluetoothControllerInformation')"></q-toggle>
			</q-item-side>
			<q-item-main>
				<q-item-tile label>High Speed</q-item-tile>
				<q-item-tile sublabel>Use alternate data channel (802.11)</q-item-tile>
			</q-item-main>
		</q-item>

		<q-item tag="label">
			<q-item-side>
				<q-toggle v-model="controllerInfo.CurrentSettings.LowEnergy" @input="$store.dispatch('deployCurrentBluetoothControllerInformation')"></q-toggle>
			</q-item-side>
			<q-item-main>
				<q-item-tile label>Low Energy</q-item-tile>
				<q-item-tile sublabel>Enable Bluetooth LE (Bluetooth Smart)</q-item-tile>
			</q-item-main>
		</q-item>

		<q-item tag="label">
			<q-item-side>
				<q-toggle v-model="controllerInfo.CurrentSettings.SecureSimplePairing" @input="$store.dispatch('deployCurrentBluetoothControllerInformation')"></q-toggle>
			</q-item-side>
			<q-item-main>
				<q-item-tile label>Secure Simple Pairing</q-item-tile>
				<q-item-tile sublabel>If disabled, insecure PIN based pairing is used and HighSpeed isn't available</q-item-tile>
			</q-item-main>
		</q-item>



	</q-list>

</q-card>
`

const templateBluetoothControllerNetworkServices = `
<q-card>
	<q-card-title>
		BNEP server services
		<span slot="subtitle">Bluetooth Network Encapsulation Protocol services provided by the controller</span>
	</q-card-title>


	<q-list link>
		<q-item tag="label">
			<q-item-side>
				<q-toggle v-model="controllerInfo.ServiceNetworkServerNAP" @input="$store.dispatch('deployCurrentBluetoothControllerInformation')"></q-toggle>
			</q-item-side>
			<q-item-main>
				<q-item-tile label>NAP</q-item-tile>
				<q-item-tile sublabel>Provide Network Access Point</q-item-tile>
			</q-item-main>
		</q-item>

		<q-item tag="label">
			<q-item-side>
				<q-toggle v-model="controllerInfo.ServiceNetworkServerPANU" @input="$store.dispatch('deployCurrentBluetoothControllerInformation')"></q-toggle>
			</q-item-side>
			<q-item-main>
				<q-item-tile label>PANU</q-item-tile>
				<q-item-tile sublabel>Provide Protable Area Network Unit</q-item-tile>
			</q-item-main>
		</q-item>

		<q-item tag="label">
			<q-item-side>
				<q-toggle v-model="controllerInfo.ServiceNetworkServerGN" @input="$store.dispatch('deployCurrentBluetoothControllerInformation')"></q-toggle>
			</q-item-side>
			<q-item-main>
				<q-item-tile label>GN</q-item-tile>
				<q-item-tile sublabel>Provide Group Ad-hoc Network</q-item-tile>
			</q-item-main>
		</q-item>

	</q-list>


</q-card>
`

const templateBluetoothAgent = `
<q-card>
	<q-card-title>
		Authentication Agent
	</q-card-title>


	<q-list link>
		<q-item tag="label">
			<q-item-main>
				<q-item-tile label>Pin</q-item-tile>
				<q-item-tile sublabel>PIN requested from remote devices on bonding (only if SSP is off)</q-item-tile>
				<q-item-tile>
					<q-input :value="bluetoothAgent.Pin" @change="bluetoothAgent.Pin = $event; $store.dispatch('deployCurrentBluetoothAgentSettings')" type="password" inverted></q-input>
				</q-item-tile>
			</q-item-main>
		</q-item>

	</q-list>


</q-card>
`
