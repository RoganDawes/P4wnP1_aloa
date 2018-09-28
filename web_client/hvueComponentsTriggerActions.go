// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/mame82/hvue"
)

func ExportDefaultTriggerActions() {
	// create test trigger

	// Trigger to run startup script
	triggerData := &jsTriggerServiceStarted{Object:O()}
	trigger := &jsTriggerAction_ServiceStarted{Object:O()}
	trigger.ServiceStarted = triggerData
	actionData := &jsActionStartBashScript{Object:O()}
	actionData.ScriptPath = "/usr/local/P4wnP1/scripts/servicestart.sh"
	action := &jsTriggerAction_BashScript{Object:O()}
	action.BashScript = actionData
	svcUpRunScript := &jsTriggerAction{Object:O()}
	svcUpRunScript.OneShot = false
	svcUpRunScript.Id = 0
	svcUpRunScript.Trigger = trigger.Object
	svcUpRunScript.Action = action.Object

	js.Global.Set("testtriggeraction", svcUpRunScript)

	// Try to cast back (shouldn't work because of the interfaces
	copyobj := &jsTriggerAction{Object:js.Global.Get("testtriggeraction")}
	js.Global.Set("copyobj", copyobj)
	println("copyobj", copyobj)
	println("copyobjtrigger", copyobj.Trigger) //<--- this wouldn't work

	if isJsTriggerAction_ServiceStarted(copyobj.Trigger) {
		println("is service started trigger")
	}
	if isJsTriggerAction_UsbGadgetConnected(copyobj.Trigger) {
		println("is USB gadget connected trigger")
	}
	if isJsTriggerAction_BashScript(copyobj.Action) {
		println("is BashScript action")
	}
	if isJsTriggerAction_HidScript(copyobj.Trigger) {
		println("is HIDScript action")
	}


	/*
	serviceUpRunScript := &pb.TriggerAction{
		Trigger: &pb.TriggerAction_ServiceStarted{
			ServiceStarted: &pb.TriggerServiceStarted{},
		},
		Action: &pb.TriggerAction_BashScript{
			BashScript: &pb.ActionStartBashScript{
				ScriptPath: "/usr/local/P4wnP1/scripts/servicestart.sh", // ToDo: use real script path once ready
			},
		},
	}
	a[0] = serviceUpRunScript

	logServiceStart := &pb.TriggerAction{
		Trigger: &pb.TriggerAction_ServiceStarted{
			ServiceStarted: &pb.TriggerServiceStarted{},
		},
		Action: &pb.TriggerAction_Log{
			Log: &pb.ActionLog{},
		},
	}
	a[1]= logServiceStart

	logDHCPLease := &pb.TriggerAction{
		Trigger: &pb.TriggerAction_DhcpLeaseGranted{
			DhcpLeaseGranted: &pb.TriggerDHCPLeaseGranted{},
		},
		Action: &pb.TriggerAction_Log{
			Log: &pb.ActionLog{},
		},
	}
	a[2] = logDHCPLease

	logUSBGadgetConnected := &pb.TriggerAction{
		Trigger: &pb.TriggerAction_UsbGadgetConnected{
			UsbGadgetConnected: &pb.TriggerUSBGadgetConnected{},
		},
		Action: &pb.TriggerAction_Log{
			Log: &pb.ActionLog{},
		},
	}
	a[3] = logUSBGadgetConnected

	logUSBGadgetDisconnected := &pb.TriggerAction{
		Trigger: &pb.TriggerAction_UsbGadgetDisconnected{
			UsbGadgetDisconnected: &pb.TriggerUSBGadgetDisconnected{},
		},
		Action: &pb.TriggerAction_Log{
			Log: &pb.ActionLog{},
		},
	}
	a[4] = logUSBGadgetDisconnected

	logWifiAp := &pb.TriggerAction{
		Trigger: &pb.TriggerAction_WifiAPStarted{
			WifiAPStarted: &pb.TriggerWifiAPStarted{},
		},
		Action: &pb.TriggerAction_Log{
			Log: &pb.ActionLog{},
		},
	}
	a[5] = logWifiAp

	logWifiSta := &pb.TriggerAction{
		Trigger: &pb.TriggerAction_WifiConnectedAsSta{
			WifiConnectedAsSta: &pb.TriggerWifiConnectedAsSta{},
		},
		Action: &pb.TriggerAction_Log{
			Log: &pb.ActionLog{},
		},
	}
	a[6] = logWifiSta

	logSSHLogin := &pb.TriggerAction{
		Trigger: &pb.TriggerAction_SshLogin{
			SshLogin: &pb.TriggerSSHLogin{},
		},
		Action: &pb.TriggerAction_Log{
			Log: &pb.ActionLog{},
		},
	}
	a[7] = logSSHLogin
	*/

}

