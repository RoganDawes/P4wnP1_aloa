package common_web

const (
	EVT_ANY                 = int64(0)
	EVT_LOG                 = int64(1)
	EVT_HID                 = int64(3)
	EVT_TRIGGER             = int64(4)
	EVT_NOTIFY_STATE_CHANGE = int64(5) // fired if settings or a state changes (inform client about needed state update)
)

var EventTypeName = map[int64]string{
	EVT_TRIGGER:             "TRIGGER",
	EVT_LOG:                 "LOG",
	EVT_NOTIFY_STATE_CHANGE: "NOTIFY_STATE_CHANGE",
	EVT_HID:                 "HID",
}

type EvtTriggerType int64

const (
	TRIGGER_EVT_TYPE_SERVICE_STARTED         = EvtTriggerType(0)
	TRIGGER_EVT_TYPE_USB_GADGET_CONNECTED    = EvtTriggerType(1)
	TRIGGER_EVT_TYPE_USB_GADGET_DISCONNECTED = EvtTriggerType(2)
	TRIGGER_EVT_TYPE_WIFI_AP_STARTED         = EvtTriggerType(3)
	TRIGGER_EVT_TYPE_WIFI_CONNECTED_AS_STA   = EvtTriggerType(4)
	TRIGGER_EVT_TYPE_SSH_LOGIN               = EvtTriggerType(5)
	TRIGGER_EVT_TYPE_DHCP_LEASE_GRANTED      = EvtTriggerType(6)
	TRIGGER_EVT_TYPE_GPIO_IN                 = EvtTriggerType(7)
	TRIGGER_EVT_TYPE_GROUP_RECEIVE           = EvtTriggerType(8) //used for group receive and group receive sequence
)

type EvtStateChangeType int64

const (
	STATE_CHANGE_EVT_TYPE_USB             = EvtStateChangeType(0)
	STATE_CHANGE_EVT_TYPE_WIFI            = EvtStateChangeType(1)
	STATE_CHANGE_EVT_TYPE_NETWORK         = EvtStateChangeType(2)
	STATE_CHANGE_EVT_TYPE_BLUETOOTH       = EvtStateChangeType(3)
	STATE_CHANGE_EVT_TYPE_HID             = EvtStateChangeType(4)
	STATE_CHANGE_EVT_TYPE_TRIGGER_ACTIONS = EvtStateChangeType(5)
	STATE_CHANGE_EVT_TYPE_LED             = EvtStateChangeType(6)

	STATE_CHANGE_EVT_TYPE_STORED_HID_SCRIPTS_LIST                 = EvtStateChangeType(7)
	STATE_CHANGE_EVT_TYPE_STORED_USB_SETTINGS_LIST                = EvtStateChangeType(8)
	STATE_CHANGE_EVT_TYPE_STORED_ETHERNET_INTERFACE_SETTINGS_LIST = EvtStateChangeType(9)
	STATE_CHANGE_EVT_TYPE_STORED_WIFI_SETTINGS_LIST               = EvtStateChangeType(10)
	STATE_CHANGE_EVT_TYPE_STORED_BLUETOOTH_SETTINGS_LIST          = EvtStateChangeType(11)
	STATE_CHANGE_EVT_TYPE_STORED_TRIGGER_ACTION_SETS_LIST         = EvtStateChangeType(12)
	STATE_CHANGE_EVT_TYPE_STORED_BASH_SCRIPTS_LIST                = EvtStateChangeType(13)
	STATE_CHANGE_EVT_TYPE_STORED_GLOBAL_SETTINGS_LIST             = EvtStateChangeType(14)
)

var EventTypeStateChangeName = map[int64]string{
	int64(STATE_CHANGE_EVT_TYPE_USB):             "USB",
	int64(STATE_CHANGE_EVT_TYPE_WIFI):            "WIFI",
	int64(STATE_CHANGE_EVT_TYPE_NETWORK):         "NETWORK",
	int64(STATE_CHANGE_EVT_TYPE_BLUETOOTH):       "BLUETOOTH",
	int64(STATE_CHANGE_EVT_TYPE_HID):             "HID",
	int64(STATE_CHANGE_EVT_TYPE_TRIGGER_ACTIONS): "TRIGGER_ACTIONS",
	int64(STATE_CHANGE_EVT_TYPE_LED):             "LED",
}

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

var EventTypeHIDName = map[int64]string{
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
