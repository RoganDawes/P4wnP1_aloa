// +build linux

package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/mame82/P4wnP1_aloa/common"
	"github.com/mame82/P4wnP1_aloa/common_web"
	pb "github.com/mame82/P4wnP1_aloa/proto"
	"github.com/mame82/P4wnP1_aloa/service/util"
	"io/ioutil"
	"sync"
)

var (
	ErrTaNotFound = errors.New("Couldn't find given TriggerAction")
	ErrTaImmutable = errors.New("Not allowed to change immutable TriggerAction")
)


type triggerType int
const (
	triggerTypeServiceStarted triggerType = iota
	triggerTypeUsbGadgetConnected
	triggerTypeUsbGadgetDisconnected
	triggerTypeWifiAPStarted
	triggerTypeWifiConnectedAsSta
	triggerTypeSshLogin
	triggerTypeDhcpLeaseGranted
	triggerTypeGroupReceive
	triggerTypeGroupReceiveMulti
	triggerTypeGpioIn
)
var triggerTypeString = map[triggerType]string {
	triggerTypeServiceStarted:        "TRIGGER_SERVICE_STARTED",
	triggerTypeUsbGadgetConnected:    "TRIGGER_USB_GADGET_CONNECTED",
	triggerTypeUsbGadgetDisconnected: "TRIGGER_USB_GADGET_DISCONNECTED",
	triggerTypeWifiAPStarted:         "TRIGGER_WIFI_AP_STARTED",
	triggerTypeWifiConnectedAsSta:    "TRIGGER_WIFI_CONNECTED_AS_STA",
	triggerTypeSshLogin:              "TRIGGER_SSH_LOGIN",
	triggerTypeDhcpLeaseGranted:      "TRIGGER_DHCP_LEASE_GRANTED",
	triggerTypeGroupReceive:          "TRIGGER_GROUP_RECEIVE",
	triggerTypeGroupReceiveMulti:     "TRIGGER_GROUP_RECEIVE_MULTI",
	triggerTypeGpioIn:                "TRIGGER_GPIO_IN",
}

type actionType int
const (
	actionTypeBashScript actionType = iota
	actionTypeHidScript
	actionTypeDeploySettingsTemplate
	actionTypeLog
	actionTypeGpioOut
	actionTypeGroupSend
)
var actionTypeString = map[actionType]string {
	actionTypeBashScript: "ACTION_BASH_SCRIPT",
	actionTypeHidScript: "ACTION_HID_SCRIPT",
	actionTypeDeploySettingsTemplate: "ACTION_DEPLOY_SETTINGS_TEMPLATE",
	actionTypeLog: "ACTION_LOG",
	actionTypeGpioOut: "ACTION_GPIO_OUT",
	actionTypeGroupSend: "ACTION_GROUP_SEND",

}


func retrieveTriggerActionTypes(ta *pb.TriggerAction) (ttype triggerType, atype actionType) {
	// Trigger
	switch x := ta.Trigger.(type) {
	case *pb.TriggerAction_ServiceStarted:
		ttype = triggerTypeServiceStarted
	case *pb.TriggerAction_UsbGadgetConnected:
		ttype = triggerTypeUsbGadgetConnected
	case *pb.TriggerAction_UsbGadgetDisconnected:
		ttype = triggerTypeUsbGadgetDisconnected
	case *pb.TriggerAction_WifiAPStarted:
		ttype = triggerTypeWifiAPStarted
	case *pb.TriggerAction_WifiConnectedAsSta:
		ttype = triggerTypeWifiConnectedAsSta
	case *pb.TriggerAction_SshLogin:
		ttype = triggerTypeSshLogin
	case *pb.TriggerAction_DhcpLeaseGranted:
		ttype = triggerTypeDhcpLeaseGranted
	case *pb.TriggerAction_GroupReceive:
		ttype = triggerTypeGroupReceive
	case *pb.TriggerAction_GroupReceiveMulti:
		ttype = triggerTypeGroupReceiveMulti
	case *pb.TriggerAction_GpioIn:
		ttype = triggerTypeGpioIn
	case nil:
	default:
		panic(fmt.Sprintf("unexpected trigger type %T", x))
	}
	// Action
	switch x := ta.Action.(type) {
	case *pb.TriggerAction_BashScript:
		atype = actionTypeBashScript
	case *pb.TriggerAction_HidScript:
		atype = actionTypeHidScript
	case *pb.TriggerAction_DeploySettingsTemplate:
		atype = actionTypeDeploySettingsTemplate
	case *pb.TriggerAction_Log:
		atype = actionTypeLog
	case *pb.TriggerAction_GpioOut:
		atype = actionTypeGpioOut
	case *pb.TriggerAction_GroupSend:
		atype = actionTypeGroupSend
	case nil:
	default:
		panic(fmt.Sprintf("unexpected action type %T", x))
	}

	return
}

