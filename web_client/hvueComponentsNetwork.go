// +build js

package main

import (
	//"github.com/mame82/hvue"
	"github.com/gopherjs/gopherjs/js"
	pb "github.com/mame82/P4wnP1_aloa/proto/gopherjs"

	"github.com/mame82/hvue"
)

type jsDataTablePagination struct {
	*js.Object
	RowsPerPage int  `js:"rowsPerPage"`
	Descending  bool `js:"descending"`
	Page        int  `js:"page"`
}

func newPagination(rowsPerPage int, startPage int) (res *jsDataTablePagination) {
	res = &jsDataTablePagination{Object: O()}
	res.RowsPerPage = rowsPerPage
	res.Page = startPage
	res.Descending = false
	return
}

func InitComponentsNetwork() {

	hvue.NewComponent(
		"network",
		hvue.Template(templateNetwork),
		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := &struct {
				*js.Object
				CurrentInterface int `js:"currentIdx"`
				ShowStoreModal bool   `js:"showStoreModal"`
				ShowLoadModal bool   `js:"showLoadModal"`
				ShowDeployStoredModal bool   `js:"showDeployStoredModal"`
			}{Object: O()}
			data.CurrentInterface = 0
			data.ShowStoreModal = false
			data.ShowLoadModal = false
			data.ShowDeployStoredModal = false
			return data
		}),

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
					Value int `js:"value"`
				}{Object: O()}
				currentIf := &jsEthernetInterfaceSettings{
					Object: interfaces.Index(i),
				}
				option.Label = currentIf.Name
				option.Value = i
				selectIf.Call("push", option)
			}
			return selectIf
		}),
		hvue.Computed("current", func(vm *hvue.VM) interface{} {
			interfaces := vm.Get("$store").Get("state").Get("InterfaceSettings").Get("interfaces")
			idx := vm.Get("currentIdx").Int()
			currentIface := interfaces.Index(idx)
			if currentIface == js.Undefined {
				return &jsEthernetInterfaceSettings{Object:O()}
			}
			return currentIface
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
		hvue.Method("updateStoredSettingsList",
			func(vm *hvue.VM) {
				vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_STORED_ETHERNET_INTERFACE_SETTINGS_LIST)
			}),
		hvue.Method("store",
			func(vm *hvue.VM, name *js.Object) {
				sReq := NewEthernetRequestSettingsStorage()
				sReq.TemplateName = name.String()
				sReq.Settings = &jsEthernetInterfaceSettings{
					Object: vm.Get("current"),
				}
				println("Storing :", sReq)
				vm.Get("$store").Call("dispatch", VUEX_ACTION_STORE_ETHERNET_INTERFACE_SETTINGS, sReq)
				vm.Set("showStoreModal", false)
			}),
		hvue.Method("load",
			func(vm *hvue.VM, name *js.Object) {
				println("Loading :", name.String())
				vm.Get("$store").Call("dispatch", VUEX_ACTION_LOAD_ETHERNET_INTERFACE_SETTINGS, name)
			}),
		hvue.Method("deployStored",
			func(vm *hvue.VM, name *js.Object) {
				println("Loading :", name.String())
				vm.Get("$store").Call("dispatch", VUEX_ACTION_DEPLOY_STORED_ETHERNET_INTERFACE_SETTINGS, name)
			}),
		hvue.Method("deleteStored",
			func(vm *hvue.VM, name *js.Object) {
				println("Deleting template :", name.String())
				vm.Get("$store").Call("dispatch", VUEX_ACTION_DELETE_STORED_ETHERNET_INTERFACE_SETTINGS, name)
			}),

		// The following method doesn't make much sense anymore, but is kept as an example for working with promises
		hvue.Mounted(func(vm *hvue.VM) {
			// update network interface
			println("ethernet settings component mounted")
			promise := vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_ALL_ETHERNET_INTERFACE_SETTINGS)

			promise.Call("then",
				func() {
					/*
					println("Mounting network interface settings, try to select first interface")
					// current" could be an empty settings object, set to first interface of computed property "interfaces" (if there is one)
					interfaces := vm.Get("$store").Get("state").Get("InterfaceSettings").Get("interfaces")
					if interfaces.Length() > 0 {
						hvue.Set(vm, "current", interfaces.Index(0))
						println("... current is", vm.Get("current"))
					} else {
						println("... No interface found")
					}
					*/
					println("ethernet interface settings reloaded")
					hvue.Set(vm, "currentIdx", 0)
				},
				func() {
					println("error in THEN ")
				},
			)

		}),

	)

	hvue.NewComponent(
		"network-interface-settings",
		hvue.Template(templateNetworkInterface),
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

	hvue.NewComponent("dhcp-config",
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

	hvue.NewComponent("dhcp-ranges",
		hvue.Template(templateDHCPRanges),
		hvue.Props("config"),
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
		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := &struct {
				*js.Object
				Pagination *jsDataTablePagination `js:"pagination"`
			}{Object: O()}
			data.Pagination = newPagination(3, 1)
			return data
		}),
	)

	hvue.NewComponent("dhcp-options",
		hvue.Props("config"),
		hvue.Template(templateDHCPOptions),
		hvue.Method("addOption", func(vm *hvue.VM) {
			s := &jsDHCPServerSettings{Object: vm.Get("config")}
			o := &jsDHCPServerOption{Object: O()}
			o.Option = 3
			o.Value = ""
			s.AddOption(o)
		}),
		hvue.Method("removeOption", func(vm *hvue.VM, delOption *jsDHCPServerOption) {
			s := &jsDHCPServerSettings{Object: vm.Get("config")}
			s.RemoveOption(delOption)
		}),
		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := &struct {
				*js.Object
				Pagination *jsDataTablePagination `js:"pagination"`
			}{Object: O()}
			data.Pagination = newPagination(3, 1)
			return data
		}),
	)

	hvue.NewComponent("dhcp-static-hosts",
		hvue.Props("config"),
		hvue.Template(templateDHCPStaticHosts),
		hvue.Method("addStaticHost", func(vm *hvue.VM) {
			s := &jsDHCPServerSettings{Object: vm.Get("config")}
			sh := &jsDHCPServerStaticHost{Object: O()}
			sh.Ip = ""
			sh.Mac = ""
			s.AddStaticHost(sh)
		}),
		hvue.Method("removeStaticHost", func(vm *hvue.VM, delStaticHost *jsDHCPServerStaticHost) {
			s := &jsDHCPServerSettings{Object: vm.Get("config")}
			s.RemoveStaticHost(delStaticHost)
		}),
		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := &struct {
				*js.Object
				Pagination *jsDataTablePagination `js:"pagination"`
			}{Object: O()}
			data.Pagination = newPagination(3, 1)
			return data
		}),
	)

}

