package service

import (
	"fmt"
	pb "../proto"
	"../common"
	"context"
	"sync"
	"time"
)

var (
	EvMgr    *EventManager
	evmMutex  = &sync.Mutex{}
)

func pDEBUG(message string) {
	fmt.Println("EVENT DEBUG: " + message)
}

type EventManager struct {
	eventQueue chan *pb.Event
	ctx context.Context
	cancel context.CancelFunc

	registeredReceivers map[*EventReceiver]bool
	receiverDeleteList map[*EventReceiver]bool
	receiverRegisterList map[*EventReceiver]bool
	receiverDelListMutex *sync.Mutex
	receiverRegListMutex *sync.Mutex
}

func StartEventManager(queueSize int) *EventManager {
	if EvMgr != nil { StopEventManager() }

	evmMutex.Lock()
	defer evmMutex.Unlock()

	EvMgr = &EventManager{
		eventQueue: make(chan *pb.Event, queueSize),
		receiverDelListMutex: &sync.Mutex{},
		receiverRegListMutex: &sync.Mutex{},
		receiverRegisterList: make(map[*EventReceiver]bool),
		registeredReceivers: make(map[*EventReceiver]bool),
		receiverDeleteList: make(map[*EventReceiver]bool),
	}
	EvMgr.ctx, EvMgr.cancel = context.WithCancel(context.Background())

	pDEBUG("EvtMgr started")
	go EvMgr.dispatch()

	return EvMgr
}

func StopEventManager() {
	evmMutex.Lock()
	defer evmMutex.Unlock()

	if EvMgr == nil { return }
	EvMgr.cancel()
	close(EvMgr.eventQueue)

}

func (em *EventManager) Emit(event *pb.Event) {
	em.eventQueue <-event
//	fmt.Println("Event enqueued")
}

func (em *EventManager) Write(p []byte) (n int, err error) {
	ev := ConstructEventLog("logWriter", 1, string(p))
	em.Emit(ev)
	return len(p),nil
}


func (em *EventManager) RegisterReceiver(filterEventType int64) *EventReceiver {
	ctx,cancel := context.WithCancel(context.Background())
	er := &EventReceiver{
		EventQueue: make(chan *pb.Event, 10), //allow buffering 10 events per receiver
		Ctx: ctx,
		Cancel: cancel,
		FilterEventType: filterEventType,
	}
	em.receiverRegListMutex.Lock()
	em.receiverRegisterList[er] = true
	em.receiverRegListMutex.Unlock()

	return er
}

func (em *EventManager) UnregisterReceiver(receiver *EventReceiver) {
	em.receiverDelListMutex.Lock()
	em.receiverDeleteList[receiver] = true
	em.receiverDelListMutex.Unlock()
}

func (em *EventManager) dispatch() {
	fmt.Println("Started event dispatcher")
	pDEBUG("Started dispatcher")
	loop:
	for {
		select {
		case evToDispatch := <- em.eventQueue:
			// distribute to registered receiver
			// Note: no mutex on em.registeredReceivers needed, only accessed in this method
			for receiver := range em.registeredReceivers {
				// check if this receiver is listening for this event type
				if receiver.FilterEventType == evToDispatch.Type || receiver.FilterEventType == common.EVT_ANY {
					select {
					case <-receiver.Ctx.Done():
						//receiver canceled
						em.UnregisterReceiver(receiver)
						continue // go on with next registered receiver
					case receiver.EventQueue <- evToDispatch:
						//Do nothing
					}
				}
			}

			// delete receivers marked for deletion (only unregister function is allowed to put data into this map)
			em.receiverDelListMutex.Lock()
			for delReceiver := range em.receiverDeleteList {
				delete(em.registeredReceivers, delReceiver)
				close(delReceiver.EventQueue)
			}
			//Replace the delete list with a new one and let the GC take care of the old
			em.receiverDeleteList = make(map[*EventReceiver]bool)
			em.receiverDelListMutex.Unlock()

			//add newly registered receivers
			em.receiverRegListMutex.Lock()
			for addReceiver := range em.receiverRegisterList {
				em.registeredReceivers[addReceiver] = true
			}
			//Replace the register list with a new one and let the GC take care of the old
			em.receiverRegisterList = make(map[*EventReceiver]bool)
			em.receiverRegListMutex.Unlock()


		case <-em.ctx.Done():
			//EventManage aborted

			// ToDo: close all eventReceivers eventQueues, to notify them of the stopped dispatcher
			pDEBUG("EvtMgr cancelled")
			break loop
		}
	}
	fmt.Println("Stopped event dispatcher")
	pDEBUG("Stopped dispatcher")
}



type EventReceiver struct {
	Ctx context.Context
	Cancel context.CancelFunc
	EventQueue chan *pb.Event
	FilterEventType int64
}



func ConstructEventLog(source string, level int, message string) *pb.Event {

	return &pb.Event{
		Type: common.EVT_LOG,
		Values: []*pb.EventValue{
			{Val: &pb.EventValue_Tstring{Tstring:source} },
			{Val: &pb.EventValue_Tint64{Tint64:int64(level)} },
			{Val: &pb.EventValue_Tstring{Tstring:message} },
			{Val: &pb.EventValue_Tstring{Tstring:time.Now().String()} },
		},
	}
}