type TriggerActionManager struct {
	rootSvc *Service
	evtRcv *EventReceiver

	registeredTriggerActionMutex *sync.Mutex
	//registeredTriggerAction      []*pb.TriggerAction
	registeredTriggerActions      pb.TriggerActionSet

	groupReceiveSequenceCheckers map[*pb.TriggerAction]*util.ValueSequenceChecker
	groupReceiveSequenceCheckersMutex *sync.Mutex

	nextID                       uint32
}

func (tam *TriggerActionManager) processing_loop() {

	fmt.Println("TAM processing loop started")
	Outer:
		for {
			select {
			case evt := <- tam.evtRcv.EventQueue:
				// avoid consuming empty messages, because channel is closed
				if evt == nil {
					break Outer // abort loop on "nil" event, as this indicates the EventQueue channel has been closed
				}
				//fmt.Println("TriggerActionManager received unfiltered event", evt)
				tam.processTriggerEvent(evt)
				// check if relevant and dispatch to triggers
			case <- tam.evtRcv.Ctx.Done():
				// evvent Receiver cancelled or unregistered
				break Outer
			}
		}
	fmt.Println("TAM processing loop finished")
}

// iterates over registered trigger actions
// if event matches a trigger, pass execution to respective on{Event} method, along with the arguments
// from the respective events
//
// Tasks of on{event} method:
// 	- decide if the given trigger fires, based on the event arguments
//  - disable the TriggerAction, in case the trigger has fired
//  - call the execute{Action} method, according to the Action defined in the trigger Action, in case the trigger fires
//
// Note: a event doesn't necessarily map to a trigger (f.e. a TRIGGER_EVT_TYPE_GROUP_RECEIVE carries a single value to a
// group, but a triggerTypeGroupReceive doesn't trigger if it is the wrong value)
//
func (tam *TriggerActionManager) processTriggerEvent(evt *pb.Event) {
	//fmt.Printf("Remaining triggerActions: %+v\n", tam.registeredTriggerAction)
	//fmt.Printf("TriggerActionManager Received event: %+v\n", evt)
	tam.registeredTriggerActionMutex.Lock()
	defer tam.registeredTriggerActionMutex.Unlock()

	for _,ta := range tam.registeredTriggerActions.TriggerActions {
		// skip disabled triggeractions
		if !ta.IsActive { continue }
		ttype,atype := retrieveTriggerActionTypes(ta)
		if taTriggerTypeMatchesEvtTriggerType(ttype, evt) {
			switch ttEvt := common_web.EvtTriggerType(evt.Values[0].GetTint64()); ttEvt {
			case common_web.TRIGGER_EVT_TYPE_SERVICE_STARTED:
				tam.onServiceStarted(evt, ta, ttype, atype)
			case common_web.TRIGGER_EVT_TYPE_SSH_LOGIN:
				tam.onSSHLogin(evt, ta, ttype, atype)
			case common_web.TRIGGER_EVT_TYPE_WIFI_CONNECTED_AS_STA:
				tam.onWifiConnectedAsSta(evt, ta, ttype, atype)
			case common_web.TRIGGER_EVT_TYPE_WIFI_AP_STARTED:
				tam.onWifiApStarted(evt, ta, ttype, atype)
			case common_web.TRIGGER_EVT_TYPE_DHCP_LEASE_GRANTED:
				// extract iface, mac and ip from event
				tam.onDhcpLeaseGranted(evt, ta, ttype, atype)
			case common_web.TRIGGER_EVT_TYPE_USB_GADGET_CONNECTED:
				tam.onUsbGadgetConnected(evt, ta, ttype, atype)
			case common_web.TRIGGER_EVT_TYPE_USB_GADGET_DISCONNECTED:
				tam.onUsbGadgetDisconnected(evt, ta, ttype, atype)
			case common_web.TRIGGER_EVT_TYPE_GPIO_IN:
				tam.onGpioIn(evt, ta, ttype, atype)
			case common_web.TRIGGER_EVT_TYPE_GROUP_RECEIVE:
				tam.onGroupReceive(evt, ta, ttype, atype)
			default:
				fmt.Println("unhandled trigger: ", ttEvt)
			}

		}
	}

	return
}

func (tam *TriggerActionManager) onServiceStarted(evt *pb.Event, ta *pb.TriggerAction, tt triggerType, at actionType) error {
	// always triggers
	tam.executeAction(evt, ta, tt, at)
	return nil
}


func (tam *TriggerActionManager) onSSHLogin(evt *pb.Event, ta *pb.TriggerAction, tt triggerType, at actionType) error {
	//ToDo: allow filtering by login user
	// always triggers
	tam.executeAction(evt, ta, tt, at)
	return nil
}

func (tam *TriggerActionManager) onWifiConnectedAsSta(evt *pb.Event, ta *pb.TriggerAction, tt triggerType, at actionType) error {
	//ToDo: filter by AP name
	// always triggers
	tam.executeAction(evt, ta, tt, at)
	return nil
}

func (tam *TriggerActionManager) onWifiApStarted(evt *pb.Event, ta *pb.TriggerAction, tt triggerType, at actionType) error {
	//ToDo: provide AP name with event and hand it over to the action
	// always triggers
	tam.executeAction(evt, ta, tt, at)
	return nil
}