const templateNetwork = `
<q-page padding>
	<select-string-from-array :values="$store.state.StoredEthernetInterfaceSettingsList" v-model="showLoadModal" title="加载以太网接口设置" @load="load($event)" @delete="deleteStored($event)" with-delete></select-string-from-array>
	<select-string-from-array :values="$store.state.StoredEthernetInterfaceSettingsList" v-model="showDeployStoredModal" title="应用已保存的以太网接口设置" @load="deployStored($event)" @delete="deleteStored($event)" with-delete></select-string-from-array>
	<modal-string-input v-model="showStoreModal" title="保存当前的以太网接口设置" @save="store($event)"></modal-string-input>

	<div class="row gutter-sm">

		<div class="col-12">
			<q-card>
				<q-card-title>
					网络接口设置
				</q-card-title>

				<q-card-main>
					<div class="row gutter-sm">

						<div class="col-6 col-sm""><q-btn class="fit" color="primary" @click="deploy(current)" label="应用" icon="launch"></q-btn></div>
						<div class="col-6 col-sm""><q-btn class="fit" color="primary" @click="updateStoredSettingsList(); showDeployStoredModal=true" label="应用已保存" icon="settings_backup_restore"></q-btn></div>
<!--
						<div class="col-6 col-sm""><q-btn class="fit" color="secondary" @click="UpdateFromDeployedGadgetSettings" label="重置" icon="autorenew"></q-btn></div>
-->
						<div class="col-6 col-sm""><q-btn class="fit" color="secondary" @click="showStoreModal=true" label="保存" icon="cloud_upload"></q-btn></div>
						<div class="col-6 col-sm"><q-btn class="fit" color="warning" @click="updateStoredSettingsList(); showLoadModal=true" label="加载已保存" icon="cloud_download"></q-btn></div>

					</div>
  				</q-card-main>


			</q-card>
		</div>


		<div class="col-12 col-xl-3">
		<q-card class="full-height">
			<q-card-title>
		    	通用设置
			</q-card-title>

			<q-list link>
				<q-item-separator />
				<q-item tag="label">
					<q-item-main>
						<q-item-tile label>接口</q-item-tile>
						<q-item-tile sublabel>选择一个接口以配置</q-item-tile>
						<q-item-tile>
							<q-select v-model="currentIdx" :options="selectOptionsInterface" color="secondary" inverted></q-select>
						</q-item-tile>
					</q-item-main>
				</q-item>
			</q-list>

			<network-interface-settings v-if="current" :interface="current"></network-interface-settings>
		</q-card>
		</div>

		<div class="col-12 col-xl-9" v-if="currentWithDhcp">
			<dhcp-config :interface="current"></dhcp-config>
		</div>
	</div>
</q-page>

`

