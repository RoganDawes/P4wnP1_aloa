package main

import (
	"github.com/mame82/hvue"
	"github.com/gopherjs/gopherjs/js"
	pb "github.com/mame82/P4wnP1_go/proto/gopherjs"

)



func InitComponentsNetwork() {
	hvue.NewComponent(
		"networkinterface",
		hvue.Template(templateNetworkInterface),
		hvue.Props("interface"),
		hvue.Computed("modes", func(vm *hvue.VM) interface{} {
			modes := js.Global.Get("Array").New()
			for val,name := range pb.EthernetInterfaceSettings_Mode_name {
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
		hvue.Computed("withIP", func(vm *hvue.VM) interface{} {
			if mode := vm.Get("interface").Get("mode").Int(); mode == pb.EthernetInterfaceSettings_Mode_value["MANUAL"] ||  mode == pb.EthernetInterfaceSettings_Mode_value["DHCP_SERVER"] {
				return true
			} else {
				return false
			}
		}),
	)

	hvue.NewComponent(
		"network",
		hvue.Template(templateNetwork),
		hvue.Computed("interfaces", func(vm *hvue.VM) interface{} {
			return vm.Store.Get("state").Get("InterfaceSettings").Get("interfaces")
		}),

		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := &struct {
				*js.Object
				CurrentInterface *jsEthernetInterfaceSettings `js:"current"`
			}{Object:O()}
			data.CurrentInterface = &jsEthernetInterfaceSettings{Object:js.Undefined}
			return data
		}),
		hvue.Created(func(vm *hvue.VM) {
			// data field "current" is still undefined, set to first interface of computed property "interfaces" (if there is one)
			if vm.Get("interfaces").Length() > 0 {
				hvue.Set(vm, "current", vm.Get("interfaces").Index(0))
			}
		}),
	)
}

const templateNetwork = `
<div class="network-master">
<select v-model="current">
	<option v-for="iface in interfaces" :key="iface.name" :value="iface">{{ iface.name }}</option>
</select>
<networkinterface v-if="current" :interface="current"></networkinterface>

</div>
`

const templateNetworkInterface = `
<div class="network-interface">
<h1>{{interface.name}}</h1>
<p>{{ interface }}</p>
<table>
<tr>
	<td>Mode</td>
	<td>
		<select v-model="interface.mode">
			<option v-for="mode in modes" :value="mode.val">{{mode.name}}</option>
		</select>
	</td>
</tr>
<tr v-if="withIP">
	<td>IP</td>
	<td><input v-model="interface.ipAddress4"></input></td>
</tr>
<tr v-if="withIP">
	<td>Mask</td>
	<td><input v-model="interface.netmask4"></input></td>
</tr>
</table>
</div>
`