type jsIsTriggerAction_Trigger interface {
	isTriggerAction_Trigger()
}
type jsIsTriggerAction_Action interface{ isTriggerAction_Action() }

type jsTriggerAction struct {
	*js.Object

	Id      uint32 `js:"Id"`
	OneShot bool `js:"OnShot"`

	Trigger *js.Object `js:"Trigger"`
	Action *js.Object `js:"Action"`
}

// TriggerAction_ServiceStarted is assignable to Trigger
type jsTriggerAction_ServiceStarted struct {
	*js.Object
	ServiceStarted *jsTriggerServiceStarted `js:"ServiceStarted"`
}

func isJsTriggerAction_ServiceStarted(src *js.Object) bool {
	test := jsTriggerAction_ServiceStarted{Object:src}
	if test.ServiceStarted.Object == js.Undefined { return false }
	return true
}

// TriggerAction_UsbGadgetConnected is assignable to Trigger
type jsTriggerAction_UsbGadgetConnected struct {
	*js.Object
	UsbGadgetConnected *jsTriggerUSBGadgetConnected `js:"UsbGadgetConnected"`
}

func isJsTriggerAction_UsbGadgetConnected(src *js.Object) bool {
	test := jsTriggerAction_UsbGadgetConnected{Object:src}
	if test.UsbGadgetConnected.Object == js.Undefined { return false }
	return true
}

// TriggerAction_UsbGadgetDisconnected is assignable to Trigger
type jsTriggerAction_UsbGadgetDisconnected struct {
	*js.Object
	UsbGadgetDisconnected *jsTriggerUSBGadgetDisconnected `js:"UsbGadgetDisconnected"`
}
func isJsTriggerAction_UsbGadgetDisconnected(src *js.Object) bool {
	test := jsTriggerAction_UsbGadgetDisconnected{Object:src}
	if test.UsbGadgetDisconnected.Object == js.Undefined { return false }
	return true
}

// TriggerAction_WifiAPStarted is assignable to Trigger
type jsTriggerAction_WifiAPStarted struct {
	*js.Object
	WifiAPStarted *jsTriggerWifiAPStarted `js:"WifiAPStarted"`
}
func iJsTriggerAction_WifiAPStarted(src *js.Object) bool {
	test := jsTriggerAction_WifiAPStarted{Object:src}
	if test.WifiAPStarted.Object == js.Undefined { return false }
	return true
}

// TriggerAction_WifiConnectedAsSta is assignable to Trigger
type jsTriggerAction_WifiConnectedAsSta struct {
	*js.Object
	WifiConnectedAsSta *jsTriggerWifiConnectedAsSta `js:"WifiConnectedAsSta"`
}
func isJsTriggerAction_WifiConnectedAsSta(src *js.Object) bool {
	test := jsTriggerAction_WifiConnectedAsSta{Object:src}
	if test.WifiConnectedAsSta.Object == js.Undefined { return false }
	return true
}

// TriggerAction_SshLogin is assignable to Trigger
type jsTriggerAction_SshLogin struct {
	*js.Object
	SshLogin *jsTriggerSSHLogin `js:"SshLogin"`
}
func isJsTriggerAction_SshLogin(src *js.Object) bool {
	test := jsTriggerAction_SshLogin{Object:src}
	if test.SshLogin.Object == js.Undefined { return false }
	return true
}

// TriggerAction_DhcpLeaseGranted is assignable to Trigger
type jsTriggerAction_DhcpLeaseGranted struct {
	*js.Object
	DhcpLeaseGranted *jsTriggerDHCPLeaseGranted `js:"DhcpLeaseGranted"`
}
func isJsTriggerAction_DhcpLeaseGranted(src *js.Object) bool {
	test := jsTriggerAction_DhcpLeaseGranted{Object:src}
	if test.DhcpLeaseGranted.Object == js.Undefined { return false }
	return true
}

// TriggerAction_BashScript is assignable to Action
type jsTriggerAction_BashScript struct {
	*js.Object
	BashScript *jsActionStartBashScript `js:"BashScript"`
}
func isJsTriggerAction_BashScript(src *js.Object) bool {
	test := jsTriggerAction_BashScript{Object:src}
	if test.BashScript.Object == js.Undefined { return false }
	return true
}

// TriggerAction_HidScript is assignable to Action
type jsTriggerAction_HidScript struct {
	*js.Object
	HidScript *jsActionStartHIDScript `js:"HidScript"`
}
func isJsTriggerAction_HidScript(src *js.Object) bool {
	test := jsTriggerAction_HidScript{Object:src}
	if test.HidScript.Object == js.Undefined { return false }
	return true
}