func (tam *TriggerActionManager) onDhcpLeaseGranted(evt *pb.Event, ta *pb.TriggerAction, tt triggerType, at actionType) error {
	//ToDo: filter by source interface, mac, IP
	// always triggers
	tam.executeAction(evt, ta, tt, at)
	return nil
}

func (tam *TriggerActionManager) onUsbGadgetConnected(evt *pb.Event, ta *pb.TriggerAction, tt triggerType, at actionType) error {
	// always triggers
	tam.executeAction(evt, ta, tt, at)
	return nil
}

func (tam *TriggerActionManager) onUsbGadgetDisconnected(evt *pb.Event, ta *pb.TriggerAction, tt triggerType, at actionType) error {
	// always triggers
	tam.executeAction(evt, ta, tt, at)
	return nil
}

func (tam *TriggerActionManager) onGpioIn(evt *pb.Event, ta *pb.TriggerAction, tt triggerType, at actionType) error {
	evtGpioName, evtGpioLevel,err := DeconstructEventTriggerGpioIn(evt)
	if err != nil { return err }

	taGpioName := ta.Trigger.(*pb.TriggerAction_GpioIn).GpioIn.GpioName

	if taGpioName != evtGpioName {
		return nil // ignore
	}

	tam.executeAction(evt, ta, tt, at)
	fmt.Printf("Gpio in trigger '%s' new state: %v\n", evtGpioName, evtGpioLevel)

	return nil
}

func (tam *TriggerActionManager) onGroupReceive(evt *pb.Event, ta *pb.TriggerAction, tt triggerType, at actionType) error {
	evGroupName,evValue,err := DeconstructEventTriggerGroupReceive(evt)
	if err != nil { return err }

	switch tt {
	case triggerTypeGroupReceive:
		triggerVal := ta.Trigger.(*pb.TriggerAction_GroupReceive).GroupReceive.Value
		triggerGroupName := ta.Trigger.(*pb.TriggerAction_GroupReceive).GroupReceive.GroupName
		if evGroupName != triggerGroupName {
			return nil // don't handle on group mismatch, but return without error
		}
		if evValue != triggerVal {
			return nil // don't handle on value mismatch, but return without error
		}
		tam.executeAction(evt, ta, tt, at) // fire action
		return nil
	case triggerTypeGroupReceiveMulti:
	//	fmt.Println("### Processing GroupReceive event for trigger type GroupReceiveSequence")
		triggerGroupName := ta.Trigger.(*pb.TriggerAction_GroupReceiveMulti).GroupReceiveMulti.GroupName
		if evGroupName != triggerGroupName {
			return nil
		}
		// retrieve the sequence checker
		if sc,exists := tam.groupReceiveSequenceCheckers[ta]; exists {
			if sc.Check(evValue) {
				tam.executeAction(evt, ta, tt, at) // fire action
			}
		//	fmt.Printf("GrpRcvSeq '%s' received '%d': %s\n", triggerGroupName, evValue, sc)
		}

		return nil // don't handle on group mismatch, but return without error


	default:
		return errors.New("Wrong trigger for onGroupReceive event")
	}

	return nil
}



func (tam *TriggerActionManager) executeAction(evt *pb.Event, ta *pb.TriggerAction, tt triggerType, at actionType) error {
	if ta.OneShot { ta.IsActive = false }

	switch actionType := ta.Action.(type) {
	case *pb.TriggerAction_BashScript:
		go tam.executeActionBashScript(evt, ta, tt, at, actionType.BashScript)
	case *pb.TriggerAction_HidScript:
		go tam.executeActionStartHidScript(evt, ta, tt, at, actionType.HidScript)
	case *pb.TriggerAction_Log:
		go tam.executeActionLog(evt, ta, tt, at, actionType.Log)
	case *pb.TriggerAction_DeploySettingsTemplate:
		go tam.executeActionDeploySettingsTemplate(evt, ta, tt, at, actionType.DeploySettingsTemplate)
	case *pb.TriggerAction_GroupSend:
		tam.executeActionGroupSend(evt, ta, tt, at, actionType.GroupSend)
	case *pb.TriggerAction_GpioOut:
		tam.executeActionGPIOOut(evt, ta, tt, at, actionType.GpioOut)
	}

	return nil
}


