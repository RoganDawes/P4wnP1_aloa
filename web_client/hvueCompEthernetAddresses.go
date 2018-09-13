// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/mame82/hvue"
)

type CompEthernetAddressesData2 struct {
	*js.Object

}

func newCompEthernetAddressesData2(vm *hvue.VM) interface{} {

	cc := &CompEthernetAddressesData2{
		Object: js.Global.Get("Object").New(),
	}


	return cc
}

func InitCompEthernetAddresses2() {
	/*
	o := vue.NewOption()
	o.Name = "EthernetAddresses"
	o.SetDataWithMethods(newCompEthernetAddressesData2)
	o.Template = compEthernetAddressesTemplate2
	o.AddProp("settings")
	*/

	hvue.NewComponent(
		"ethernet-addresses",
		hvue.Template(compEthernetAddressesTemplate2),
		hvue.DataFunc(newCompEthernetAddressesData2),
		hvue.PropObj("settings", hvue.Types(hvue.PObject)),
	)
}

const (

	compEthernetAddressesTemplate2 = `
<div>
	<table>
	<tr>
		<td>Host MAC address</td><td><input v-bind:value="settings.HostAddr" v-on:input="$emit('hostAddrChange', $event.target.value)"></td>
	</tr>
	<tr>
		<td>Device MAC address</td><td><input v-bind:value="settings.DevAddr" v-on:input="$emit('devAddrChange', $event.target.value)"></td>
	</tr>
	</table>
</div>
`
)



