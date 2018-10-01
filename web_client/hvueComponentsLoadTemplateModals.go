package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/mame82/hvue"
)


type modalLoadBashScriptData struct {
	*js.Object
	IsVisible bool `js:"isVisible"`
}

func InitCompsLoadModals() {
	hvue.NewComponent(
		"modalLoadBashScript",
		hvue.Template(templateLoadBashScript),
		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := &modalLoadBashScriptData{Object:O()}
			data.IsVisible = false

			return data
		}),
		hvue.PropObj(
			"show",
			hvue.Types(hvue.PBoolean),
			),
		hvue.Method("setVisible",
			func(vm *hvue.VM, visible bool) {
				data := &modalLoadBashScriptData{Object:vm.Data}
				data.IsVisible = visible
			},
		),
		hvue.Mounted(func(vm *hvue.VM) {
			data := &modalLoadBashScriptData{Object:vm.Data}
			data.IsVisible = vm.Get("show").Bool()
			// ToDo: update BashScriptList via vuex store action
			return
		}),
	)
}

const templateLoadBashScript = `
	<q-modal v-model="isVisible">
		<q-modal-layout>
			<q-toolbar slot="header">
				<q-toolbar-title>
					Load WiFi settings
				</q-toolbar-title>
			</q-toolbar>

			<q-list>
				<q-item link tag="label" v-for="tname in this.$store.state.StoredWifiSettingsList" :key="tname">
					<q-item-side>
						<q-radio v-model="templateName" :val="tname"/>
					</q-item-side>
					<q-item-main>
						<q-item-tile label>{{ tname }}</q-item-tile>
					</q-item-main>
				</q-item>
				<q-item tag="label">
					<q-item-main>
						<q-item-tile>
							<q-btn color="secondary" v-close-overlay label="close" />
							<q-btn color="primary" label="load" />
							<q-btn color="primary" label="deploy" />
						</q-item-tile>
					</q-item-main>
				</q-item>
			</q-list>
		</q-modal-layout>
	</q-modal>

`
