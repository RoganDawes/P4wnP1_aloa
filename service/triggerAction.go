package service

import (
	"fmt"
	"github.com/mame82/P4wnP1_go/common"
	"github.com/mame82/P4wnP1_go/common_web"
	pb "github.com/mame82/P4wnP1_go/proto"
	"sync"
)

type TriggerActionManager struct {
	rootSvc *Service
	evtRcv *EventReceiver

	registeredTriggerActionMutex *sync.Mutex
	registeredTriggerAction      []*pb.TriggerAction

	nextID                       uint32
}

func (tam *TriggerActionManager) processing_loop() {
	// (un)register event listener(s)
	tam.evtRcv = tam.rootSvc.SubSysEvent.RegisterReceiver(common_web.EVT_ANY) // ToDo: change to trigger event type, once defined
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
				tam.dispatchTriggerEvent(evt)
				// check if relevant and dispatch to triggers
			case <- tam.evtRcv.Ctx.Done():
				// evvent Receiver cancelled or unregistered
				break Outer
			}
		}
	fmt.Println("TAM processing loop finished")
}

func (tam *TriggerActionManager) fireActionNoArgs(ta *pb.TriggerAction) (err error ) {
	switch actionType := ta.Action.(type) {
	case *pb.TriggerAction_BashScript:
		bs := actionType.BashScript
		go common.RunBashScriptEnv(bs.ScriptPath)
		fmt.Printf("Fire bash script '%s'\n", bs.ScriptPath)
	case *pb.TriggerAction_HidScript:
		// ToDo: Implement
		hs := actionType.HidScript
		fmt.Printf("Placeholder: Starting HID script '%s'\n", hs.ScriptName)
	case *pb.TriggerAction_DeploySettingsTemplate:
		// ToDo: Implement
		st := actionType.DeploySettingsTemplate
		strType := pb.ActionDeploySettingsTemplate_TemplateType_name[int32(st.Type)]
		fmt.Printf("Placeholder: Deploy settings template of type [%s] with name '%s'\n", strType, st.TemplateName)
	case *pb.TriggerAction_Log:
		fmt.Printf("Logging trigger '%+v'\n", ta.Trigger)
	}
	return nil
}


func (tam *TriggerActionManager) fireActionServiceStarted(ta *pb.TriggerAction) error {
	return tam.fireActionNoArgs(ta)
}

func (tam *TriggerActionManager) fireActionSSHLogin(loginUser string, ta *pb.TriggerAction) error {
	switch actionType := ta.Action.(type) {
	case *pb.TriggerAction_BashScript:
		bs := actionType.BashScript
		envUser := fmt.Sprintf("SSH_LOGIN_USER=%s", loginUser)
		go common.RunBashScriptEnv(bs.ScriptPath, envUser)
		//go common.RunBashScript(bs.ScriptPath)
		fmt.Printf("Started bash script '%s' (%s)\n", bs.ScriptPath, envUser)
	case *pb.TriggerAction_HidScript:
		// ToDo: Implement
		hs := actionType.HidScript
		fmt.Printf("Placeholder: Starting HID script '%s'\n", hs.ScriptName)
	case *pb.TriggerAction_DeploySettingsTemplate:
		// ToDo: Implement
		st := actionType.DeploySettingsTemplate
		strType := pb.ActionDeploySettingsTemplate_TemplateType_name[int32(st.Type)]
		fmt.Printf("Placeholder: Deploy settings template of type [%s] with name '%s'\n", strType, st.TemplateName)
	case *pb.TriggerAction_Log:
		fmt.Printf("Logging action: SSHLogin user: '%s'\n", loginUser)
	}
	return nil
}

func (tam *TriggerActionManager) fireActionWifiConnectedAsSta(ta *pb.TriggerAction) error {
	return tam.fireActionNoArgs(ta)
}

func (tam *TriggerActionManager) fireActionWifiApStarted(ta *pb.TriggerAction) error {
	return tam.fireActionNoArgs(ta)
}

