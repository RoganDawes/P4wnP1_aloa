package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/mame82/hvue"
)

type CompTabData struct {
	*js.Object
	Id int `js:"id"`
	IsActive bool `js:"isActive"`
}

func NewCompTabData(vm *hvue.VM) interface{} {

	cc := &CompTabData{
		Object: js.Global.Get("Object").New(),
	}

	cc.Id = -1
	cc.IsActive = false
	return cc
}



func InitCompTab()  {


	hvue.NewComponent(
		"tab",
		hvue.Template(compTabTemplate),
		hvue.DataFunc(NewCompTabData),
		hvue.PropObj("header", hvue.Types(hvue.PString), hvue.Default("tab name"), hvue.Required),
		hvue.PropObj("disabled", hvue.Types(hvue.PBoolean), hvue.Default(false)),
		hvue.PropObj("selected", hvue.Types(hvue.PBoolean), hvue.Default(false)),

		/*
		hvue.Computed("active", func(vm *hvue.VM) interface{} {
			return true
		}),
		*/
		hvue.Computed("_isTab", func(vm *hvue.VM) interface{} {
			return true
		}),

		hvue.Computed("hasOverride", func(vm *hvue.VM) interface{} {
			return vm.Slots.Get("override").Bool()
		}),

		hvue.Mounted(func(vm *hvue.VM) {
			vm.Set("isActive", vm.Get("selected")) //propagate "selected" property over to "isActive" from data
		}),
	)
	//return o.NewComponent()
}

const (

	compTabTemplate = `
		<div v-if="isActive">
			<slot></slot>
		</div>
	
`
)

