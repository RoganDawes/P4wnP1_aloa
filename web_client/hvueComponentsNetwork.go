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

		// converts interface array to array which could be used with Quasar q-select (every object item has label and value)
		hvue.Computed("selectOptionsInterface", func(vm *hvue.VM) interface{} {
			selectIf := js.Global.Get("Array").New()
			interfaces := vm.Get("$store").Get("state").Get("InterfaceSettings").Get("interfaces")
			for i := 0; i < interfaces.Length(); i++ {
				option := struct {
					*js.Object
					Label string                       `js:"label"`
					Value *jsEthernetInterfaceSettings `js:"value"`
				}{Object: O()}
				currentIf := &jsEthernetInterfaceSettings{
					Object: interfaces.Index(i),
				}
				option.Label = currentIf.Name
				option.Value = currentIf
				selectIf.Call("push", option)
			}
			return selectIf
		}),
		hvue.Computed("currentWithDhcp", func(vm *hvue.VM) interface{} {
			if mode := vm.Get("current").Get("mode").Int(); mode == pb.EthernetInterfaceSettings_Mode_value["DHCP_SERVER"] {
				return true
			} else {
				return false
			}
		}),

		hvue.Method("deploy", func(vm *hvue.VM, ifaceSettings *jsEthernetInterfaceSettings) {
			vm.Get("$store").Call("dispatch", VUEX_ACTION_DEPLOY_ETHERNET_INTERFACE_SETTINGS, ifaceSettings)
		}),

		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := &struct {
				*js.Object
				CurrentInterface *jsEthernetInterfaceSettings `js:"current"`
			}{Object: O()}
			data.CurrentInterface = &jsEthernetInterfaceSettings{Object: js.Undefined}
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
		"networkinterface2",
		hvue.Template(templateNetworkInterface2),
		hvue.Props("interface"),
		hvue.Computed("selectOptionsInterfaceModes", func(vm *hvue.VM) interface{} {
			modes := js.Global.Get("Array").New()
			for val, name := range pb.EthernetInterfaceSettings_Mode_name {
				mode := struct {
					*js.Object
					Val  int    `js:"value"`
					Name string `js:"label"`
				}{Object: O()}
				mode.Val = val
				mode.Name = name

				modes.Call("push", mode)
			}
			return modes
		}),
		hvue.Computed("withIP", func(vm *hvue.VM) interface{} {
			if mode := vm.Get("interface").Get("mode").Int(); mode == pb.EthernetInterfaceSettings_Mode_value["MANUAL"] || mode == pb.EthernetInterfaceSettings_Mode_value["DHCP_SERVER"] {
				return true
			} else {
				return false
			}
		}),
		hvue.Computed("withDhcp", func(vm *hvue.VM) interface{} {
			if mode := vm.Get("interface").Get("mode").Int(); mode == pb.EthernetInterfaceSettings_Mode_value["DHCP_SERVER"] {
				return true
			} else {
				return false
			}
		}),
		hvue.Method("deploy", func(vm *hvue.VM) {
			vm.Get("$store").Call("dispatch", VUEX_ACTION_DEPLOY_ETHERNET_INTERFACE_SETTINGS, vm.Get("interface"))
		}),
	)

	hvue.NewComponent("dhcpconfig2",
		hvue.Props("interface"),
		hvue.Template(templateDHCPConfig2),
		hvue.Computed("config", func(vm *hvue.VM) interface{} {
			if vm.Get("interface").Get("dhcpServerSettings") == js.Undefined {
				// no DHCP server settings present

				//cast interface to struct
				iface := &jsEthernetInterfaceSettings{Object: vm.Get("interface")}
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
		hvue.Method("addRange", func(vm *hvue.VM) {
			s := &jsDHCPServerSettings{Object: vm.Get("config")}
			r := &jsDHCPServerRange{Object: O()}
			r.RangeLower = ""
			r.RangeUpper = ""
			r.LeaseTime = "1m"
			s.AddRange(r)
		}),
		hvue.Method("removeRange", func(vm *hvue.VM, delRange *jsDHCPServerRange) {
			s := &jsDHCPServerSettings{Object: vm.Get("config")}
			s.RemoveRange(delRange)
		}),

	)


	hvue.NewComponent(
		"networkinterface",
		hvue.Template(templateNetworkInterface),
		hvue.Props("interface"),
		hvue.Computed("modes", func(vm *hvue.VM) interface{} {
			modes := js.Global.Get("Array").New()
			for val, name := range pb.EthernetInterfaceSettings_Mode_name {
				mode := struct {
					*js.Object
					Val  int    `js:"val"`
					Name string `js:"name"`
				}{Object: O()}
				mode.Val = val
				mode.Name = name

				modes.Call("push", mode)
			}
			return modes
		}),
		hvue.Computed("withIP", func(vm *hvue.VM) interface{} {
			if mode := vm.Get("interface").Get("mode").Int(); mode == pb.EthernetInterfaceSettings_Mode_value["MANUAL"] || mode == pb.EthernetInterfaceSettings_Mode_value["DHCP_SERVER"] {
				return true
			} else {
				return false
			}
		}),
		hvue.Computed("withDhcp", func(vm *hvue.VM) interface{} {
			if mode := vm.Get("interface").Get("mode").Int(); mode == pb.EthernetInterfaceSettings_Mode_value["DHCP_SERVER"] {
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
				iface := &jsEthernetInterfaceSettings{Object: vm.Get("interface")}
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
			s := &jsDHCPServerSettings{Object: vm.Get("serversettings")}
			r := &jsDHCPServerRange{Object: O()}
			r.RangeLower = ""
			r.RangeUpper = ""
			r.LeaseTime = "1m"
			s.AddRange(r)
		}),
		hvue.Method("removeRange", func(vm *hvue.VM, delRange *jsDHCPServerRange) {
			s := &jsDHCPServerSettings{Object: vm.Get("serversettings")}
			s.RemoveRange(delRange)
		}),
	)
	hvue.NewComponent("dhcpoptions",
		hvue.Props("serversettings"),
		hvue.Template(templateDHCPOptions),
		hvue.Method("addOption", func(vm *hvue.VM) {
			s := &jsDHCPServerSettings{Object: vm.Get("serversettings")}
			o := &jsDHCPServerOption{Object: O()}
			o.Option = 3
			o.Value = ""
			s.AddOption(o)
		}),
		hvue.Method("removeOption", func(vm *hvue.VM, delOption *jsDHCPServerOption) {
			s := &jsDHCPServerSettings{Object: vm.Get("serversettings")}
			s.RemoveOption(delOption)
		}),
	)
	hvue.NewComponent("dhcpstatichosts",
		hvue.Props("serversettings"),
		hvue.Template(templateDHCPStaticHosts),
		hvue.Method("addStaticHost", func(vm *hvue.VM) {
			s := &jsDHCPServerSettings{Object: vm.Get("serversettings")}
			sh := &jsDHCPServerStaticHost{Object: O()}
			sh.Ip = ""
			sh.Mac = ""
			s.AddStaticHost(sh)
		}),
		hvue.Method("removeStaticHost", func(vm *hvue.VM, delStaticHost *jsDHCPServerStaticHost) {
			s := &jsDHCPServerSettings{Object: vm.Get("serversettings")}
			s.RemoveStaticHost(delStaticHost)
		}),
	)
}

const templateNetwork = `
<q-page>
<q-card inline class="q-ma-sm">
	<q-card-title>
    	Network interface settings
	</q-card-title>

	<q-card-actions>
		<q-btn color="primary" @click="deploy(current)" label="deploy"></q-btn>
	</q-card-actions>

	<q-list link>
		<q-item-separator />
		<q-item tag="label">
			<q-item-main>
				<q-item-tile label>Interface</q-item-tile>
				<q-item-tile sublabel>Select which interface to configure</q-item-tile>
				<q-item-tile>
					<q-select v-model="current" :options="selectOptionsInterface" color="secondary" inverted></q-select>
				</q-item-tile>
			</q-item-main>
		</q-item>
	</q-list>

	<networkinterface2 v-if="current" :interface="current"></networkinterface2>

</q-card>
<dhcpconfig2 :interface="current" v-if="currentWithDhcp"></dhcpconfig2>


<div class="network-master">
Interface selection <select v-model="current">
	<option v-for="iface in interfaces" :key="iface.name" :value="iface">{{ iface.name }}</option>
</select>
<networkinterface v-if="current" :interface="current"></networkinterface>

</div>

</q-page>

`


const templateDHCPConfig2 = `
<q-card inline class="q-ma-sm">
	<q-card-title>
    	DHCP Server settings for {{ interface.name }}
	</q-card-title>


	<q-list link>
		<q-item-separator />
		<q-item tag="label">
			<q-item-side>
				<q-toggle v-model="authoritative"></q-toggle>
			</q-item-side>
			<q-item-main>
				<q-item-tile label>Authoritative</q-item-tile>
				<q-item-tile sublabel>If disabled, the DHCP Server isn't authoritative</q-item-tile>
			</q-item-main>
		</q-item>
		<q-item tag="label" disabled>
			<q-item-main>
				<q-item-tile label>Path to lease file</q-item-tile>
				<q-item-tile sublabel>{{ config.leaseFile }}</q-item-tile>
			</q-item-main>
		</q-item>

	</q-list>

		<q-item tag="label">
			<q-item-main>
				<q-item-tile>

					<q-table
						title="DHCP ranges"
						:data="config.ranges"
						:columns="[{name:'lower', field: 'rangeLower', label: 'Lower IP', align: 'left'}, {name:'upper', field: 'rangeUpper', label: 'Upper IP', align: 'left'}, {name:'lease', field: 'leaseTime', label: 'Lease Time', align: 'left'}, {name:'remove', label: 'Delete range', align: 'left'}]"
						row-key="name"
					>

      					<q-tr slot="header" slot-scope="props" :props="props">
        					<q-th :key="props.cols[0].name" :props="props"> {{ props.cols[0].label }} </q-th>
        					<q-th :key="props.cols[1].name" :props="props"> {{ props.cols[1].label }} </q-th>
        					<q-th :key="props.cols[2].name" :props="props"> {{ props.cols[2].label }} </q-th>
        					<q-th :key="props.cols[3].name" :props="props">
								<q-btn @click="addRange()">add</q-btn>
							</q-th> 
						</q-tr>

						<q-tr slot="body" slot-scope="props" :props="props">
							<q-td key="lower" :props="props">
								<q-input v-model="props.row.rangeLower" inverted></q-input>
							</q-td>
							<q-td key="upper" :props="props">
								<q-input v-model="props.row.rangeUpper" inverted></q-input>
							</q-td>
        					<q-td key="lease" :props="props">
								<q-input v-model="props.row.leaseTime" inverted></q-input>
							</q-td>
        					<q-td key="remove" :props="props">
								<q-btn @click="removeRange(props.row)">del</q-btn>
							</q-td>
        					
						</q-tr>

					</q-table>

				</q-item-tile>
			</q-item-main>
		</q-item>
		

	<networkinterface2 v-if="current" :interface="current"></networkinterface2>
</q-card>
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

const templateNetworkInterface2 = `
	<q-list link>
		<q-item-separator />
		<q-list-header>Generic settings for {{interface.name}}</q-list-header>
		<q-item tag="label">
			<q-item-side>
				<q-toggle v-model="interface.enabled"></q-toggle>
			</q-item-side>
			<q-item-main>
				<q-item-tile label>Enabled</q-item-tile>
				<q-item-tile sublabel>Enable/Disable interface</q-item-tile>
			</q-item-main>
		</q-item>

		<q-item tag="label">
			<q-item-main>
				<q-item-tile label>Mode</q-item-tile>
				<q-item-tile sublabel>Enable DHCP server, client or manual configuration</q-item-tile>
				<q-item-tile>
					<q-select v-model="interface.mode" :options="selectOptionsInterfaceModes" inverted></q-select>
				</q-item-tile>
			</q-item-main>
		</q-item>

		<q-item tag="label" v-show="withIP">
			<q-item-main>
				<q-item-tile label>IP</q-item-tile>
				<q-item-tile sublabel>IPv4 address of interface in dotted decimal (f.e. 172.16.0.1)</q-item-tile>
				<q-item-tile>
					<q-input v-model="interface.ipAddress4" inverted></q-input>
				</q-item-tile>
			</q-item-main>
		</q-item>

		<q-item tag="label" v-show="withIP">
			<q-item-main>
				<q-item-tile label>Netmask</q-item-tile>
				<q-item-tile sublabel>Netmask of interface in dotted decimal (f.e. 255.255.255.0)</q-item-tile>
				<q-item-tile>
					<q-input v-model="interface.netmask4" inverted></q-input>
				</q-item-tile>
			</q-item-main>
		</q-item>

	</q-list link>

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
