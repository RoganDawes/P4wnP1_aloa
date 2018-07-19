package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/HuckRidgeSW/hvue"
	"strconv"
)

type CompHIDScriptData struct {
	*js.Object
	ScriptContent string `js:"scriptContent"`
}

func (data *CompHIDScriptData) SendAndRun(vm *hvue.VM) {
	md5 := StringToMD5(data.ScriptContent) //Calculate MD5 hexstring of current script content
	//js.Global.Call("alert", md5)

	go func() {
		timeout := uint32(0)
		err := UploadHIDScript(md5, data.ScriptContent)
		if err != nil { Alert("Error uploading script: " + err.Error()); return }
		job,err := RunHIDScript(md5, timeout)
		if err != nil { Alert("Error starting script as background job: " + err.Error()); return }
		Alert("Script started as background job: " + strconv.Itoa(int(job.Id)))
	}()
}

func newCompHIDScriptData(vm *hvue.VM) interface{} {
	newVM := &CompHIDScriptData{
		Object: js.Global.Get("Object").New(),
	}
	newVM.ScriptContent = "layout('us');\ntype('hello');"
	return newVM
}

func InitCompHIDScript() {
	hvue.NewComponent(
		"hid-script",
		hvue.Template(compHIDScriptTemplate),
		hvue.DataFunc(newCompHIDScriptData),
		hvue.MethodsOf(&CompHIDScriptData{}),
		)
}

const (

	compHIDScriptTemplate = `
<div>
	<span>P4wnP1 HID Script</span>
	<button @click="SendAndRun()">as Job</button><br>
	<keep-alive>
		<code-editor v-model="scriptContent"></code-editor>
	</keep-alive>
</div>
`
)