const templateDHCPConfig = `
<q-card>
	<q-card-title>
    	为 {{ interface.name }} 设置DHCP服务器
	</q-card-title>


	<q-list>
		<q-item-separator />
		<q-item tag="label" link>
			<q-item-side>
				<q-toggle v-model="authoritative"></q-toggle>
			</q-item-side>
			<q-item-main>
				<q-item-tile label>可信的</q-item-tile>
				<q-item-tile sublabel>如果禁用，则DHCP服务器不具有权威性</q-item-tile>
			</q-item-main>
		</q-item>
		<q-item tag="label" disabled link>
			<q-item-main>
				<q-item-tile label>凭据文件的路径</q-item-tile>
				<q-item-tile sublabel>{{ config.leaseFile }}</q-item-tile>
			</q-item-main>
		</q-item>

		<q-list-header>DHCP范围</q-list-header>
		<dhcp-ranges :config="config"></dhcp-ranges>

		<q-list-header>DHCP选项</q-list-header>
		<dhcp-options :config="config"></dhcp-options>

		<q-list-header>DHCP静态主机</q-list-header>
		<dhcp-static-hosts :config="config"></dhcp-static-hosts>
	</q-list>
</q-card>
`

const templateDHCPRanges = `
<q-item :link="false">
	<q-item-main>
		<q-item-tile>
			<q-table
				:data="config.ranges"
				:columns="[{name:'lower', field: 'rangeLower', label: '起始IP', align: 'left'}, {name:'upper', field: 'rangeUpper', label: '结束IP', align: 'left'}, {name:'lease', field: 'leaseTime', label: '租期', align: 'left'}, {name:'remove', label: '删除范围', align: 'left'}]"
				row-key="name"
				:pagination.sync="pagination"
				v-if="$q.platform.is.desktop"
			>
				<q-tr slot="header" slot-scope="props" :props="props">
					<q-th :key="props.cols[0].name" :props="props"> {{ props.cols[0].label }} </q-th>
					<q-th :key="props.cols[1].name" :props="props"> {{ props.cols[1].label }} </q-th>
					<q-th :key="props.cols[2].name" :props="props"> {{ props.cols[2].label }} </q-th>
					<q-th :key="props.cols[3].name" :props="props">
					<q-btn @click="addRange()">添加</q-btn>
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

<!--
			<q-table
				:data="config.ranges"
				:columns="[{name:'lower', field: 'rangeLower', label: 'Lower IP', align: 'left'}, {name:'upper', field: 'rangeUpper', label: 'Upper IP', align: 'left'}, {name:'lease', field: 'leaseTime', label: 'Lease Time', align: 'left'}, {name:'remove', label: 'Delete range', align: 'left'}]"
				row-key="name"
				:pagination.sync="pagination"
				
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
						{{ props.row.rangeLower }}
						<q-popup-edit v-model="props.row.rangeLower" title="Set lower IP" persistent buttons>
							<q-input v-model="props.row.rangeLower" inverted></q-input>
						</q-popup-edit>
					</q-td>
					<q-td key="upper" :props="props">
						{{ props.row.rangeUpper }}
						<q-popup-edit v-model="props.row.rangeUpper" title="Set upper IP">
							<q-input v-model="props.row.rangeUpper" inverted></q-input>
						</q-popup-edit>
					</q-td>
					<q-td key="lease" :props="props">
						{{ props.row.leaseTime }}
						<q-popup-edit v-model="props.row.leaseTime" title="Set lease time">
							<q-input v-model="props.row.leaseTime" inverted></q-input>
						</q-popup-edit>
					</q-td>
					<q-td key="remove" :props="props">
						<q-btn @click="removeRange(props.row)">del</q-btn>
					</q-td>	
				</q-tr>
			</q-table>
-->

			<q-card v-else>
				<q-card-main>
					<q-table
						:data="config.ranges"
						:columns="[{name:'lower', field: 'rangeLower', label: '起始IP', align: 'left'}, {name:'upper', field: 'rangeUpper', label: '结束IP', align: 'left'}, {name:'lease', field: 'leaseTime', label: '租期', align: 'left'}, {name:'remove', label: '删除范围', align: 'left'}]"
						row-key="name"
						hide-header
						:pagination.sync="pagination"
						grid
					>
						<template slot="top-right" slot-scope="props" class="q-mr-sm">
							<q-btn @click="addRange()" color="primary">添加范围</q-btn>
						</template>
						<div
							slot="item"
							slot-scope="props"
							class="col-lg-6"
						>
							<q-card-separator />
							<q-list link no-border>
								<q-item tag="label" :key="remove">
									<q-item-main>
										<q-btn @click="removeRange(props.row)" color="secondary">删除范围 {{ props.row.__index + 1 }}</q-btn>
									</q-item-main>
								</q-item>
								<q-item tag="label" :key="props.colsMap.lower.name">
									<q-item-main>
										<q-item-tile label>
											{{ props.colsMap.lower.label }}
										</q-item-tile>
										<q-item-tile>
											<q-input v-model="props.row.rangeLower" inverted></q-input>
										</q-item-tile>
									</q-item-main>
								</q-item>
								<q-item tag="label" :key="props.colsMap.upper.name">
									<q-item-main>
										<q-item-tile label>
											{{ props.colsMap.upper.label }}
										</q-item-tile>
										<q-item-tile>
											<q-input v-model="props.row.rangeUpper" inverted></q-input>
										</q-item-tile>
									</q-item-main>
								</q-item>
								<q-item tag="label" :key="props.colsMap.lease.name">
									<q-item-main>
										<q-item-tile label>
											{{ props.colsMap.lease.label }}
										</q-item-tile>
										<q-item-tile>
											<q-input v-model="props.row.leaseTime" inverted></q-input>
										</q-item-tile>
									</q-item-main>
								</q-item>
							</q-list>
						</div>
					</q-table>
				</q-card-main>
			</q-card>

		</q-item-tile>
	</q-item-main>
</q-item>
`

