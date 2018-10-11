package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/mame82/hvue"
)

func InitComponentsDialog() {
	hvue.NewComponent(
		"select-string-from-array",
		hvue.Template(templateSelectStringModal),
		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := struct {
				*js.Object
				CurrentSelection *js.Object `js:"CurrentSelection"`
			}{Object:O()}
			data.CurrentSelection = O()

			return &data
		}),
		hvue.Mounted(func(vm *hvue.VM) {
			vm.Set("CurrentSelection", vm.Get("values").Index(0))
			println("Index 0 on mount", vm.Get("values").Index(0))
		}),
		hvue.ComputedWithGetSet(
			"visible",
			func(vm *hvue.VM) interface{} {
				return vm.Get("value")
			},
			func(vm *hvue.VM, newValue *js.Object) {
				vm.Call("$emit", "input", newValue)
			},
			),
		hvue.Method(
			"onLoadPressed",
			func(vm *hvue.VM) {
				vm.Call("$emit", "load", vm.Get("CurrentSelection"))
				println(vm.Get("CurrentSelection"))

			},
		),
		hvue.PropObj(
			"values",
			hvue.Required,
			),
		hvue.PropObj(
			"value",
			hvue.Required,
			hvue.Types(hvue.PBoolean),
			),
		hvue.PropObj(
			"title",
			hvue.Types(hvue.PString),
			),
	)
}

const templateSelectStringModal = `
	<q-modal v-model="visible">
		<q-modal-layout>
			<q-toolbar slot="header">
				<q-toolbar-title>
					{{ title }}
				</q-toolbar-title>
			</q-toolbar>

			<q-list>
				<q-item link tag="label" v-for="name in values" :key="name">
					<q-item-side>
						<q-radio v-model="CurrentSelection" :val="name"/>
					</q-item-side>
					<q-item-main>
						<q-item-tile label>{{ name }}</q-item-tile>
					</q-item-main>
				</q-item>
				<q-item tag="label">
					<q-item-main>
						<q-item-tile>
							<q-btn color="primary" v-show="CurrentSelection != undefined" label="load" @click="onLoadPressed(); visible=false"/>							
							<q-btn color="secondary" v-close-overlay label="close" />
						</q-item-tile>
					</q-item-main>
				</q-item>
			</q-list>
		</q-modal-layout>
	</q-modal>
`