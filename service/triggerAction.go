package service

import (
	"errors"
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
		scriptPath := PATH_BASH_SCRIPTS + "/" + bs.ScriptName
		go common.RunBashScriptEnv(scriptPath)
		fmt.Printf("Fire bash script '%s'\n", scriptPath)
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
		scriptPath := PATH_BASH_SCRIPTS + "/" + bs.ScriptName
		envUser := fmt.Sprintf("SSH_LOGIN_USER=%s", loginUser)
		go common.RunBashScriptEnv(scriptPath, envUser)
		//go common.RunBashScript(bs.ScriptPath)
		fmt.Printf("Started bash script '%s' (%s)\n", scriptPath, envUser)
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
		scriptPath := PATH_BASH_SCRIPTS + "/" + bs.ScriptName
		envIface := fmt.Sprintf("DHCP_LEASE_IFACE=%s", iface)
		envMac := fmt.Sprintf("DHCP_LEASE_MAC=%s", mac)
		envIp := fmt.Sprintf("DHCP_LEASE_IP=%s", ip)
		go common.RunBashScriptEnv(scriptPath, envIface, envMac, envIp)
		//go common.RunBashScript(bs.ScriptPath)
		fmt.Printf("Started bash script '%s' (%s, %s, %s)\n", scriptPath, envIface, envMac, envIp)
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
	for _,ta := range tam.registeredTriggerAction {
		// ToDo: handle errors of fireAction* methods
		// ToDo: fire action methods have to return an additional int, indicating if the action has really fired (filter function in action could prevent this, which would cause wrong behavior for OneShot riggerActions)

		// skip disabled triggeractions
		if !ta.IsActive { continue }
		if taTriggerTypeMatchesEvtTriggerType(ta, evt) {
			hasFired := true
			switch ttEvt := common_web.EvtTriggerType(evt.Values[0].GetTint64()); ttEvt {
			case common_web.EVT_TRIGGER_TYPE_SERVICE_STARTED:
				tam.fireActionServiceStarted(ta)
			case common_web.EVT_TRIGGER_TYPE_SSH_LOGIN:
				loginUser := evt.Values[1].GetTstring()
				tam.fireActionSSHLogin(loginUser, ta)
			case common_web.EVT_TRIGGER_TYPE_WIFI_CONNECTED_AS_STA:
				tam.fireActionWifiConnectedAsSta(ta)
			case common_web.EVT_TRIGGER_TYPE_WIFI_AP_STARTED:
				tam.fireActionWifiApStarted(ta)
			case common_web.EVT_TRIGGER_TYPE_DHCP_LEASE_GRANTED:
				// extract iface, mac and ip from event
				iface := evt.Values[1].GetTstring()
				mac := evt.Values[2].GetTstring()
				ip := evt.Values[3].GetTstring()
				tam.fireActionDhcpLeaseGranted(iface, mac, ip, ta)
			case common_web.EVT_TRIGGER_TYPE_USB_GADGET_CONNECTED:
				tam.fireActionUsbGadgetConnected(ta)
			case common_web.EVT_TRIGGER_TYPE_USB_GADGET_DISCONNECTED:
				tam.fireActionUsbGadgetDisconnected(ta)
			default:
				hasFired = false
				fmt.Println("unhandled trigger: ", ttEvt)
			}

			if hasFired && ta.OneShot {
				//markedForRemoval = append(markedForRemoval,idx)
				ta.IsActive = false // don't delete, but deactivate
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

// returns the TriggerAction with assigned ID
func (tam *TriggerActionManager) AddTriggerAction(ta *pb.TriggerAction) (taAdded *pb.TriggerAction, err error) {
	tam.registeredTriggerActionMutex.Lock()
	defer tam.registeredTriggerActionMutex.Unlock()
	ta.Id = tam.nextID
	tam.nextID++
	tam.registeredTriggerAction = append(tam.registeredTriggerAction, ta)

	return taAdded,nil
}

var (
	ErrTaNotFound = errors.New("Couldn't find given TriggerAction")
	ErrTaImmutable = errors.New("Not allowed to change immutable TriggerAction")
)

func (tam *TriggerActionManager) GetTriggerActionByID(Id uint32) (ta *pb.TriggerAction ,err error) {
	for _,ta = range tam.registeredTriggerAction {
		if ta.Id == Id {
			return ta, nil
		}
	}
	return nil, ErrTaNotFound
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

		targetTA.OneShot = srcTa.OneShot
		targetTA.IsActive = srcTa.IsActive
		targetTA.Immutable = srcTa.Immutable
		targetTA.Action = srcTa.Action
		targetTA.Trigger = srcTa.Trigger
		return nil
	}
}

func (tam *TriggerActionManager) ClearTriggerActions(keepImmutable bool) (err error) {
	tam.registeredTriggerActionMutex.Lock()
	defer tam.registeredTriggerActionMutex.Unlock()

	if !keepImmutable {
		tam.registeredTriggerAction = []*pb.TriggerAction{}
		return
	}

	newTas := []*pb.TriggerAction{}
	for _,ta := range tam.registeredTriggerAction {
		if ta.Immutable {
			newTas = append(newTas, ta)
		}
	}
	tam.registeredTriggerAction = newTas
	return nil
}

func (tam *TriggerActionManager) GetCurrentTriggerActionSet() (ta *pb.TriggerActionSet) {
	tam.registeredTriggerActionMutex.Lock()
	resTAs := make([]*pb.TriggerAction, len(tam.registeredTriggerAction))
	copy(resTAs, tam.registeredTriggerAction)
	tam.registeredTriggerActionMutex.Unlock()

	return &pb.TriggerActionSet{ TriggerActions: resTAs }
}

func (tam *TriggerActionManager) Start() {
	tam.evtRcv = tam.rootSvc.SubSysEvent.RegisterReceiver(common_web.EVT_TRIGGER) // ToDo: change to trigger event type, once defined
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