const templateDHCPOptions = `
<q-item :link="false">
	<q-item-main>
		<q-item-tile>
			<q-table
				:data="config.options"
				:columns="[{name:'optnumber', field: 'number', label: 'Option number (RFC 2132)', align: 'left'}, {name:'optvalue', field: 'value', label: 'Option string', align: 'left'}, {name:'remove', label: 'Delete option', align: 'left'}]"
				row-key="name"
				:pagination.sync="pagination"
				v-if="$q.platform.is.desktop"
			>
				<q-tr slot="header" slot-scope="props" :props="props">
					<q-th :key="props.cols[0].name" :props="props"> {{ props.cols[0].label }} </q-th>
					<q-th :key="props.cols[1].name" :props="props"> {{ props.cols[1].label }} </q-th>
					<q-th :key="props.cols[2].name" :props="props">
						<q-btn @click="addOption()">add</q-btn>
					</q-th>
				</q-tr>
				
				<q-tr slot="body" slot-scope="props" :props="props">
					<q-td key="optnumber" :props="props">
						<q-input v-model="props.row.option" type="number" inverted></q-input>
					</q-td>
					<q-td key="optvalue" :props="props">
						<q-input v-model="props.row.value" inverted></q-input>
					</q-td>
					<q-td key="remove" :props="props">
						<q-btn @click="removeOption(props.row)">del</q-btn>
					</q-td>	
				</q-tr>
			</q-table>

			<q-card v-else>
				<q-card-main>
					<q-table
						:data="config.options"
						:columns="[{name:'optnumber', field: 'number', label: 'Option number (RFC 2132)', align: 'left'}, {name:'optvalue', field: 'value', label: 'Option string', align: 'left'}, {name:'remove', label: 'Delete option', align: 'left'}]"
						row-key="name"
						hide-header
						:pagination.sync="pagination"
						grid
					>
						<template slot="top-right" slot-scope="props" class="q-mr-sm">
							<q-btn @click="addOption()" color="primary">add option</q-btn>
						</template>
						<div
							slot="item"
							slot-scope="props"
							class="col-lg-6"
						>
							<q-card-separator />
							<q-list link no-border>
								<q-item tag="label" :key="remove">
									<q-item-main>
										<q-btn @click="removeOption(props.row)" color="secondary">delete option {{ props.row.__index + 1 }}</q-btn>
									</q-item-main>
								</q-item>
								<q-item tag="label" :key="props.colsMap.optnumber.name">
									<q-item-main>
										<q-item-tile label>
											{{ props.colsMap.optnumber.label }}
										</q-item-tile>
										<q-item-tile>
											<q-input v-model="props.row.option" type="number" inverted></q-input>
										</q-item-tile>
									</q-item-main>
								</q-item>
								<q-item tag="label" :key="props.colsMap.optvalue.name">
									<q-item-main>
										<q-item-tile label>
											{{ props.colsMap.optvalue.label }}
										</q-item-tile>
										<q-item-tile>
											<q-input v-model="props.row.value" inverted></q-input>
										</q-item-tile>
									</q-item-main>
								</q-item>
							</q-list>

						</div>
					</q-table>
				</q-card-main>
			</q-card>

		</q-item-tile>
	</q-item-main>
</q-item>
`

