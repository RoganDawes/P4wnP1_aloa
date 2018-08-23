// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/HuckRidgeSW/hvue"
)

//https://github.com/cnu4/vue-codemirror-lite/blob/master/codemirror.vue

type CodeMirrorOptionsType struct {
	*js.Object
	Mode interface{} `js:"mode"`
	LineNumbers bool `js:"lineNumbers"`
	LineWrapping bool `js:"lineWrapping"`
	AutoCloseBrackets bool `js:"autoCloseBrackets"`
	ExtraKeys interface{} `js:"extraKeys"`
}

type CompCodeEditorData struct {
	*js.Object
	ScriptContent string `js:"scriptContent"`
	CodeMirrorOptions *CodeMirrorOptionsType `js:"codemirrorOptions"`

}

func NewCodeEditorData(vm *hvue.VM) interface{} {
	data := &CompCodeEditorData{ Object: O() }
	data.ScriptContent = ""
	cmo := &CodeMirrorOptionsType{Object: O()}

	mode := struct{
		*js.Object
		Name string `js:"name"`
		GlobalVars bool `js:"globalVars"`
	}{ Object: O() }
	mode.Name = "text/javascript"
	mode.GlobalVars = true //expose globalVars of mode for auto-complete with addon/hint/show-hint.js, addon/hint/javascript-hint.js"
	cmo.Mode = &mode


	extraKeys := struct{
		*js.Object
		CtrlSpace string `js:"Ctrl-Space"`
	}{ Object: O() }
	extraKeys.CtrlSpace = "autocomplete"
	cmo.ExtraKeys = &extraKeys



	cmo.LineNumbers = true
	cmo.LineWrapping = true
	cmo.AutoCloseBrackets = true
	data.CodeMirrorOptions = cmo
	return data
}



func initCodeMirror(vm *hvue.VM) {
	// this.value = this.scriptContent (copy "value" property over to "scriptContent")
	val := vm.Get("value")
	vm.Set("scriptContent", val)


	//this.editor = CodeMirror.fromTextArea(this.$el.querySelector('#CodeEditor'), this.codemirrorOptions)
	editorEl := vm.El.Call("querySelector","#CodeEditor")
	editor := js.Global.Get("CodeMirror").Call("fromTextArea", editorEl, vm.Get("codemirrorOptions"))

	//copy value property to initial editor state
	editor.Call("setValue", val)

	editor.Call("on", "change", func(cm *js.Object) {
		newVal := cm.Call("getValue")

		//update ViewModel data scriptContent
		vm.Set("scriptContent", newVal)
		//propagate up change
		vm.Emit("change", newVal)
		vm.Emit("input", newVal)
	})

}

func InitCompCodeEditor() {
	hvue.NewComponent(
		"code-editor",
		hvue.Template(compCodeEditorTemplate),
		hvue.DataFunc(NewCodeEditorData),
		hvue.PropObj("value",hvue.Types(hvue.PString),hvue.Default("type('Hello');")),
		hvue.Mounted(initCodeMirror),

	)
}

const(
	compCodeEditorTemplate = `
<div>
	<textarea id="CodeEditor"></textarea>
</div>
`
)


