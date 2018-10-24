// +build js

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
				//println(vm.Get("CurrentSelection"))

			},
		),
		hvue.Method(
			"onDeletePressed",
			func(vm *hvue.VM, name *js.Object) {
				vm.Call("$emit", "delete", name)
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
			"with-delete",
			hvue.Types(hvue.PBoolean),
			hvue.Default(false),
			),
		hvue.PropObj(
			"title",
			hvue.Types(hvue.PString),
			),
	)

	hvue.NewComponent(
		"modal-string-input",
		hvue.Template(templateInputStringModal),
		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := struct {
				*js.Object
				Text string `js:"text"`
			}{Object:O()}
			data.Text = ""

			return &data
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
			"onSavePressed",
			func(vm *hvue.VM) {
				if vm.Get("text") == js.Undefined || vm.Get("text").String() == "" {
					QuasarNotifyError("Can't store to empty template name", "", QUASAR_NOTIFICATION_POSITION_TOP)
					return
				}
				vm.Call("$emit", "save", vm.Get("text"))
				println(vm.Get("text"))

			},
		),
		hvue.PropObj(
			"title",
			hvue.Types(hvue.PString),
			),
		hvue.PropObj(
			"value",
			hvue.Required,
			hvue.Types(hvue.PBoolean),
		),
	)

	hvue.NewComponent(
		"ransom-note",
		hvue.Template(templateRansomModal),
		hvue.ComputedWithGetSet(
			"visible",
			func(vm *hvue.VM) interface{} {
				return vm.Get("value")
			},
			func(vm *hvue.VM, newValue *js.Object) {
				vm.Call("$emit", "input", newValue)
			},
			),
		hvue.PropObj(
			"value",
			hvue.Required,
			hvue.Types(hvue.PBoolean),
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
					<q-item-side v-if="withDelete" right>
						<q-btn icon="delete" color="negative" @click="onDeletePressed(name)" round flat />
					</q-item-side>
				</q-item>
				<q-item tag="label">
					<q-item-main>
						<q-item-tile>
							<q-btn color="primary" v-show="CurrentSelection != undefined" label="ok" @click="onLoadPressed(); visible=false"/>							
							<q-btn color="secondary" v-close-overlay label="close" />
						</q-item-tile>
					</q-item-main>
				</q-item>
			</q-list>
		</q-modal-layout>
	</q-modal>
`

const templateInputStringModal = `
<div>
	<q-modal v-model="visible">
		<q-modal-layout>
			<q-toolbar slot="header">
				<q-toolbar-title>
					{{ title }}
				</q-toolbar-title>
			</q-toolbar>
			<q-list>
				<q-item tag="label">
					<q-item-main>
						<q-item-tile label>Name</q-item-tile>
						<q-item-tile>
							<q-input v-model="text" inverted></q-input>
						</q-item-tile>
					</q-item-main>
				</q-item>
				<q-item tag="label">
					<q-item-main>
						<q-item-tile>
							<q-btn color="primary" v-show="text != undefined && text.length > 0" @click="onSavePressed(); visible=false" label="store" />
							<q-btn color="secondary" v-close-overlay label="close" />
						</q-item-tile>
					</q-item-main>
				</q-item>
			</q-list>
		</q-modal-layout>
	</q-modal>
</div>
`
const templateRansomModal = `
<div>
	<q-modal v-model="visible" content-css="background: red;" no-route-dismiss no-backdrop-dismiss>
			<div style="color: white; font-size: 1.5em; font-family: monospace; padding: 10%">
				
You became victim of a VERY SILLY IDEA</br>
███████████████████████████████████████████████████████████████████████████████</br>
</br>
The web page you've been viewing, provided a sophisticated experience</br>
in terms of keyboard automation and scripting. There were LED based triggers,</br>
there was scriptable mouse control, there were complex control structures</br>
like if-else-branching and for-loops. Not to mention the capability of running</br>
multiple asynchronous jobs.</br>
</br>
If you really need a converter for a limited, old-school language:</br>
</br>
1. Ask somebody else to write one.</br>
2. Send me 10+ BTC and I'll write one'.</br>
or</br>
3. Write one yourself and don't send a PR'.</br>
</br>
If you want your DuckyScript encrypted, please enter it elsewhere!</br>
				
			</div>
		</q-card>
	</q-modal>
</div>
`