// TriggerAction_DeploySettingsTemplate is assignable to Action
type jsTriggerAction_DeploySettingsTemplate struct {
	*js.Object
	DeploySettingsTemplate *jsActionDeploySettingsTemplate `js:"DeploySettingsTemplate"`
}
func isJsTriggerAction_DeploySettingsTemplate(src *js.Object) bool {
	test := jsTriggerAction_DeploySettingsTemplate{Object:src}
	if test.DeploySettingsTemplate.Object == js.Undefined { return false }
	return true
}

// TriggerAction_Log is assignable to Action
type jsTriggerAction_Log struct {
	*js.Object
	Log *jsActionLog `js:"Log"`
}
func isJsTriggerAction_Log(src *js.Object) bool {
	test := jsTriggerAction_Log{Object:src}
	if test.Log.Object == js.Undefined { return false }
	return true
}

func (*jsTriggerAction_ServiceStarted) isTriggerAction_Trigger()        {}
func (*jsTriggerAction_UsbGadgetConnected) isTriggerAction_Trigger()    {}
func (*jsTriggerAction_UsbGadgetDisconnected) isTriggerAction_Trigger() {}
func (*jsTriggerAction_WifiAPStarted) isTriggerAction_Trigger()         {}
func (*jsTriggerAction_WifiConnectedAsSta) isTriggerAction_Trigger()    {}
func (*jsTriggerAction_SshLogin) isTriggerAction_Trigger()              {}
func (*jsTriggerAction_DhcpLeaseGranted) isTriggerAction_Trigger()      {}
func (*jsTriggerAction_BashScript) isTriggerAction_Action()             {}
func (*jsTriggerAction_HidScript) isTriggerAction_Action()              {}
func (*jsTriggerAction_DeploySettingsTemplate) isTriggerAction_Action() {}
func (*jsTriggerAction_Log) isTriggerAction_Action()                    {}

type jsTriggerServiceStarted struct {
	*js.Object
}

type jsTriggerUSBGadgetConnected struct {
	*js.Object
}

type jsTriggerUSBGadgetDisconnected struct {
	*js.Object
}

type jsTriggerWifiAPStarted struct {
	*js.Object
}
type jsTriggerWifiConnectedAsSta struct {
	*js.Object
}
type jsTriggerSSHLogin struct {
	*js.Object
	ResLoginUser string `js:"ResLoginUser"`
}
type jsTriggerDHCPLeaseGranted struct {
	*js.Object
	ResInterface string `js:"ResInterface"`
	ResClientIP  string `js:"ResClientIP"`
	ResClientMac string `js:"ResClientMac"`
}

type jsActionStartBashScript struct {
	*js.Object
	ScriptPath string `js:"ScriptPath"`
}
type jsActionStartHIDScript struct {
	*js.Object
	ScriptName string `js:"ScriptName"`
}
type jsActionDeploySettingsTemplate struct {
	*js.Object
	TemplateName string `js:"TemplateName"`
	Type         string `js:"Type"`
}
type jsActionLog struct {
	*js.Object
}



func InitComponentsTriggerActions() {
	// ToDo: delete test
	ExportDefaultTriggerActions()

	hvue.NewComponent(
		"triggeraction",
		hvue.Template(templateTriggerAction),
		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := struct {
				*js.Object
				ShowStoreModal bool   `js:"showStoreModal"`
				ShowLoadModal bool   `js:"showLoadModal"`
				TemplateName   string `js:"templateName"`
			}{Object: O()}
			data.ShowStoreModal = false
			data.ShowLoadModal = false
			data.TemplateName = ""
			return &data
		}),
		hvue.Computed("ta", func(vm *hvue.VM) interface{} {
			return js.Global.Get("testtriggeraction")
		}),
		hvue.Computed("triggertypes", func(vm *hvue.VM) interface{} {
			tts := js.Global.Get("Array").New()
			type entry struct {
				*js.Object
				Label string `js:"label"`
				Value *js.Object `js:"value"`
			}

			//trigger service started
			entrySvcSt := entry{Object:O()}
			entrySvcSt.Label = "Service started"
			trigger := &jsTriggerAction_ServiceStarted{Object:O()}
			trigger.ServiceStarted = &jsTriggerServiceStarted{Object:O()}
			entrySvcSt.Value = trigger.Object
			tts.Call("push", entrySvcSt)

			return tts
		}),
	)
}

const templateTriggerAction = `
<q-page padding>
<div class="row gutter-sm">
Hello TA
{{ ta }} <br />
{{ triggertypes }}
</div>
</q-page>	

`
