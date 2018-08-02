// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/mame82/hvue"
	"strconv"
	"google.golang.org/grpc/status"
)

type CompHIDScriptData struct {
	*js.Object
	//ScriptContent string `js:"scriptContent"`
}

func (data *CompHIDScriptData) SendAndRun(vm *hvue.VM) {
	sourceCode := vm.Get("scriptContent").String()

	md5 := StringToMD5(sourceCode) //Calculate MD5 hexstring of current script content
	//js.Global.Call("alert", md5)

	go func() {
		timeout := uint32(0)
		err := UploadHIDScript(md5, sourceCode)
		if err != nil { Alert("Error uploading script: " + err.Error()); return }
		job,err := RunHIDScript(md5, timeout)
		if err != nil {
			println(status.Convert(err))
			Alert("Error starting script as background job: " + err.Error())
			return
		}
		Alert("Script started as background job: " + strconv.Itoa(int(job.Id)))
	}()
}

func newCompHIDScriptData(vm *hvue.VM) interface{} {
	newVM := &CompHIDScriptData{
		Object: js.Global.Get("Object").New(),
	}
	//newVM.ScriptContent = "layout('us');\ntype('hello');"
	return newVM
}

func InitCompHIDScript() {
	hvue.NewComponent(
		"hid-script",
		hvue.Template(compHIDScriptTemplate),
		hvue.DataFunc(newCompHIDScriptData),
		hvue.MethodsOf(&CompHIDScriptData{}),
		hvue.ComputedWithGetSet(
			"scriptContent",
			func(vm *hvue.VM) interface{} {
				return vm.Store.Get("state").Get("currentHIDScriptSource")
			},
			func(vm *hvue.VM, newScriptContent *js.Object) {
				vm.Store.Call("commit", VUEX_MUTATION_SET_CURRENT_HID_SCRIPT_SOURCE_TO, newScriptContent)
			}),
		)
}

const (

	compHIDScriptTemplate = `
<div>
	<span>P4wnP1 HID Script</span>
	<button @click="SendAndRun()">as Job</button><br>
	<code-editor v-model="scriptContent"></code-editor>
	<hidjobs></hidjobs>
</div>
`
)

