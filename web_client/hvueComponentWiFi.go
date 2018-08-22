package main

import (
	"github.com/HuckRidgeSW/hvue"
	"github.com/gopherjs/gopherjs/js"
	pb "github.com/mame82/P4wnP1_go/proto/gopherjs"
)

func InitComponentsWiFi() {
	hvue.NewComponent(
		"wifi",
		hvue.Template(templateWiFi),
		hvue.Computed("settings", func(vm *hvue.VM) interface{} {
			return vm.Get("$store").Get("state").Get("wifiSettings")
		}),
		hvue.ComputedWithGetSet("enabled",
			func(vm *hvue.VM) interface{} {
				return !vm.Get("settings").Get("disabled").Bool()
			},
			func(vm *hvue.VM, newValue *js.Object) {
				vm.Get("settings").Set("disabled", !newValue.Bool())
			},
		),
		hvue.ComputedWithGetSet("enableNexmon",
			func(vm *hvue.VM) interface{} {
				return !vm.Get("settings").Get("disableNexmon").Bool()
			},
			func(vm *hvue.VM, newValue *js.Object) {
				vm.Get("settings").Set("disableNexmon", !newValue.Bool())
			},
		),

		hvue.Computed("modes", func(vm *hvue.VM) interface{} {
			modes := js.Global.Get("Array").New()
			for val,name := range pb.WiFiSettings_Mode_name {
				mode := struct {
					*js.Object
					Val int `js:"val"`
					Name string `js:"name"`
				}{Object:O()}
				mode.Val = val
				mode.Name = name

				modes.Call("push", mode)
			}
			return modes
		}),
		hvue.Computed("authModes", func(vm *hvue.VM) interface{} {
			modes := js.Global.Get("Array").New()
			for val,name := range pb.WiFiSettings_APAuthMode_name {
				mode := struct {
					*js.Object
					Val int `js:"val"`
					Name string `js:"name"`
				}{Object:O()}
				mode.Val = val
				mode.Name = name

				modes.Call("push", mode)
			}
			return modes
		}),
	)
}

const templateWiFi = `
<div>
<table>
<tr>
	<td>Enabled</td>
	<td><toggle-switch v-model="enabled"></toggle-switch></td>
</tr>
<tr>
	<td>Enable Nexmon</td>
	<td><toggle-switch v-model="enableNexmon"></toggle-switch></td>
</tr>
<tr>
	<td>Regulatory Domain</td>
	<td><input v-model="settings.reg"></input></td>
</tr>
<tr>
	<td>Mode</td>
	<td>
		<select v-model="settings.mode">
			<option v-for="mode in modes" :value="mode.val">{{mode.name}}</option>
		</select>
	</td>
</tr>
<tr v-if="settings.mode != 1">
	<td>AP Channel</td>
	<td><input v-model.number="settings.channel"></input></td>
</tr>
<tr  v-if="settings.mode != 1">
	<td>AP Auth Algo</td>
	<td>
		<select v-model="settings.authMode">
			<option v-for="mode in authModes" :value="mode.val">{{mode.name}}</option>
		</select>
	</td>
</tr>
<tr  v-if="settings.mode != 1">
	<td>Hide SSID</td>
	<td><toggle-switch v-model="settings.hideSsid"></toggle-switch></td></tr>
</table>
{{ settings }}
</div>
`