func (tam *TriggerActionManager) executeActionDeploySettingsTemplate(evt *pb.Event, ta *pb.TriggerAction, tt triggerType, at actionType, action *pb.ActionDeploySettingsTemplate) {
	triggerName := triggerTypeString[tt]
	actionName := actionTypeString[at]

	templateTypeName := pb.ActionDeploySettingsTemplate_TemplateType_name[int32(action.Type)]
	fmt.Printf("Trigger '%s' fired -> executing action '%s' (%s: '%s')\n", triggerName, actionName, templateTypeName, action.TemplateName)

	switch action.Type {
	case pb.ActionDeploySettingsTemplate_FULL_SETTINGS:
		_,err := tam.rootSvc.SubSysRPC.DeployStoredMasterTemplate(context.Background(), &pb.StringMessage{Msg: action.TemplateName})
		if err == nil {
			fmt.Println("... stored settings deployed")
		} else {
			fmt.Println("... deploying stored settings failed: ", err.Error())
		}
	case pb.ActionDeploySettingsTemplate_NETWORK:
		_,err := tam.rootSvc.SubSysRPC.DeployStoredEthernetInterfaceSettings(context.Background(), &pb.StringMessage{Msg: action.TemplateName})
		if err == nil {
			fmt.Println("... stored settings deployed")
		} else {
			fmt.Println("... deploying stored settings failed: ", err.Error())
		}
	case pb.ActionDeploySettingsTemplate_USB:
		_,err := tam.rootSvc.SubSysRPC.DeployStoredUSBSettings(context.Background(), &pb.StringMessage{Msg: action.TemplateName})
		if err == nil {
			fmt.Println("... stored settings deployed")
		} else {
			fmt.Println("... deploying stored settings failed: ", err.Error())
		}
	case pb.ActionDeploySettingsTemplate_WIFI:
		_,err := tam.rootSvc.SubSysRPC.DeployStoredWifiSettings(context.Background(), &pb.StringMessage{Msg: action.TemplateName})
		if err == nil {
			fmt.Println("... stored settings deployed")
		} else {
			fmt.Println("... deploying stored settings failed: ", err.Error())
		}
	case pb.ActionDeploySettingsTemplate_BLUETOOTH:
		_,err := tam.rootSvc.SubSysRPC.DeployStoredBluetoothSettings(context.Background(), &pb.StringMessage{Msg: action.TemplateName})
		if err == nil {
			fmt.Println("... stored settings deployed")
		} else {
			fmt.Println("... deploying stored settings failed: ", err.Error())
		}
	case pb.ActionDeploySettingsTemplate_TRIGGER_ACTIONS:
		_,err := tam.rootSvc.SubSysRPC.DeployStoredTriggerActionSetReplace(context.Background(), &pb.StringMessage{Msg: action.TemplateName})
		if err == nil {
			fmt.Println("... stored settings deployed")
		} else {
			fmt.Println("... deploying stored settings failed: ", err.Error())
		}
	}


}

func (tam *TriggerActionManager) executeActionGPIOOut(evt *pb.Event, ta *pb.TriggerAction, tt triggerType, at actionType, action *pb.ActionGPIOOut) {
	triggerName := triggerTypeString[tt]
	actionName := actionTypeString[at]

	gpioNumName := action.GpioName
	fmt.Printf("Trigger '%s' fired -> executing action '%s' ('%s')\n", triggerName, actionName, gpioNumName)

	tam.rootSvc.SubSysGpio.FireGpioAction(action)
}

func (tam *TriggerActionManager) executeActionGroupSend(evt *pb.Event, ta *pb.TriggerAction, tt triggerType, at actionType, action *pb.ActionGroupSend) {
	triggerName := triggerTypeString[tt]
	actionName := actionTypeString[at]

	groupName := action.GroupName
	value := action.Value
	fmt.Printf("Trigger '%s' fired -> executing action '%s' ('%s': %d)\n", triggerName, actionName, groupName, value)

	tam.rootSvc.SubSysEvent.Emit(ConstructEventTriggerGroupReceive(groupName, value))
}

