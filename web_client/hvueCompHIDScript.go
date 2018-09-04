// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/HuckRidgeSW/hvue"
	"strconv"
)

type CompHIDScriptCodeEditorData struct {
	*js.Object
	CodeMirrorOptions *CodeMirrorOptionsType `js:"codemirrorOptions"`
}

func (data *CompHIDScriptCodeEditorData) SendAndRun(vm *hvue.VM) {
	sourceCode := vm.Get("scriptContent").String()

	md5 := StringToMD5(sourceCode) //Calculate MD5 hexstring of current script content

	go func() {
		timeout := uint32(0)
		err := UploadHIDScript(md5, sourceCode)
		if err != nil {
			QuasarNotifyError("Error uploading script", err.Error(), QUASAR_NOTIFICATION_POSITION_TOP)
			return
		}
		job,err := RunHIDScript(md5, timeout)
		if err != nil {
			QuasarNotifyError("Error starting script as background job", err.Error(), QUASAR_NOTIFICATION_POSITION_TOP)
			return
		}

		QuasarNotifySuccess("Script started successfully", "Job ID " + strconv.Itoa(int(job.Id)), QUASAR_NOTIFICATION_POSITION_TOP)
	}()
}

type CodeMirrorMode struct {
	*js.Object
	Name string `js:"name"`
	GlobalVars bool `js:"globalVars"`
}

type CodeMirrorExtraKeys struct {
	*js.Object
	CtrlSpace string `js:"Ctrl-Space"`
}

type CodeMirrorOptionsType struct {
	*js.Object
	Mode *CodeMirrorMode `js:"mode"`
	LineNumbers bool `js:"lineNumbers"`
	LineWrapping bool `js:"lineWrapping"`
	AutoCloseBrackets bool `js:"autoCloseBrackets"`
	ExtraKeys *CodeMirrorExtraKeys `js:"extraKeys"`
}

func newCompHIDScriptCodeEditorData(vm *hvue.VM) interface{} {
	data := &CompHIDScriptCodeEditorData{ Object: O() }

	data.CodeMirrorOptions = &CodeMirrorOptionsType{Object: O()}

	data.CodeMirrorOptions.Mode = &CodeMirrorMode{ Object: O() }
	data.CodeMirrorOptions.Mode.Name = "text/javascript"
	data.CodeMirrorOptions.Mode.GlobalVars = true //expose globalVars of mode for auto-complete with addon/hint/show-hint.js, addon/hint/javascript-hint.js"

	data.CodeMirrorOptions.ExtraKeys = &CodeMirrorExtraKeys{ Object: O() }
	data.CodeMirrorOptions.ExtraKeys.CtrlSpace = "autocomplete"

	data.CodeMirrorOptions.LineNumbers = true
	//data.CodeMirrorOptions.LineWrapping = true
	data.CodeMirrorOptions.AutoCloseBrackets = true

	return data
}


func InitComponentsHIDScript() {
	hvue.NewComponent(
		"hid-script-code-editor",
		hvue.Template(compHIDScriptCodeEditorTemplate),
		hvue.DataFunc(newCompHIDScriptCodeEditorData),
		hvue.MethodsOf(&CompHIDScriptCodeEditorData{}),
		hvue.ComputedWithGetSet(
			"scriptContent",
			func(vm *hvue.VM) interface{} {
				return vm.Get("$store").Get("state").Get("currentHIDScriptSource")
			},
			func(vm *hvue.VM, newScriptContent *js.Object) {
				vm.Get("$store").Call("commit", VUEX_MUTATION_SET_CURRENT_HID_SCRIPT_SOURCE_TO, newScriptContent)
			}),
		)

	hvue.NewComponent(
		"hid-script",
		hvue.Template(compHIDScriptTemplate),
	)
}

const (

	compHIDScriptTemplate = `
<q-page>
<div class="row content-stretch">
	<div class="col-10 self-stretch">
		<hid-script-code-editor></hid-script-code-editor>
	</div>
	<div class="col-2">
		<hid-job-overview></hid-job-overview>
	</div>
</div>
<div class="row content-stretch">
	<hid-job-event-overview></hid-job-event-overview>
</div>


</q-page>
`
	compHIDScriptCodeEditorTemplate = `
	<q-card class="q-ma-sm">
  		<q-card-title>
    		HIDScript editor
  		</q-card-title>

		<q-card-separator />

		<q-card-actions>
    		<q-btn color="primary" @click="SendAndRun()">run</q-btn>
		</q-card-actions>

		<q-card-separator />

		<q-card-main>
			<codemirror v-model="scriptContent" :options="codemirrorOptions"></codemirror>
	  	</q-card-main>
	
	</q-card>
`
)

