// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/HuckRidgeSW/hvue"
	"strconv"
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
		if err != nil {
			notification := &QuasarNotification{Object: O()}
			notification.Message = "Error uploading script"
			notification.Detail = err.Error()
			notification.Position = QUASAR_NOTIFICATION_POSITION_TOP
			notification.Type = QUASAR_NOTIFICATION_TYPE_NEGATIVE
			notification.Timeout = 5000
			QuasarNotify(notification)
			return
		}
		job,err := RunHIDScript(md5, timeout)
		if err != nil {
			notification := &QuasarNotification{Object: O()}
			notification.Message = "Error starting script as background job"
			notification.Detail = err.Error()
			notification.Position = QUASAR_NOTIFICATION_POSITION_TOP
			notification.Type = QUASAR_NOTIFICATION_TYPE_NEGATIVE
			notification.Timeout = 5000
			QuasarNotify(notification)
			return
		}

		notification := &QuasarNotification{Object: O()}
		notification.Message = "Script started successfully"
		notification.Detail = "Job ID " + strconv.Itoa(int(job.Id))
		notification.Position = QUASAR_NOTIFICATION_POSITION_TOP
		notification.Type = QUASAR_NOTIFICATION_TYPE_POSITIVE
		notification.Timeout = 5000
		QuasarNotify(notification)
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
				return vm.Get("$store").Get("state").Get("currentHIDScriptSource")
			},
			func(vm *hvue.VM, newScriptContent *js.Object) {
				vm.Get("$store").Call("commit", VUEX_MUTATION_SET_CURRENT_HID_SCRIPT_SOURCE_TO, newScriptContent)
			}),
		)
}

const (

	compHIDScriptTemplate = `
<q-page class="row item-start">

	
	<q-card class="q-ma-sm" :inline="$q.platform.is.desktop">
  		<q-card-title>
    		HIDScript editor
  		</q-card-title>

		<q-card-separator />

		<q-card-actions>
    		<q-btn color="primary" @click="SendAndRun()">run</q-btn>
		</q-card-actions>

		<q-card-separator />

		<q-card-main>
	    	<code-editor v-model="scriptContent"></code-editor>
	  	</q-card-main>
	</q-card>


	<hid-job-overview></hid-job-overview>


<q-page>
`
)