func (tam *TriggerActionManager) executeActionStartHidScript(evt *pb.Event, ta *pb.TriggerAction, tt triggerType, at actionType, action *pb.ActionStartHIDScript) {
	triggerName := triggerTypeString[tt]
	actionName := actionTypeString[at]

	fmt.Printf("Trigger '%s' fired -> executing action '%s' ('%s')\n", triggerName, actionName, action.ScriptName)

	scriptPath := common.PATH_HID_SCRIPTS + "/" + action.ScriptName
	preScript := fmt.Sprintf("var TRIGGER='%s';\n", triggerName)

	switch tt {
	case triggerTypeGpioIn:
		gpioPinName := ta.Trigger.(*pb.TriggerAction_GpioIn).GpioIn.GpioName
		preScript += fmt.Sprintf("var GPIO_PIN='%s';\n", gpioPinName)
		_,level,_ := DeconstructEventTriggerGpioIn(evt)
		if level {
			preScript += fmt.Sprintf("var GPIO_LEVEL=true;\n")
		} else {
			preScript += fmt.Sprintf("var GPIO_LEVEL=false;\n")
		}
	case triggerTypeGroupReceiveMulti:
		groupName := ta.Trigger.(*pb.TriggerAction_GroupReceiveMulti).GroupReceiveMulti.GroupName
		values := ta.Trigger.(*pb.TriggerAction_GroupReceiveMulti).GroupReceiveMulti.Values
		rtype := ta.Trigger.(*pb.TriggerAction_GroupReceiveMulti).GroupReceiveMulti.Type
		// create bash array of values
		jsArray := "["
		for idx,v := range values {
			if idx >= len(values) - 1 {
				jsArray += fmt.Sprintf("%d", v)
			} else {
				jsArray += fmt.Sprintf("%d, ", v)
			}
		}
		jsArray += "]"
		preScript += fmt.Sprintf("var GROUP='%s';\n", groupName)
		preScript += fmt.Sprintf("var VALUES=%s;\n", jsArray)
		preScript += fmt.Sprintf("var MULTI_TYPE='%s';\n", pb.GroupReceiveMultiType_name[int32(rtype)])
	case triggerTypeGroupReceive:
		groupName := ta.Trigger.(*pb.TriggerAction_GroupReceive).GroupReceive.GroupName
		value := ta.Trigger.(*pb.TriggerAction_GroupReceive).GroupReceive.Value
		preScript += fmt.Sprintf("var GROUP='%s';\n", groupName)
		preScript += fmt.Sprintf("var VALUE=%d;\n", value)
	case triggerTypeDhcpLeaseGranted:
		iface := evt.Values[1].GetTstring()
		mac := evt.Values[2].GetTstring()
		ip := evt.Values[3].GetTstring()
		host := evt.Values[4].GetTstring()
		preScript += fmt.Sprintf("var DHCP_LEASE_IFACE='%s';\n", iface)
		preScript += fmt.Sprintf("var DHCP_LEASE_MAC='%s';\n", mac)
		preScript += fmt.Sprintf("var DHCP_LEASE_IP='%s';\n", ip)
		preScript += fmt.Sprintf("var DHCP_LEASE_HOST='%s';\n", host)
	case triggerTypeSshLogin:
		loginUser := evt.Values[1].GetTstring()
		preScript += fmt.Sprintf("var SSH_LOGIN_USER='%s';\n", loginUser)

	}

	err := tam.rootSvc.SubSysUSB.HidScriptUsable()
	if err != nil {
		fmt.Printf("Couldn't start HIDScript: %v\n", err)
		return
	}

	scriptFile, err := ioutil.ReadFile(scriptPath)
	if err != nil {
		fmt.Printf("Couldn't load HIDScript '%s': %v\n", scriptPath, err)
		return
	}

	newScriptFile := preScript + string(scriptFile)

	_,err = tam.rootSvc.SubSysUSB.HidScriptStartBackground(context.Background(), newScriptFile)
	if err != nil {
		fmt.Printf("Couldn't start HIDScript as background job'%s': %v\n", action.ScriptName, err)
		return
	}

	return
}

func (tam *TriggerActionManager) executeActionBashScript(evt *pb.Event, ta *pb.TriggerAction, tt triggerType, at actionType, action *pb.ActionStartBashScript) {
	triggerName := triggerTypeString[tt]
	actionName := actionTypeString[at]

	scriptPath := common.PATH_BASH_SCRIPTS + "/" + action.ScriptName
	env := []string{
		fmt.Sprintf("TRIGGER=%s", triggerName),
	}

	switch tt {
	case triggerTypeGpioIn:
		gpioPinName := ta.Trigger.(*pb.TriggerAction_GpioIn).GpioIn.GpioName
		env = append(env, fmt.Sprintf("GPIO_PIN='%s'", gpioPinName))
		_,level,_ := DeconstructEventTriggerGpioIn(evt)
		if level {
			env = append(env, fmt.Sprintf("GPIO_LEVEL=HIGH"))
		} else {
			env = append(env, fmt.Sprintf("GPIO_LEVEL=LOW"))
		}
	case triggerTypeGroupReceiveMulti:
		groupName := ta.Trigger.(*pb.TriggerAction_GroupReceiveMulti).GroupReceiveMulti.GroupName
		values := ta.Trigger.(*pb.TriggerAction_GroupReceiveMulti).GroupReceiveMulti.Values
		rtype := ta.Trigger.(*pb.TriggerAction_GroupReceiveMulti).GroupReceiveMulti.Type
		// create bash array of values
		bashArray := "("
		for _,v := range values { bashArray += fmt.Sprintf("%d ", v)}
		bashArray += ")"
		env = append(env,
			fmt.Sprintf("GROUP=%s", groupName),
			fmt.Sprintf("VALUES=%s", bashArray),
			fmt.Sprintf("MULTI_TYPE=%s", pb.GroupReceiveMultiType_name[int32(rtype)]),
		)
	case triggerTypeGroupReceive:
		groupName := ta.Trigger.(*pb.TriggerAction_GroupReceive).GroupReceive.GroupName
		value := ta.Trigger.(*pb.TriggerAction_GroupReceive).GroupReceive.Value
		env = append(env,
			fmt.Sprintf("GROUP=%s", groupName),
			fmt.Sprintf("VALUE=%d", value),
		)
	case triggerTypeDhcpLeaseGranted:
		iface := evt.Values[1].GetTstring()
		mac := evt.Values[2].GetTstring()
		ip := evt.Values[3].GetTstring()
		host := evt.Values[4].GetTstring()
		env = append(env,
			fmt.Sprintf("DHCP_LEASE_IFACE=%s", iface),
			fmt.Sprintf("DHCP_LEASE_MAC=%s", mac),
			fmt.Sprintf("DHCP_LEASE_IP=%s", ip),
			fmt.Sprintf("DHCP_LEASE_HOST=\"%s\"", host),
		)
	case triggerTypeSshLogin:
		loginUser := evt.Values[1].GetTstring()
		env = append(env,
			fmt.Sprintf("SSH_LOGIN_USER=%s", loginUser),
		)
	}

	fmt.Printf("Trigger '%s' fired -> executing action '%s' ('%s')\n", triggerName, actionName, scriptPath)
	common.RunBashScriptEnv(scriptPath, env...)
}

