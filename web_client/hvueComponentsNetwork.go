package main

import (
	//"github.com/HuckRidgeSW/hvue"
	"github.com/gopherjs/gopherjs/js"
	pb "github.com/mame82/P4wnP1_go/proto/gopherjs"

	"github.com/HuckRidgeSW/hvue"
)



func InitComponentsNetwork() {

	hvue.NewComponent(
		"network",
		hvue.Template(templateNetwork),
		hvue.Computed("interfaces", func(vm *hvue.VM) interface{} {
			return vm.Get("$store").Get("state").Get("InterfaceSettings").Get("interfaces")
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
		hvue.Computed("withDhcp", func(vm *hvue.VM) interface{} {
			if mode := vm.Get("interface").Get("mode").Int();  mode == pb.EthernetInterfaceSettings_Mode_value["DHCP_SERVER"] {
				return true
			} else {
				return false
			}
		}),
		hvue.Method("deploy", func(vm *hvue.VM) {
			vm.Get("$store").Call("dispatch", VUEX_ACTION_DEPLOY_ETHERNET_INTERFACE_SETTINGS, vm.Get("interface"))
		}),
	)

	hvue.NewComponent("dhcpconfig",
		hvue.Props("interface"),
		hvue.Template(templateDHCPConfig),
		hvue.Computed("config", func(vm *hvue.VM) interface{} {
			if vm.Get("interface").Get("dhcpServerSettings") == js.Undefined {
				// no DHCP server settings present

				//cast interface to struct
				iface := &jsEthernetInterfaceSettings{Object:vm.Get("interface")}
				iface.CreateDhcpSettingsForInterface() //Create proper DHCP server settings for interface

			}
			return vm.Get("interface").Get("dhcpServerSettings")

		}),
		hvue.ComputedWithGetSet("authoritative",
			func(vm *hvue.VM) interface{} {
				return !vm.Get("config").Get("nonAuthoritative").Bool()
			},
			func(vm *hvue.VM, newValue *js.Object) {
				vm.Get("config").Set("nonAuthoritative", !newValue.Bool())
			}),
	)

	hvue.NewComponent("dhcpranges",
		hvue.Props("serversettings"),
		hvue.Template(templateDHCPRanges),
		hvue.Method("addRange", func(vm *hvue.VM) {
			s := &jsDHCPServerSettings{Object:vm.Get("serversettings")}
			r := &jsDHCPServerRange{Object:O()}
			r.RangeLower = ""
			r.RangeUpper = ""
			r.LeaseTime = "1m"
			s.AddRange(r)
		}),
		hvue.Method("removeRange", func(vm *hvue.VM, delRange *jsDHCPServerRange) {
			s := &jsDHCPServerSettings{Object:vm.Get("serversettings")}
			s.RemoveRange(delRange)
		}),
	)
	hvue.NewComponent("dhcpoptions",
		hvue.Props("serversettings"),
		hvue.Template(templateDHCPOptions),
		hvue.Method("addOption", func(vm *hvue.VM) {
			s := &jsDHCPServerSettings{Object:vm.Get("serversettings")}
			o := &jsDHCPServerOption{Object:O()}
			o.Option = 3
			o.Value = ""
			s.AddOption(o)
		}),
		hvue.Method("removeOption", func(vm *hvue.VM, delOption *jsDHCPServerOption) {
			s := &jsDHCPServerSettings{Object:vm.Get("serversettings")}
			s.RemoveOption(delOption)
		}),
	)
	hvue.NewComponent("dhcpstatichosts",
		hvue.Props("serversettings"),
		hvue.Template(templateDHCPStaticHosts),
		hvue.Method("addStaticHost", func(vm *hvue.VM) {
			s := &jsDHCPServerSettings{Object:vm.Get("serversettings")}
			sh := &jsDHCPServerStaticHost{Object:O()}
			sh.Ip = ""
			sh.Mac = ""
			s.AddStaticHost(sh)
		}),
		hvue.Method("removeStaticHost", func(vm *hvue.VM, delStaticHost *jsDHCPServerStaticHost) {
			s := &jsDHCPServerSettings{Object:vm.Get("serversettings")}
			s.RemoveStaticHost(delStaticHost)
		}),
	)
}

const templateNetwork = `
<div class="network-master">
Interface selection <select v-model="current">
	<option v-for="iface in interfaces" :key="iface.name" :value="iface">{{ iface.name }}</option>
</select>
<networkinterface v-if="current" :interface="current"></networkinterface>

</div>
`

const templateNetworkInterface = `
<div class="network-interface">
<h1>
	Interface settings for "{{interface.name}}"
</h1>
Deploy settings <button @click="deploy">DEPLOY</button>

<!-- <p>{{ interface }}</p> -->
<table>
<tr>
	<td>Enabled</td>
	<td>
		<toggle-switch v-model="interface.enabled"></toggle-switch>
	</td>
</tr>
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
<dhcpconfig v-if="withDhcp" :interface="interface"></dhcpconfig>
</div>
`

const templateDHCPConfig = `
<div>
<p><b>DHCP server settings</b></p>
<table>
<tr>
	<td>Listen port</td>
	<td><input v-model="config.listenPort"></input></td>
	<td>Port for DNS server (0 to disable DNS and use DHCP only)</td>
</tr>
<tr>
	<td>Lease file</td>
	<td>{{ config.leaseFile }}</td>
	<td>Path to lease file</td>
</tr>
<tr>
	<td>Authoritative</td>
	<td><toggle-switch type="checkbox" v-model="authoritative"></toggle-switch></td>
	<td></td>
</tr>
<tr>
	<td>Bind only to {{ config.listenInterface }}</td>
	<td>{{ !config.doNotBindInterface }}</td>
	<td></td>
</tr>
</table>
<dhcpranges :serversettings="config"></dhcpranges>
<dhcpoptions :serversettings="config"></dhcpoptions>
<dhcpstatichosts :serversettings="config"></dhcpstatichosts>
<!-- {{ config }} -->
</div>
`

const templateDHCPRanges = `
<div>
<p><b>DHCP ranges</b></p>
<button @click="addRange">ADD</button>
<table>
	<tr v-for="range in serversettings.ranges">
		<td>First IP</td> <td><input v-model="range.rangeLower"></input></td>
		<td>Last IP</td> <td><input v-model="range.rangeUpper"></input></td>
		<td>Lease time</td> <td><input v-model="range.leaseTime"></input></td>
		<td><button @click="removeRange(range)">DEL</button></td>
	</tr>
</table>
</div>
`
const templateDHCPOptions = `
<div>
<p><b>Options</b></p>
<button @click="addOption">ADD</button>
<table>
	<tr v-for="option in serversettings.options">
		<td>Option number</td> <td><input v-model.number="option.option"></input></td>
		<td>Option string</td> <td><input v-model="option.value"></input></td>
		<td><button @click="removeOption(option)">DEL</button></td>
	</tr>
</table>
</div>
`
const templateDHCPStaticHosts = `
<div>
<p><b>Static host entries</b></p>
<button @click="addStaticHost">ADD</button>
<table>
	<tr v-for="statichost in serversettings.staticHosts">
		<td>Host Mac</td> <td><input v-model="statichost.mac"></input></td>
		<td>Host IP</td> <td><input v-model="statichost.ip"></input></td>
		<td><button @click="removeStaticHost(statichost)">DEL</button></td>
	</tr>
</table>
</div>
`