func (tam *TriggerActionManager) fireActionDhcpLeaseGranted(iface string, mac string, ip string, ta *pb.TriggerAction) error {
	switch actionType := ta.Action.(type) {
	case *pb.TriggerAction_BashScript:
		bs := actionType.BashScript
		envIface := fmt.Sprintf("DHCP_LEASE_IFACE=%s", iface)
		envMac := fmt.Sprintf("DHCP_LEASE_MAC=%s", mac)
		envIp := fmt.Sprintf("DHCP_LEASE_IP=%s", ip)
		go common.RunBashScriptEnv(bs.ScriptPath, envIface, envMac, envIp)
		//go common.RunBashScript(bs.ScriptPath)
		fmt.Printf("Started bash script '%s' (%s, %s, %s)\n", bs.ScriptPath, envIface, envMac, envIp)
	case *pb.TriggerAction_HidScript:
		// ToDo: Implement
		hs := actionType.HidScript
		fmt.Printf("Placeholder: Starting HID script '%s'\n", hs.ScriptName)
	case *pb.TriggerAction_DeploySettingsTemplate:
		// ToDo: Implement
		st := actionType.DeploySettingsTemplate
		strType := pb.ActionDeploySettingsTemplate_TemplateType_name[int32(st.Type)]
		fmt.Printf("Placeholder: Deploy settings template of type [%s] with name '%s'\n", strType, st.TemplateName)
	case *pb.TriggerAction_Log:
		fmt.Printf("Logging action: DHCPLeaseGranted interface: '%s' mac:'%s' IP: '%s'\n", iface, mac, ip)
	}
	return nil
}

func (tam *TriggerActionManager) fireActionUsbGadgetConnected(ta *pb.TriggerAction) error {
	return tam.fireActionNoArgs(ta)
}

func (tam *TriggerActionManager) fireActionUsbGadgetDisconnected(ta *pb.TriggerAction) error {
	return tam.fireActionNoArgs(ta)
}

// checks if the triggerType of the given event (if trigger event at all), matches the TriggerType of the TriggerAction
func taTriggerTypeMatchesEvtTriggerType(ta *pb.TriggerAction, evt *pb.Event) bool {
	if evt.Type != common_web.EVT_TRIGGER { return false }
	triggerTypeEvt := common_web.EvtTriggerType(evt.Values[0].GetTint64())
	switch triggerTypeEvt {
	case common_web.EVT_TRIGGER_TYPE_SERVICE_STARTED:
		if _,match := ta.Trigger.(*pb.TriggerAction_ServiceStarted); match {
			return true
		} else {
			return false
		}
	case common_web.EVT_TRIGGER_TYPE_DHCP_LEASE_GRANTED:
		if _,match := ta.Trigger.(*pb.TriggerAction_DhcpLeaseGranted); match {
			return true
		} else {
			return false
		}
	case common_web.EVT_TRIGGER_TYPE_WIFI_AP_STARTED:
		if _,match := ta.Trigger.(*pb.TriggerAction_WifiAPStarted); match {
			return true
		} else {
			return false
		}
	case common_web.EVT_TRIGGER_TYPE_WIFI_CONNECTED_AS_STA:
		if _,match := ta.Trigger.(*pb.TriggerAction_WifiConnectedAsSta); match {
			return true
		} else {
			return false
		}
	case common_web.EVT_TRIGGER_TYPE_USB_GADGET_CONNECTED:
		if _,match := ta.Trigger.(*pb.TriggerAction_UsbGadgetConnected); match {
			return true
		} else {
			return false
		}
	case common_web.EVT_TRIGGER_TYPE_USB_GADGET_DISCONNECTED:
		if _,match := ta.Trigger.(*pb.TriggerAction_UsbGadgetDisconnected); match {
			return true
		} else {
			return false
		}
	case common_web.EVT_TRIGGER_TYPE_SSH_LOGIN:
		if _,match := ta.Trigger.(*pb.TriggerAction_SshLogin); match {
			return true
		} else {
			return false
		}
	default:
		return false
	}
}

