package service

import (
	"errors"
	"fmt"
	pb "github.com/mame82/P4wnP1_go/proto"
	"context"
	"sync"
	"time"
	"log"
	"github.com/mame82/P4wnP1_go/hid"
	"github.com/mame82/P4wnP1_go/common_web"
)

type EventManager struct {
	eventQueue chan *pb.Event
	ctx        context.Context
	cancel     context.CancelFunc

	registeredReceivers  map[*EventReceiver]bool
	receiverDeleteList   map[*EventReceiver]bool
	receiverRegisterList map[*EventReceiver]bool
	receiverDelListMutex *sync.Mutex
	receiverRegListMutex *sync.Mutex
}

func NewEventManager(queueSize int) *EventManager {
	EvMgr := &EventManager{
		eventQueue:           make(chan *pb.Event, queueSize),
		receiverDelListMutex: &sync.Mutex{},
		receiverRegListMutex: &sync.Mutex{},
		receiverRegisterList: make(map[*EventReceiver]bool),
		registeredReceivers:  make(map[*EventReceiver]bool),
		receiverDeleteList:   make(map[*EventReceiver]bool),
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
	em.eventQueue <- event
}

func (em *EventManager) Write(p []byte) (n int, err error) {
	ev := ConstructEventLog("logWriter", 1, string(p))
	em.Emit(ev)
	return len(p), nil
}

func (em *EventManager) RegisterReceiver(filterEventType int64) *EventReceiver {
	//	fmt.Println("!!!Event listener registered for " + strconv.Itoa(int(filterEventType)))

	ctx, cancel := context.WithCancel(context.Background())
	er := &EventReceiver{
		EventQueue:      make(chan *pb.Event, 10), //allow buffering 10 events per receiver
		Ctx:             ctx,
		Cancel:          cancel,
		FilterEventType: filterEventType,
	}
	fmt.Printf("Registered receiver for %d\n", er.FilterEventType)
	em.receiverRegListMutex.Lock()
	em.receiverRegisterList[er] = true
	er.isRegistered = true
	em.receiverRegListMutex.Unlock()

	return er
}

func (em *EventManager) UnregisterReceiver(receiver *EventReceiver) {
	if !receiver.isRegistered {
		return
	}
	em.receiverDelListMutex.Lock()
	em.receiverDeleteList[receiver] = true
	receiver.isRegistered = false
	em.receiverDelListMutex.Unlock()
}

func (em *EventManager) dispatch() {
	fmt.Println("Started event dispatcher")
loop:
	for {
		select {
		case evToDispatch := <-em.eventQueue:
			// delete receivers marked for deletion (only unregister function is allowed to put data into this map)
			em.receiverDelListMutex.Lock()
			for delReceiver := range em.receiverDeleteList {
				delete(em.registeredReceivers, delReceiver)
				delReceiver.Cancel() // cancel context BEFORE closing the eventQueue channel
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

			// distribute to registered receiver
			// Note: no mutex on em.registeredReceivers needed, only accessed in this method
			for receiver := range em.registeredReceivers {
				// check if this receiver is listening for this event type
				if receiver.FilterEventType == evToDispatch.Type || receiver.FilterEventType == common_web.EVT_ANY {
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
	isRegistered    bool
	Ctx             context.Context
	Cancel          context.CancelFunc
	EventQueue      chan *pb.Event
	FilterEventType int64
}

func ConstructEventLog(source string, level int, message string) *pb.Event {

	tJson, _ := time.Now().MarshalJSON()

	return &pb.Event{
		Type: common_web.EVT_LOG,
		Values: []*pb.EventValue{
			{Val: &pb.EventValue_Tstring{Tstring: source}},
			{Val: &pb.EventValue_Tint64{Tint64: int64(level)}},
			{Val: &pb.EventValue_Tstring{Tstring: message}},
			{Val: &pb.EventValue_Tstring{Tstring: string(tJson)}},
		},
	}
}

// We add the Triggers to the oneof Values in proto later on (as they could carry arguments)
func ConstructEventTrigger(triggerType common_web.EvtTriggerType) *pb.Event {

	return &pb.Event{
		Type: common_web.EVT_TRIGGER,
		Values: []*pb.EventValue{
			&pb.EventValue{Val: &pb.EventValue_Tint64{Tint64: int64(triggerType)}},
		},
	}
}

func ConstructEventTriggerDHCPLease(iface, mac, ip string, hostname string) *pb.Event {
	return &pb.Event{
		Type: common_web.EVT_TRIGGER,
		Values: []*pb.EventValue{
			{Val: &pb.EventValue_Tint64{Tint64: int64(common_web.TRIGGER_EVT_TYPE_DHCP_LEASE_GRANTED)}},
			{Val: &pb.EventValue_Tstring{Tstring: iface}},
			{Val: &pb.EventValue_Tstring{Tstring: mac}},
			{Val: &pb.EventValue_Tstring{Tstring: ip}},
			{Val: &pb.EventValue_Tstring{Tstring: hostname}},
		},
	}
}

func ConstructEventTriggerSSHLogin(username string) *pb.Event {
	return &pb.Event{
		Type: common_web.EVT_TRIGGER,
		Values: []*pb.EventValue{
			{Val: &pb.EventValue_Tint64{Tint64: int64(common_web.TRIGGER_EVT_TYPE_SSH_LOGIN)}},
			{Val: &pb.EventValue_Tstring{Tstring: username}},
		},
	}
}

func ConstructEventTriggerGroupReceive(groupName string, value int32) *pb.Event {
	return &pb.Event{
		Type: common_web.EVT_TRIGGER,
		Values: []*pb.EventValue{
			{Val: &pb.EventValue_Tint64{Tint64: int64(common_web.TRIGGER_EVT_TYPE_GROUP_RECEIVE)}},
			{Val: &pb.EventValue_Tstring{Tstring: groupName}},
			{Val: &pb.EventValue_Tint64{Tint64: int64(value)}},
		},
	}
}

func DeconstructEventTriggerGroupReceive(evt *pb.Event) (groupName string, value int32, err error) {
	e := errors.New("Malformed GroupReceiveEvent")
	if evt.Type != common_web.EVT_TRIGGER {
		err = e
		return
	}
	if evTypeInt64,match := evt.Values[0].Val.(*pb.EventValue_Tint64); !match {
		err = e
		return
	} else {
		evType := common_web.EvtTriggerType(evTypeInt64.Tint64)
		if evType != common_web.TRIGGER_EVT_TYPE_GROUP_RECEIVE {
			err = e
			return
		}
	}

	groupName = evt.Values[1].GetTstring()
	value = int32(evt.Values[2].GetTint64())
	return
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
		resString, _ = job.ResultJsonString()
	}
	if eVM := hidEvent.Vm; eVM != nil {
		vmID = eVM.Id
	}

	tJson, _ := time.Now().MarshalJSON()

	return &pb.Event{
		Type: common_web.EVT_HID, //Type
		Values: []*pb.EventValue{
			{Val: &pb.EventValue_Tint64{Tint64: int64(hidEvent.Type)}}, //SubType = Type of hid.Event
			{Val: &pb.EventValue_Tint64{Tint64: int64(vmID)}},          //ID of VM
			{Val: &pb.EventValue_Tint64{Tint64: int64(jobID)}},         //ID of job
			{Val: &pb.EventValue_Tbool{Tbool: hasError}},               //isError (f.e. if a job was interrupted)
			{Val: &pb.EventValue_Tstring{Tstring: resString}},          //result String
			{Val: &pb.EventValue_Tstring{Tstring: errString}},          //error String (message in case of error)
			{Val: &pb.EventValue_Tstring{Tstring: message}},            //Mesage text of event
			{Val: &pb.EventValue_Tstring{Tstring: string(tJson)}},      //Timestamp of event genration
		},
	}
}
