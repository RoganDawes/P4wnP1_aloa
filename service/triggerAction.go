package service

import (
	"fmt"
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
				fmt.Println("TriggerActionManager received unfiltered event", evt)
				tam.dispatchTriggerEvent(evt)
				// check if relevant and dispatch to triggers
			case <- tam.evtRcv.Ctx.Done():
				// evvent Receiver cancelled or unregistered
				break Outer
			}
		}
	fmt.Println("TAM processing loop finished")
}

func (tam *TriggerActionManager) dispatchTriggerEvent(evt *pb.Event) {
	if evt.Type != common_web.EVT_TRIGGER { return }

	triggerTypeRcv := common_web.EvtTriggerType(evt.Values[0].GetTint64())

	tam.registeredTriggerActionMutex.Lock()
	defer tam.registeredTriggerActionMutex.Unlock()
	for _,ta := range tam.registeredTriggerAction {
		switch triggerType := ta.Trigger.(type) {
		case *pb.TriggerAction_ServiceStarted:
			if triggerTypeRcv == common_web.EVT_TRIGGER_TYPE_SERVICE_STARTED {
				// trigger action of ta
				tam.fireAction(ta.Trigger, ta.Action)
			}
		case *pb.TriggerAction_UsbGadgetConnected:
			if triggerTypeRcv == common_web.EVT_TRIGGER_TYPE_USB_GADGET_CONNECTED {
				// trigger action of ta
				tam.fireAction(triggerType.UsbGadgetConnected, ta.Action)
			}
		case *pb.TriggerAction_UsbGadgetDisconnected:
			if triggerTypeRcv == common_web.EVT_TRIGGER_TYPE_USB_GADGET_DISCONNECTED {
				// trigger action of ta
				tam.fireAction(triggerType.UsbGadgetDisconnected, ta.Action)
			}
		case *pb.TriggerAction_DhcpLeaseGranted:
			if triggerTypeRcv == common_web.EVT_TRIGGER_TYPE_USB_GADGET_DISCONNECTED {
				// trigger action of ta
				tam.fireAction(triggerType.DhcpLeaseGranted, ta.Action)
			}
		}
	}
}

func (tam *TriggerActionManager) fireAction(trigger interface{}, action interface{}) (err error ) {
	switch actionType := action.(type) {
	case *pb.TriggerAction_BashScript:
		bs := actionType.BashScript
		fmt.Printf("Fire bash script '%s'\n", bs.ScriptPath)
	case *pb.TriggerAction_HidScript:
		hs := actionType.HidScript
		fmt.Printf("Starting HID script '%s'\n", hs.ScriptName)
	case *pb.TriggerAction_DeploySettingsTemplate:
		st := actionType.DeploySettingsTemplate
		strType := pb.ActionDeploySettingsTemplate_TemplateType_name[int32(st.Type)]
		fmt.Printf("Deploy settings template of type [%s] with name '%s'\n", strType, st.TemplateName)
	case *pb.TriggerAction_Log:
		fmt.Printf("Logging trigger '%+v'\n", trigger)
	}
	return nil
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



	// create test trigger
	serviceUpRunScript := &pb.TriggerAction{
		Trigger: &pb.TriggerAction_ServiceStarted{
			ServiceStarted: &pb.TriggerServiceStarted{},
		},
		Action: &pb.TriggerAction_BashScript{
			BashScript: &pb.ActionStartBashScript{
				ScriptPath: "/usr/local/P4wnP1/scripts/servicestart.sh",
			},
		},
	}
	serviceUpLog := &pb.TriggerAction{
		Trigger: &pb.TriggerAction_ServiceStarted{
			ServiceStarted: &pb.TriggerServiceStarted{},
		},
		Action: &pb.TriggerAction_Log{
			Log: &pb.ActionLog{},
		},
	}
	tam.AddTriggerAction(serviceUpRunScript)
	tam.AddTriggerAction(serviceUpLog)

	fmt.Printf("TEST TRIGGER ACTION: Service up %+v\n", serviceUpRunScript)
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

