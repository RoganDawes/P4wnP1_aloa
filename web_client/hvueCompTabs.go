package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/HuckRidgeSW/hvue"
)

type CompTabsData struct {
	*js.Object
	headers []string `js:"headers"`
	tabs []*js.Object `js:"tabs"`
}

func NewCompTabsData(vm *hvue.VM) interface{} {

	cc := &CompTabsData{
		Object: js.Global.Get("Object").New(),
	}
	cc.headers = []string{}
	cc.tabs = []*js.Object{}

	return cc
}

func initTabs(vm *hvue.VM) {
	// ToDo: clear children to allow dynamic adding/removing of tabs
	id := 0
	for _,child := range vm.Children {
		isTab := child.Get("_isTab").Bool()
		//println(child)
		//println(isTab)


		if isTab {
			vm.Data.Get("tabs").Call("push", child)
			child.Set("id", id)
			id++
		}

	}

	//ToDo: remove export (Debug only)
	js.Global.Set("vm", vm)
}

func (c *CompTabsData) UpdateSelectedTab(vm *hvue.VM, selectedID int) {
	println("Update selected ID: ", selectedID)
	for _,child := range vm.Children {
		isTab := child.Get("_isTab").Bool()
		if isTab {
			child.Set("isActive", child.Get("id").Int() == selectedID) //child.isActive = (selectedID == child.id)
		}
	}


}

func InitCompTabs() {
	hvue.NewComponent(
		"tabs",
		hvue.DataFunc(NewCompTabsData),
		hvue.Template(compTabsTemplate),
		hvue.Mounted(initTabs),
		hvue.MethodsOf(&CompTabsData{}),
	)
}

const (


	compTabsTemplate = `
<div>
	<div>
		<ul class="nav nav-tabs">
	        <li v-for="t in tabs" :class="{ 'active' : t.isActive }">
	          <a href="#" @click="UpdateSelectedTab(t.id)">{{t.id}}: {{ t.header }}</slot></a>
	        </li>
		</ul>
	</div>
	<div class="tab-content">
		<slot></slot>
	</div>
</div>
`
)

