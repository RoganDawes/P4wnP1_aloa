package common_web

const (
	EVT_ANY     = int64(0)
	EVT_LOG     = int64(1)
	EVT_HID     = int64(2)
	EVT_TRIGGER = int64(2)
)

type EvtTriggerType int64

const (
	EVT_TRIGGER_TYPE_SERVICE_STARTED         = EvtTriggerType(0)
	EVT_TRIGGER_TYPE_USB_GADGET_CONNECTED    = EvtTriggerType(1)
	EVT_TRIGGER_TYPE_USB_GADGET_DISCONNECTED = EvtTriggerType(2)
	EVT_TRIGGER_TYPE_WIFI_AP_STARTED         = EvtTriggerType(3)
	EVT_TRIGGER_TYPE_WIFI_CONNECTED_AS_STA   = EvtTriggerType(4)
	EVT_TRIGGER_TYPE_SSH_LOGIN               = EvtTriggerType(5)
	EVT_TRIGGER_TYPE_DHCP_LEASE_GRANTED      = EvtTriggerType(6)
)

const (
	HidEventType_JOB_STARTED                    = int64(0)
	HidEventType_JOB_STOPPED                    = int64(1)
	HidEventType_CONTROLLER_ABORTED             = int64(2)
	HidEventType_JOB_CANCELLED                  = int64(3)
	HidEventType_JOB_SUCCEEDED                  = int64(4)
	HidEventType_JOB_SUCCEEDED_NO_RESULT        = int64(5)
	HidEventType_JOB_FAILED                     = int64(6)
	HidEventType_JOB_WAIT_LED_FINISHED          = int64(7)
	HidEventType_JOB_WAIT_LED_REPEATED_FINISHED = int64(8)
	HidEventType_JOB_NO_FREE_VM                 = int64(9)
)

var EventType_name = map[int64]string{
	0: "JOB STARTED",
	1: "JOB STOPPED",
	2: "CONTROLLER ABORTED",
	3: "JOB CANCELLED",
	4: "JOB SUCCEEDED",
	5: "JOB SUCCEEDED WITHOUT RESULT",
	6: "JOB FAILED",
	7: "JOB WAIT LED FINISHED",
	8: "JOB WAIT LED REPEATED FINISHED",
	9: "JOB NO FREE VM",
}

/*
var EventType_value = map[string]int32{
	"JOB_STARTED": 0,
	"JOB_STOPPED": 1,
	"CONTROLLER_ABORTED": 2,
	"JOB_CANCELLED": 3,
}
*/
