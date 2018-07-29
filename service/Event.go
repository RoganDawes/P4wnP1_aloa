package service

import (
	"fmt"
	pb "../proto"
	"../common"
	"context"
	"sync"
	"time"
	"log"
	"../hid"
)


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

func NewEventManager(queueSize int) *EventManager {
	EvMgr := &EventManager{
		eventQueue: make(chan *pb.Event, queueSize),
		receiverDelListMutex: &sync.Mutex{},
		receiverRegListMutex: &sync.Mutex{},
		receiverRegisterList: make(map[*EventReceiver]bool),
		registeredReceivers: make(map[*EventReceiver]bool),
		receiverDeleteList: make(map[*EventReceiver]bool),
	}
	EvMgr.ctx, EvMgr.cancel = context.WithCancel(context.Background())
	return EvMgr
}

func (evm *EventManager) Start() {
	log.Println("Event Manager: Starting event dispatcher")
	go evm.dispatch()
}

func (evm *EventManager) Stop() {
	log.Println("Event Manager: Stopping ...")
	evm.cancel()
	close(evm.eventQueue)
}

func (em *EventManager) Emit(event *pb.Event) {
	//fmt.Printf("Emitting event: %+v\n", event)
	em.eventQueue <-event
}

func (em *EventManager) Write(p []byte) (n int, err error) {
	ev := ConstructEventLog("logWriter", 1, string(p))
	em.Emit(ev)
	return len(p),nil
}


func (em *EventManager) RegisterReceiver(filterEventType int64) *EventReceiver {
//	fmt.Println("!!!Event listener registered for " + strconv.Itoa(int(filterEventType)))

	ctx,cancel := context.WithCancel(context.Background())
	er := &EventReceiver{
		EventQueue: make(chan *pb.Event, 10), //allow buffering 10 events per receiver
		Ctx: ctx,
		Cancel: cancel,
		FilterEventType: filterEventType,
	}
	fmt.Printf("Registered receiver for %d\n", er.FilterEventType)
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
			//fmt.Println("EvtMgr cancelled")
			break loop
		}
	}
	fmt.Println("Stopped event dispatcher")
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

func ConstructEventHID(hidEvent hid.Event) *pb.Event {
	//subType, vmID, jobID int, error bool, resString, errString, message string
	vmID := -1
	jobID := -1
	hasError := false
	errString := ""
	message := hidEvent.Message
	resString := ""
	if job := hidEvent.Job; job != nil {
		jobID = job.Id
		if job.ResultErr != nil {
			hasError = true
			errString = job.ResultErr.Error()
		}
		resString,_ = job.ResultJsonString()
	}
	if eVM := hidEvent.Vm; eVM != nil { vmID = eVM.Id }

	return &pb.Event{
		Type: common.EVT_HID, //Type
		Values: []*pb.EventValue{
			{Val: &pb.EventValue_Tint64{Tint64:int64(hidEvent.Type)} }, 		//SubType = Type of hid.Event
			{Val: &pb.EventValue_Tint64{Tint64:int64(vmID)} },			//ID of VM
			{Val: &pb.EventValue_Tint64{Tint64:int64(jobID)} },			//ID of job
			{Val: &pb.EventValue_Tbool{Tbool:hasError} },					//isError (f.e. if a job was interrupted)
			{Val: &pb.EventValue_Tstring{Tstring:resString} },			//result String
			{Val: &pb.EventValue_Tstring{Tstring:errString} },			//error String (message in case of error)
			{Val: &pb.EventValue_Tstring{Tstring:message} },			//Mesage text of event
			{Val: &pb.EventValue_Tstring{Tstring:time.Now().String()} },//Timestamp of event genration
		},
	}
}