func (tam *TriggerActionManager) dispatchTriggerEvent(evt *pb.Event) {
	//fmt.Printf("Remaining triggerActions: %+v\n", tam.registeredTriggerAction)
	//fmt.Printf("Received event: %+v\n", evt)
	tam.registeredTriggerActionMutex.Lock()
	defer tam.registeredTriggerActionMutex.Unlock()
	markedForRemoval := []int{}
	for idx,ta := range tam.registeredTriggerAction {
		// ToDo: handle errors of fireAction* methods
		if taTriggerTypeMatchesEvtTriggerType(ta, evt) {
			switch ttEvt := common_web.EvtTriggerType(evt.Values[0].GetTint64()); ttEvt {
			case common_web.EVT_TRIGGER_TYPE_SERVICE_STARTED:
				tam.fireActionServiceStarted(ta)
				if ta.OneShot {
					markedForRemoval = append(markedForRemoval,idx)
				}
			case common_web.EVT_TRIGGER_TYPE_SSH_LOGIN:
				loginUser := evt.Values[1].GetTstring()
				tam.fireActionSSHLogin(loginUser, ta)
				if ta.OneShot {
					markedForRemoval = append(markedForRemoval,idx)
				}
			case common_web.EVT_TRIGGER_TYPE_WIFI_CONNECTED_AS_STA:
				tam.fireActionWifiConnectedAsSta(ta)
				if ta.OneShot {
					markedForRemoval = append(markedForRemoval,idx)
				}
			case common_web.EVT_TRIGGER_TYPE_WIFI_AP_STARTED:
				tam.fireActionWifiApStarted(ta)
				if ta.OneShot {
					markedForRemoval = append(markedForRemoval,idx)
				}
			case common_web.EVT_TRIGGER_TYPE_DHCP_LEASE_GRANTED:
				// extract iface, mac and ip from event
				iface := evt.Values[1].GetTstring()
				mac := evt.Values[2].GetTstring()
				ip := evt.Values[3].GetTstring()
				tam.fireActionDhcpLeaseGranted(iface, mac, ip, ta)
				if ta.OneShot {
					markedForRemoval = append(markedForRemoval,idx)
				}
			case common_web.EVT_TRIGGER_TYPE_USB_GADGET_CONNECTED:
				tam.fireActionUsbGadgetConnected(ta)
				if ta.OneShot {
					markedForRemoval = append(markedForRemoval,idx)
				}
			case common_web.EVT_TRIGGER_TYPE_USB_GADGET_DISCONNECTED:
				tam.fireActionUsbGadgetDisconnected(ta)
				if ta.OneShot {
					markedForRemoval = append(markedForRemoval,idx)
				}
			default:
				fmt.Println("unhandled trigger: ", ttEvt)
			}
		}
	}

	//fmt.Println("Indexes of TriggerActions to remove, because thy are OneShots and have fired", markedForRemoval)
	for _,delIdx := range markedForRemoval {
		// doesn't preserve order
		tam.registeredTriggerAction[delIdx] = tam.registeredTriggerAction[len(tam.registeredTriggerAction)-1]
		tam.registeredTriggerAction = tam.registeredTriggerAction[:len(tam.registeredTriggerAction)-1]
	}
	return
}


func (tam *TriggerActionManager) AddTriggerAction(ta *pb.TriggerAction) (err error) {
	tam.registeredTriggerActionMutex.Lock()
	defer tam.registeredTriggerActionMutex.Unlock()
	ta.Id = tam.nextID
	tam.nextID++
	tam.registeredTriggerAction = append(tam.registeredTriggerAction, ta)

	return nil
}

func (tam *TriggerActionManager) Start() {
	go tam.processing_loop()
}

func (tam *TriggerActionManager) Stop() {
	tam.rootSvc.SubSysEvent.UnregisterReceiver(tam.evtRcv) // should end the processing loop, as the context of the event receiver is closed
}

func NewTriggerActionManager(rootService *Service) (tam *TriggerActionManager) {
	tam = &TriggerActionManager{
		registeredTriggerAction:      []*pb.TriggerAction{},
		registeredTriggerActionMutex: &sync.Mutex{},
		rootSvc: rootService,
	}

	return tam
}

