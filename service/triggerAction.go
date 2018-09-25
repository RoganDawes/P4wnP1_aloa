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
				// check if relevant and dispatch to triggers
			case <- tam.evtRcv.Ctx.Done():
				// evvent Receiver cancelled or unregistered
				break Outer
			}
		}
	fmt.Println("TAM processing loop finished")
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
	tam.AddTriggerAction(serviceUpRunScript)

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

