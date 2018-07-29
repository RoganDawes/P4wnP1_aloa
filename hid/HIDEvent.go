package hid

type EventType int32

const (
	EventType_JOB_STARTED EventType = 0
	EventType_JOB_STOPPED EventType = 1
	EventType_CONTROLLER_ABORTED EventType = 2
	EventType_JOB_CANCELLED EventType = 3
	EventType_JOB_SUCCEEDED EventType = 4
	EventType_JOB_SUCCEEDED_NO_RESULT EventType = 5
	EventType_JOB_FAILED EventType = 6
	EventType_JOB_WAIT_LED_FINISHED EventType = 7
	EventType_JOB_WAIT_LED_REPEATED_FINISHED EventType = 8
	EventType_JOB_NO_FREE_VM EventType = 9

)

/*
var EventType_name = map[int32]string{
	0: "JOB_STARTED",
	1: "JOB_STOPPED",
	2: "CONTROLLER_ABORTED",
	3: "JOB_CANCELLED",

}
var EventType_value = map[string]int32{
	"JOB_STARTED": 0,
	"JOB_STOPPED": 1,
	"CONTROLLER_ABORTED": 2,
	"JOB_CANCELLED": 3,
}
*/

type Event struct {
	Type    EventType
	Job     *AsyncOttoJob
	Vm      *AsyncOttoVM
	Message string
	//ScriptSource string
}


type EventHandler interface {
	HandleEvent(event Event)
}
