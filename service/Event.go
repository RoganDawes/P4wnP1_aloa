package service

import (
	"context"
	"fmt"
	"github.com/mame82/P4wnP1_aloa/common_web"
	"github.com/mame82/P4wnP1_aloa/hid"
	pb "github.com/mame82/P4wnP1_aloa/proto"
	"log"
	"sync"
	"time"
	"errors"
)

type EventManager struct {
	eventQueue chan *pb.Event
	ctx        context.Context
	cancel     context.CancelFunc

	registeredReceiversMutex *sync.Mutex
	registeredReceivers  map[*EventReceiver]bool
	registerReceiver chan *EventReceiver
	unregisterReceiver chan *EventReceiver
}

func NewEventManager(queueSize int) *EventManager {
	EvMgr := &EventManager{
		eventQueue:           make(chan *pb.Event, queueSize),
		registeredReceivers:  make(map[*EventReceiver]bool),
		registerReceiver: make(chan *EventReceiver),
		unregisterReceiver: make(chan *EventReceiver),
		registeredReceiversMutex: &sync.Mutex{},
	}
	EvMgr.ctx, EvMgr.cancel = context.WithCancel(context.Background())
	return EvMgr
}

func (evm *EventManager) Start() {
	log.Println("Event Manager: Starting event dispatcher")
	go evm.register_unregister()
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
	ev := ConstructEventLog("logWriter", LOG_LEVEL_INFORMATION, string(p))
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
		waitRegister: make(chan struct{}),
	}



	em.registerReceiver <- er
	er.isRegistered = true

	// wait till registration is assured
	<-er.waitRegister

	go func() {
		<- er.Ctx.Done() //continue watching and assure unregister as soon as possible if canceled
		em.UnregisterReceiver(er)


	}()

	return er
}

func (em *EventManager) UnregisterReceiver(receiver *EventReceiver) {
	if !receiver.isRegistered {
		return
	}

	//mark as unregistered
	receiver.isRegistered = false
	em.unregisterReceiver <- receiver
}

func (em *EventManager) dispatch() {
	fmt.Println("Started event dispatcher")
loop:
	for {
		select {
		case evToDispatch := <-em.eventQueue:
			em.registeredReceiversMutex.Lock()
			for receiver := range em.registeredReceivers {
				// check if this receiver is listening for this event type
				if receiver != nil && receiver.isRegistered && (receiver.FilterEventType == evToDispatch.Type || receiver.FilterEventType == common_web.EVT_ANY) {
					receiver.EventQueue <- evToDispatch
				}
			}
			em.registeredReceiversMutex.Unlock()
		case <-em.ctx.Done():
			em.registeredReceiversMutex.Lock()
			for receiver := range em.registeredReceivers {
				// Calling unregister directly would dead lock on registeredReceiversMutex, as a buffer-less channel is
				// used and the receiver (register_unregister loop) locks registeredReceiversMutex, again
				// Calling cancel on the receiver itself, isn't a problem as the unregister is called by a dedicated go
				// routine per receiver, if the context is done.
				receiver.Cancel()
			}
			em.registeredReceiversMutex.Unlock()
			break loop
		}
	}
	fmt.Println("Stopped event dispatcher")
}

func (em *EventManager) register_unregister() {
	fmt.Println("Started event receiver (un)register watcher")
loop:
	for {
		select {
		case er := <- em.registerReceiver:  // Fix: this would already unlock the RegisterReceiver method ...
			em.registeredReceiversMutex.Lock()
			em.registeredReceivers[er] = true // ... but only at this point it is assured that the Listener receives events ...
			fmt.Printf("Registered event receiver type %d, overall receiver count %d\n", er.FilterEventType, len(em.registeredReceivers))
			// ... this is solved by signaling the successful registration by closing wait channel (the registerReceiver method doesn't return before this channel is closed)
			close(er.waitRegister)
			em.registeredReceiversMutex.Unlock()
		case er := <- em.unregisterReceiver:
			em.registeredReceiversMutex.Lock()
			delete(em.registeredReceivers, er)
			er.Cancel() // cancel context BEFORE closing the eventQueue channel
			close(er.EventQueue)
			fmt.Printf("Unregistered event receiver type %d, overall receiver count %d\n", er.FilterEventType, len(em.registeredReceivers))
			em.registeredReceiversMutex.Unlock()
		case <-em.ctx.Done():
			break loop
		}
	}
	fmt.Println("Stopped event receiver (un)register watcher")
}

type EventReceiver struct {
	waitRegister chan struct{}
	isRegistered    bool
	Ctx             context.Context
	Cancel          context.CancelFunc
	EventQueue      chan *pb.Event
	FilterEventType int64
}

func ConstructEventNotifyStateChange(stateType common_web.EvtStateChangeType) *pb.Event {
	return &pb.Event{
		Type: common_web.EVT_NOTIFY_STATE_CHANGE,
		Values: []*pb.EventValue{
			{Val: &pb.EventValue_Tint64{Tint64: int64(stateType)}},
		},
	}
}

/*
	case 1:
		return prefix + "critical"
	case 2:
		return prefix + "error"
	case 3:
		return prefix + "warning"
	case 4:
		return prefix + "information"
	case 5:
		return prefix + "verbose"
 */
type LogLevel int
const (
	LOG_LEVEL_UNDEFINED LogLevel = iota
	LOG_LEVEL_CRITICAL
	LOG_LEVEL_ERROR
	LOG_LEVEL_WARNING
	LOG_LEVEL_INFORMATION
	LOG_LEVEL_VERBOSE
)

func ConstructEventLog(source string, level LogLevel, message string) *pb.Event {
	//tJson, _ := time.Now().MarshalJSON()

	unixTimeMillis := time.Now().UnixNano() / 1e6

	return &pb.Event{
		Type: common_web.EVT_LOG,
		Values: []*pb.EventValue{
			{Val: &pb.EventValue_Tstring{Tstring: source}},
			{Val: &pb.EventValue_Tint64{Tint64: int64(level)}},
			{Val: &pb.EventValue_Tstring{Tstring: message}},
			{Val: &pb.EventValue_Tint64{Tint64: unixTimeMillis}}, //retrieve time in nano second accuracy and scale down to milliseconds
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

func ConstructEventTriggerGpioIn(gpioName string, level bool) *pb.Event {
	return &pb.Event{
		Type: common_web.EVT_TRIGGER,
		Values: []*pb.EventValue{
			{Val: &pb.EventValue_Tint64{Tint64: int64(common_web.TRIGGER_EVT_TYPE_GPIO_IN)}},
			{Val: &pb.EventValue_Tstring{Tstring: gpioName}},
			{Val: &pb.EventValue_Tbool{Tbool: level}},
		},
	}
}

func DeconstructEventTriggerGpioIn(evt *pb.Event) (gpioName string, level bool, err error) {
	e := errors.New("Malformed GpioEvent")
	if evt.Type != common_web.EVT_TRIGGER {
		err = e
		return
	}
	if evTypeInt64,match := evt.Values[0].Val.(*pb.EventValue_Tint64); !match {
		err = e
		return
	} else {
		evType := common_web.EvtTriggerType(evTypeInt64.Tint64)
		if evType != common_web.TRIGGER_EVT_TYPE_GPIO_IN {
			err = e
			return
		}
	}

	gpioName = evt.Values[1].GetTstring()
	level = evt.Values[2].GetTbool()
	return
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


	unixTimeMillis := time.Now().UnixNano() / 1e6

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
			{Val: &pb.EventValue_Tint64{Tint64: unixTimeMillis}},      //Timestamp of event genration
		},
	}
}