func (tam *TriggerActionManager) executeActionLog(evt *pb.Event, ta *pb.TriggerAction, tt triggerType, at actionType, action *pb.ActionLog) {
	triggerName := triggerTypeString[tt]
	actionName := actionTypeString[at]

	logMessage := fmt.Sprintf("Trigger fired: %s", triggerName)


	switch tt {
	case triggerTypeGpioIn:
		gpioPinName := ta.Trigger.(*pb.TriggerAction_GpioIn).GpioIn.GpioName
		_,level,_ := DeconstructEventTriggerGpioIn(evt)
		logMessage += fmt.Sprintf(" (GPIO_PIN=%s GPIO_HIGH=%v)", gpioPinName, level)
	case triggerTypeGroupReceiveMulti:
		groupName := ta.Trigger.(*pb.TriggerAction_GroupReceiveMulti).GroupReceiveMulti.GroupName
		values := ta.Trigger.(*pb.TriggerAction_GroupReceiveMulti).GroupReceiveMulti.Values
		logMessage += fmt.Sprintf(" (GROUP='%s', VALUES=%v)", groupName, values)
	case triggerTypeGroupReceive:
		groupName := ta.Trigger.(*pb.TriggerAction_GroupReceive).GroupReceive.GroupName
		values := ta.Trigger.(*pb.TriggerAction_GroupReceive).GroupReceive.Value
		typeName := pb.GroupReceiveMultiType_name[int32(ta.Trigger.(*pb.TriggerAction_GroupReceiveMulti).GroupReceiveMulti.Type)]
		logMessage += fmt.Sprintf(" (GROUP='%s', VALUES=%+v, VALUE='%d')", groupName, values, typeName)
	case triggerTypeDhcpLeaseGranted:
		iface := evt.Values[1].GetTstring()
		mac := evt.Values[2].GetTstring()
		ip := evt.Values[3].GetTstring()
		host := evt.Values[4].GetTstring()
		logMessage += fmt.Sprintf(" (DHCP_LEASE_IFACE=%s, DHCP_LEASE_MAC=%s, DHCP_LEASE_IP=%s, DHCP_LEASE_HOST='%s')", iface, mac, ip, host)
	case triggerTypeSshLogin:
		loginUser := evt.Values[1].GetTstring()
		logMessage += fmt.Sprintf(" (SSH_LOGIN_USER=%s)", loginUser)
	}

	fmt.Printf("Trigger '%s' fired -> executing action '%s'\n", triggerName, actionName)
	tam.rootSvc.SubSysEvent.Emit(ConstructEventLog("TriggerAction", LOG_LEVEL_INFORMATION, logMessage))
}

// checks if the triggerType of the given event (if trigger event at all), matches the TriggerType of the TriggerAction
func taTriggerTypeMatchesEvtTriggerType(ttype triggerType, evt *pb.Event) bool {
	if evt.Type != common_web.EVT_TRIGGER { return false }
	triggerTypeEvt := common_web.EvtTriggerType(evt.Values[0].GetTint64())
	switch triggerTypeEvt {
	case common_web.TRIGGER_EVT_TYPE_SERVICE_STARTED:
		if ttype == triggerTypeServiceStarted {
			return true
		}
	case common_web.TRIGGER_EVT_TYPE_DHCP_LEASE_GRANTED:
		if ttype == triggerTypeDhcpLeaseGranted {
			return true
		}
	case common_web.TRIGGER_EVT_TYPE_WIFI_AP_STARTED:
		if ttype == triggerTypeWifiAPStarted {
			return true
		}
	case common_web.TRIGGER_EVT_TYPE_WIFI_CONNECTED_AS_STA:
		if ttype == triggerTypeWifiConnectedAsSta {
			return true
		}
	case common_web.TRIGGER_EVT_TYPE_USB_GADGET_CONNECTED:
		if ttype == triggerTypeUsbGadgetConnected {
			return true
		}
	case common_web.TRIGGER_EVT_TYPE_USB_GADGET_DISCONNECTED:
		if ttype == triggerTypeUsbGadgetDisconnected {
			return true
		}
	case common_web.TRIGGER_EVT_TYPE_SSH_LOGIN:
		if ttype == triggerTypeSshLogin {
			return true
		}
	case common_web.TRIGGER_EVT_TYPE_GROUP_RECEIVE:
		if ttype == triggerTypeGroupReceive || ttype == triggerTypeGroupReceiveMulti {
			return true
		}
	case common_web.TRIGGER_EVT_TYPE_GPIO_IN:
		if ttype == triggerTypeGpioIn {
			return true
		}
	default:
		return false
	}

	return false
}

