package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/HuckRidgeSW/hvue"
)

type CompToggleSwitchData struct {
	*js.Object

}

func newCompToggleSwitchData(vm *hvue.VM) interface{} {
	newVM := &CompToggleSwitchData{
		Object: js.Global.Get("Object").New(),
	}
	return newVM
}

func InitCompToggleSwitch() {
	hvue.NewComponent(
		"toggle-switch",
		hvue.Template(compToggleSwitchTemplate),
		hvue.DataFunc(newCompToggleSwitchData),
		hvue.PropObj("value", hvue.Types(hvue.PBoolean), hvue.Required),
		)
}

const (

	compToggleSwitchTemplate = `
<label class="toggle-switch">
   	<input type="checkbox" v-bind:checked="value" v-on:change="$emit('input', $event.target.checked)">
   	<div><span class="on">On</span><span class="off">Off</span></div>
	<span class="toggle-switch-slider"></span>
</label>
`
)