const templateDHCPStaticHosts = `
<q-item :link="false">
	<q-item-main>
		<q-item-tile>
			<q-table
				:data="config.staticHosts"
				:columns="[{name:'hostmac', field: 'mac', label: 'Host MAC', align: 'left'}, {name:'hostip', field: 'ip', label: 'Host IP', align: 'left'}, {name:'remove', label: 'Delete static host', align: 'left'}]"
				row-key="name"
				:pagination.sync="pagination"
				v-if="$q.platform.is.desktop"
			>
				<q-tr slot="header" slot-scope="props" :props="props">
					<q-th :key="props.cols[0].name" :props="props"> {{ props.cols[0].label }} </q-th>
					<q-th :key="props.cols[1].name" :props="props"> {{ props.cols[1].label }} </q-th>
					<q-th :key="props.cols[2].name" :props="props">
						<q-btn @click="addStaticHost()">add</q-btn>
					</q-th>
				</q-tr>
				
				<q-tr slot="body" slot-scope="props" :props="props">
					<q-td key="hostmac" :props="props">
						<q-input v-model="props.row.mac" inverted></q-input>
					</q-td>
					<q-td key="hostip" :props="props">
						<q-input v-model="props.row.ip" inverted></q-input>
					</q-td>
					<q-td key="remove" :props="props">
						<q-btn @click="removeStaticHost(props.row)">del</q-btn>
					</q-td>	
				</q-tr>
			</q-table>

			<q-card v-else>
				<q-card-main>
					<q-table
						:data="config.staticHosts"
						:columns="[{name:'hostmac', field: 'mac', label: 'Host MAC', align: 'left'}, {name:'hostip', field: 'ip', label: 'Host IP', align: 'left'}, {name:'remove', label: 'Delete static host', align: 'left'}]"
						row-key="name"
						hide-header
						:pagination.sync="pagination"
						grid
					>
						<template slot="top-right" slot-scope="props" class="q-mr-sm">
							<q-btn @click="addStaticHost()" color="primary">add static host</q-btn>
						</template>
						<div
							slot="item"
							slot-scope="props"
							class="col-lg-6"
						>
							<q-card-separator />
							<q-list link no-border>
								<q-item tag="label" :key="remove">
									<q-item-main>
										<q-btn @click="removeStaticHost(props.row)" color="secondary">delete</q-btn>
									</q-item-main>
								</q-item>
								<q-item tag="label" :key="props.colsMap.hostmac.name">
									<q-item-main>
										<q-item-tile label>
											{{ props.colsMap.hostmac.label }}
										</q-item-tile>
										<q-item-tile>
											<q-input v-model="props.row.mac" inverted></q-input>
										</q-item-tile>
									</q-item-main>
								</q-item>
								<q-item tag="label" :key="props.colsMap.hostip.name">
									<q-item-main>
										<q-item-tile label>
											{{ props.colsMap.hostip.label }}
										</q-item-tile>
										<q-item-tile>
											<q-input v-model="props.row.ip" inverted></q-input>
										</q-item-tile>
									</q-item-main>
								</q-item>
							</q-list>

						</div>
					</q-table>
				</q-card-main>
			</q-card>

		</q-item-tile>
	</q-item-main>
</q-item>
`

const templateNetworkInterface = `
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