// removes the given TriggerAction
// ToDo: for now only the ID is compared, to assure we don't remove a TriggerAction which has been changed meanwhile, we should deep-compare the whole object
func (tam *TriggerActionManager) RemoveTriggerAction(removeTa *pb.TriggerAction) (taRemoved *pb.TriggerAction, err error) {
	tam.registeredTriggerActionMutex.Lock()
	defer tam.registeredTriggerActionMutex.Unlock()


	for idx,ta := range tam.registeredTriggerActions.TriggerActions {
		if ta.Id == removeTa.Id {
			// remove element (not a problem for running `for`-loop, as it is interrupted here)
			tam.registeredTriggerActions.TriggerActions = append(tam.registeredTriggerActions.TriggerActions[:idx], tam.registeredTriggerActions.TriggerActions[idx+1:]...)

			//if target ta trigger had a sequenceChecker assigned, remove it
			if _,match := ta.Trigger.(*pb.TriggerAction_GroupReceiveMulti); match {
				tam.groupReceiveSequenceCheckersMutex.Lock()
				delete(tam.groupReceiveSequenceCheckers, ta)
				tam.groupReceiveSequenceCheckersMutex.Unlock()
			}

			return ta, nil
		}
	}
	return nil, ErrTaNotFound

}

func (tam *TriggerActionManager) GetTriggerActionByID(Id uint32) (ta *pb.TriggerAction ,err error) {
	for _,ta = range tam.registeredTriggerActions.TriggerActions {
		if ta.Id == Id {
			return ta, nil
		}
	}
	return nil, ErrTaNotFound
}

// returns the TriggerAction with assigned ID
func (tam *TriggerActionManager) AddTriggerAction(ta *pb.TriggerAction) (taAdded *pb.TriggerAction, err error) {
	tam.registeredTriggerActionMutex.Lock()
	defer tam.registeredTriggerActionMutex.Unlock()
	ta.Id = tam.nextID
	tam.nextID++
	tam.registeredTriggerActions.TriggerActions = append(tam.registeredTriggerActions.TriggerActions, ta)
	taAdded = ta

	//if new ta trigger is GroupReceiveSequence, add a SequenceChecker
	if triggerGrpRcv,match := ta.Trigger.(*pb.TriggerAction_GroupReceiveMulti); match {
		tam.groupReceiveSequenceCheckersMutex.Lock()
		//fmt.Printf("##### New val checker %+v\n", triggerGrpRcv.GroupReceiveSequence.Values)
		switch triggerGrpRcv.GroupReceiveMulti.Type {
		case pb.GroupReceiveMultiType_AND:
			tam.groupReceiveSequenceCheckers[ta] = util.NewValueSequenceChecker(triggerGrpRcv.GroupReceiveMulti.Values, util.ValueSeqType_AND)
		case pb.GroupReceiveMultiType_OR:
			tam.groupReceiveSequenceCheckers[ta] = util.NewValueSequenceChecker(triggerGrpRcv.GroupReceiveMulti.Values, util.ValueSeqType_OR)
		case pb.GroupReceiveMultiType_SEQUENCE:
			tam.groupReceiveSequenceCheckers[ta] = util.NewValueSequenceChecker(triggerGrpRcv.GroupReceiveMulti.Values, util.ValueSeqType_SEQUENCE)
		case pb.GroupReceiveMultiType_EXACT_SEQUENCE:
			tam.groupReceiveSequenceCheckers[ta] = util.NewValueSequenceChecker(triggerGrpRcv.GroupReceiveMulti.Values, util.ValueSeqType_EXACT_SEQUENCE)
		}

		tam.groupReceiveSequenceCheckersMutex.Unlock()
	}

	//if trigger is GpioIn, configure GPIO
	if triggerGpioIn,match := ta.Trigger.(*pb.TriggerAction_GpioIn); match {
		tam.rootSvc.SubSysGpio.DeployGpioTrigger(triggerGpioIn.GpioIn)
	}

	return taAdded,nil
}


func (tam *TriggerActionManager) UpdateTriggerAction(srcTa *pb.TriggerAction, addIfMissing bool) (err error) {
	tam.registeredTriggerActionMutex.Lock()
	defer tam.registeredTriggerActionMutex.Unlock()

	if targetTA,err := tam.GetTriggerActionByID(srcTa.Id); err != nil {
		if addIfMissing {
			_,err = tam.AddTriggerAction(srcTa)
			return err
		} else {
			return ErrTaNotFound
		}
	} else {
		if targetTA.Immutable { return ErrTaImmutable }

		//if target ta trigger had a sequenceChecker assigned, remove it
		if _,match := targetTA.Trigger.(*pb.TriggerAction_GroupReceiveMulti); match {
			tam.groupReceiveSequenceCheckersMutex.Lock()
			delete(tam.groupReceiveSequenceCheckers, targetTA)
			tam.groupReceiveSequenceCheckersMutex.Unlock()
		}

		targetTA.OneShot = srcTa.OneShot
		targetTA.IsActive = srcTa.IsActive
		targetTA.Immutable = srcTa.Immutable
		targetTA.Action = srcTa.Action
		targetTA.Trigger = srcTa.Trigger

		//if new ta trigger is GroupReceiveSequence, add a SequenceChecker
		if triggerGrpRcv,match := targetTA.Trigger.(*pb.TriggerAction_GroupReceiveMulti); match {
			tam.groupReceiveSequenceCheckersMutex.Lock()

			switch triggerGrpRcv.GroupReceiveMulti.Type {
			case pb.GroupReceiveMultiType_AND:
				tam.groupReceiveSequenceCheckers[targetTA] = util.NewValueSequenceChecker(triggerGrpRcv.GroupReceiveMulti.Values, util.ValueSeqType_AND)
			case pb.GroupReceiveMultiType_OR:
				tam.groupReceiveSequenceCheckers[targetTA] = util.NewValueSequenceChecker(triggerGrpRcv.GroupReceiveMulti.Values, util.ValueSeqType_OR)
			case pb.GroupReceiveMultiType_SEQUENCE:
				tam.groupReceiveSequenceCheckers[targetTA] = util.NewValueSequenceChecker(triggerGrpRcv.GroupReceiveMulti.Values, util.ValueSeqType_SEQUENCE)
			case pb.GroupReceiveMultiType_EXACT_SEQUENCE:
				tam.groupReceiveSequenceCheckers[targetTA] = util.NewValueSequenceChecker(triggerGrpRcv.GroupReceiveMulti.Values, util.ValueSeqType_EXACT_SEQUENCE)
			}
			tam.groupReceiveSequenceCheckersMutex.Unlock()
		}

		//if trigger is GpioIn, configure GPIO
		if triggerGpioIn,match := targetTA.Trigger.(*pb.TriggerAction_GpioIn); match {
			tam.rootSvc.SubSysGpio.DeployGpioTrigger(triggerGpioIn.GpioIn)
		}

		return nil
	}



}

func (tam *TriggerActionManager) ClearTriggerActions(keepImmutable bool) (err error) {
	tam.registeredTriggerActionMutex.Lock()
	defer tam.registeredTriggerActionMutex.Unlock()

	if !keepImmutable {
		tam.registeredTriggerActions.TriggerActions = []*pb.TriggerAction{}
		return
	}

	newTas := []*pb.TriggerAction{}
	for _,ta := range tam.registeredTriggerActions.TriggerActions {
		if ta.Immutable {
			newTas = append(newTas, ta)
		}
	}
	tam.registeredTriggerActions.TriggerActions = newTas

	tam.groupReceiveSequenceCheckersMutex.Lock()
	defer tam.groupReceiveSequenceCheckersMutex.Unlock()
	tam.groupReceiveSequenceCheckers = make(map[*pb.TriggerAction]*util.ValueSequenceChecker)

	return nil
}

func (tam *TriggerActionManager) GetCurrentTriggerActionSet() (ta *pb.TriggerActionSet) {
	/*
	tam.registeredTriggerActionMutex.Lock()
	resTAs := make([]*pb.TriggerAction, len(tam.registeredTriggerActions.TriggerActions))
	copy(resTAs, tam.registeredTriggerActions.TriggerActions)
	tam.registeredTriggerActionMutex.Unlock()

	return &pb.TriggerActionSet{ TriggerActions: resTAs }
	*/
	return &tam.registeredTriggerActions
}

func (tam *TriggerActionManager) redeployGpioForAllTas() {
	tam.registeredTriggerActionMutex.Lock()
	defer tam.registeredTriggerActionMutex.Unlock()
	tam.rootSvc.SubSysGpio.ResetPins()
	for _,ta := range tam.registeredTriggerActions.TriggerActions {
		ttype, _ := retrieveTriggerActionTypes(ta)
		if ttype == triggerTypeGpioIn {
			gpioIn := ta.Trigger.(*pb.TriggerAction_GpioIn).GpioIn
			tam.rootSvc.SubSysGpio.DeployGpioTrigger(gpioIn)
		}
	}
}

func (tam *TriggerActionManager) Start() {
	tam.evtRcv = tam.rootSvc.SubSysEvent.RegisterReceiver(common_web.EVT_TRIGGER)
	go tam.processing_loop()
}

func (tam *TriggerActionManager) Stop() {
	tam.rootSvc.SubSysEvent.UnregisterReceiver(tam.evtRcv) // should end the processing loop, as the context of the event receiver is closed
}

func NewTriggerActionManager(rootService *Service) (tam *TriggerActionManager) {
	tam = &TriggerActionManager{
		registeredTriggerActions:      pb.TriggerActionSet{
			Name: "DeployedTriggerActions",
			TriggerActions: []*pb.TriggerAction{},
		},
		registeredTriggerActionMutex: &sync.Mutex{},
		rootSvc: rootService,


		groupReceiveSequenceCheckers: make(map[*pb.TriggerAction]*util.ValueSequenceChecker),
		groupReceiveSequenceCheckersMutex: &sync.Mutex{},
	}

	return tam
}

