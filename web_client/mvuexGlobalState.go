// +build js

package main

import (
	"context"
	"github.com/gopherjs/gopherjs/js"
	"github.com/mame82/P4wnP1_go/common_web"
	pb "github.com/mame82/P4wnP1_go/proto/gopherjs"
	"github.com/mame82/hvue"
	"github.com/mame82/mvuex"
	"github.com/pkg/errors"
	"io"
	"path/filepath"
	"strings"
	"time"
)

var globalState *GlobalState

const (
	maxLogEntries = 500

	VUEX_ACTION_UPDATE_ALL_STATES = "updateAllStates"

	//Events
	VUEX_ACTION_START_EVENT_LISTEN = "startEventListen"
	VUEX_ACTION_STOP_EVENT_LISTEN  = "stopEventListen"

	VUEX_MUTATION_SET_EVENT_LISTENER_RUNNING = "setEventListenerRunning"

	//Bluetooth
	VUEX_ACTION_UPDATE_CURRENT_BLUETOOTH_CONTROLLER_INFORMATION = "updateCurrentBluetoothControllerInformation"
	VUEX_ACTION_DEPLOY_CURRENT_BLUETOOTH_CONTROLLER_INFORMATION = "deployCurrentBluetoothControllerInformation"
	VUEX_ACTION_UPDATE_CURRENT_BLUETOOTH_AGENT_SETTINGS         = "updateCurrentBluetoothAgentSettings"
	VUEX_ACTION_DEPLOY_CURRENT_BLUETOOTH_AGENT_SETTINGS         = "deployCurrentBluetoothAgentSettings"
	VUEX_ACTION_STORE_BLUETOOTH_SETTINGS                        = "storedBluetoothSettings"
	VUEX_ACTION_DELETE_STORED_BLUETOOTH_SETTINGS                = "deleteStoredBluetoothSettings"
	VUEX_ACTION_DEPLOY_STORED_BLUETOOTH_SETTINGS                = "deployStoredBluetoothSettings"
	VUEX_ACTION_UPDATE_STORED_BLUETOOTH_SETTINGS_LIST           = "setStoredBluetoothSettingsList"

	VUEX_MUTATION_SET_CURRENT_BLUETOOTH_CONTROLLER_INFORMATION = "setCurrentBluetoothControllerInformation"
	VUEX_MUTATION_SET_CURRENT_BLUETOOTH_AGENT_SETTINGS         = "setCurrentBluetoothAgentSettings"
	VUEX_MUTATION_SET_STORED_BLUETOOTH_SETTINGS_LIST           = "setStoredBluetoothSettingsList"

	//HIDScripts and jobs
	VUEX_ACTION_UPDATE_RUNNING_HID_JOBS                           = "updateRunningHidJobs"
	VUEX_ACTION_REMOVE_SUCCEEDED_HID_JOBS                           = "removeSucceededHidJobs"
	VUEX_ACTION_REMOVE_FAILED_HID_JOBS                           = "removeFailedHidJobs"
	VUEX_ACTION_UPDATE_STORED_HID_SCRIPTS_LIST                    = "updateStoredHIDScriptsList"
	VUEX_ACTION_UPDATE_CURRENT_HID_SCRIPT_SOURCE_FROM_REMOTE_FILE = "updateCurrentHidScriptSourceFromRemoteFile"
	VUEX_ACTION_STORE_CURRENT_HID_SCRIPT_SOURCE_TO_REMOTE_FILE    = "storeCurrentHidScriptSourceToRemoteFile"
	VUEX_ACTION_CANCEL_HID_JOB = "cancelHIDJob"

	VUEX_MUTATION_SET_CURRENT_HID_SCRIPT_SOURCE_TO = "setCurrentHIDScriptSource"
	VUEX_MUTATION_SET_STORED_HID_SCRIPTS_LIST      = "setStoredHIDScriptsList"
	VUEX_MUTATION_DELETE_HID_JOB_ID      = "deleteHIDJobID"

	//USBGadget
	VUEX_ACTION_DEPLOY_CURRENT_USB_SETTINGS     = "deployCurrentUSBSettings"
	VUEX_ACTION_UPDATE_CURRENT_USB_SETTINGS     = "updateCurrentUSBSettings"
	VUEX_ACTION_STORE_USB_SETTINGS              = "storeUSBSettings"
	VUEX_ACTION_LOAD_USB_SETTINGS               = "loadUSBSettings"
	VUEX_ACTION_DEPLOY_STORED_USB_SETTINGS      = "deployStoredUSBSettings"
	VUEX_ACTION_UPDATE_STORED_USB_SETTINGS_LIST = "updateStoredUSBSettingsList"
	VUEX_ACTION_DELETE_STORED_USB_SETTINGS      = "deleteStoredUSBSettings"

	VUEX_MUTATION_SET_CURRENT_USB_SETTINGS     = "setCurrentUSBSettings"
	VUEX_MUTATION_SET_STORED_USB_SETTINGS_LIST = "setStoredUSBSettingsList"

	// Ethernet
	VUEX_ACTION_UPDATE_ALL_ETHERNET_INTERFACE_SETTINGS         = "updateAllEthernetInterfaceSettings"
	VUEX_ACTION_DEPLOY_ETHERNET_INTERFACE_SETTINGS             = "deployEthernetInterfaceSettings"
	VUEX_ACTION_STORE_ETHERNET_INTERFACE_SETTINGS              = "storeEthernetInterfaceSettings"
	VUEX_ACTION_LOAD_ETHERNET_INTERFACE_SETTINGS               = "loadEthernetInterfaceSettings"
	VUEX_ACTION_DEPLOY_STORED_ETHERNET_INTERFACE_SETTINGS      = "deployStoredEthernetInterfaceSettings"
	VUEX_ACTION_UPDATE_STORED_ETHERNET_INTERFACE_SETTINGS_LIST = "updateStoredEthernetInterfaceSettingsList"
	VUEX_ACTION_DELETE_STORED_ETHERNET_INTERFACE_SETTINGS      = "deleteStoredEthernetInterfaceSettings"

	VUEX_MUTATION_SET_STORED_ETHERNET_INTERFACE_SETTINGS_LIST = "setStoredEthernetInterfaceSettingsList"
	VUEX_MUTATION_SET_ALL_ETHERNET_INTERFACE_SETTINGS         = "setAllEthernetInterfaceSettings"
	VUEX_MUTATION_SET_SINGLE_ETHERNET_INTERFACE_SETTINGS      = "setSingleEthernetInterfaceSettings"

	//WiFi
	VUEX_ACTION_UPDATE_WIFI_STATE                = "updateCurrentWifiSettingsFromDeployed"
	VUEX_ACTION_DEPLOY_WIFI_SETTINGS             = "deployWifiSettings"
	VUEX_ACTION_UPDATE_STORED_WIFI_SETTINGS_LIST = "updateStoredWifiSettingsList"
	VUEX_ACTION_STORE_WIFI_SETTINGS              = "storeWifiSettings"
	VUEX_ACTION_LOAD_WIFI_SETTINGS               = "loadWifiSettings"
	VUEX_ACTION_DEPLOY_STORED_WIFI_SETTINGS      = "deployStoredWifiSettings"
	VUEX_ACTION_DELETE_STORED_WIFI_SETTINGS      = "deleteStoredWifiSettings"

	VUEX_MUTATION_SET_WIFI_STATE                = "setCurrentWifiState"
	VUEX_MUTATION_SET_STORED_WIFI_SETTINGS_LIST = "setStoredWifiSettingsList"
	VUEX_MUTATION_SET_CURRENT_WIFI_SETTINGS     = "setCurrentWifiSettings"

	//TriggerActions
	VUEX_ACTION_UPDATE_CURRENT_TRIGGER_ACTIONS_FROM_SERVER = "updateCurrentTriggerActionsFromServer"
	VUEX_ACTION_ADD_NEW_TRIGGER_ACTION                     = "addTriggerAction"
	VUEX_ACTION_REMOVE_TRIGGER_ACTIONS                     = "removeTriggerActions"
	VUEX_ACTION_STORE_TRIGGER_ACTION_SET                   = "storeTriggerActionSet"
	VUEX_ACTION_UPDATE_STORED_TRIGGER_ACTION_SETS_LIST     = "updateStoredTriggerActionSetsList"
	VUEX_ACTION_DEPLOY_STORED_TRIGGER_ACTION_SET_REPLACE   = "deployStoredTriggerActionSetReplace"
	VUEX_ACTION_DEPLOY_STORED_TRIGGER_ACTION_SET_ADD       = "deployStoredTriggerActionSetAdd"
	VUEX_ACTION_DELETE_STORED_TRIGGER_ACTION_SET           = "deleteStoredTriggerActionSet"
	VUEX_ACTION_DEPLOY_TRIGGER_ACTION_SET_REPLACE          = "deployCurrentTriggerActionSetReplace"
	VUEX_ACTION_DEPLOY_TRIGGER_ACTION_SET_ADD              = "deployCurrentTriggerActionSetAdd"

	VUEX_MUTATION_SET_STORED_TRIGGER_ACTIONS_SETS_LIST = "setStoredTriggerActionSetsList"

	//Bash scripts (used by TriggerActions)
	VUEX_ACTION_UPDATE_STORED_BASH_SCRIPTS_LIST = "updateStoredBashScriptsList"

	VUEX_MUTATION_SET_STORED_BASH_SCRIPTS_LIST = "setStoredBashScriptsList"

	defaultTimeoutShort = time.Second * 5
	defaultTimeout      = time.Second * 10
	defaultTimeoutMid   = time.Second * 30
)

type GlobalState struct {
	*js.Object
	Title                            string                   `js:"title"`
	CurrentHIDScriptSource           string                   `js:"currentHIDScriptSource"`
	CurrentGadgetSettings            *jsGadgetSettings        `js:"currentGadgetSettings"`
	CurrentlyDeployingGadgetSettings bool                     `js:"deployingGadgetSettings"`
	CurrentlyDeployingWifiSettings   bool                     `js:"deployingWifiSettings"`
	EventProcessor                   *jsEventProcessor        `js:"EventProcessor"`
	HidJobList                       *jsHidJobStateList       `js:"hidJobList"`
	TriggerActionList                *jsTriggerActionSet      `js:"triggerActionList"`
	IsModalEnabled                   bool                     `js:"isModalEnabled"`
	IsConnected                      bool                     `js:"isConnected"`
	FailedConnectionAttempts         int                      `js:"failedConnectionAttempts"`
	InterfaceSettings                *jsEthernetSettingsArray `js:"InterfaceSettings"`
	WiFiState                             *jsWiFiState                      `js:"wifiState"`
	CurrentBluetoothControllerInformation *jsBluetoothControllerInformation `js:"CurrentBluetoothControllerInformation"`
	CurrentBluetoothAgentSettings         *jsBluetoothAgentSettings         `js:"CurrentBluetoothAgentSettings"`

	StoredWifiSettingsList              []string `js:"StoredWifiSettingsList"`
	StoredEthernetInterfaceSettingsList []string `js:"StoredEthernetInterfaceSettingsList"`
	StoredTriggerActionSetsList         []string `js:"StoredTriggerActionSetsList"`
	StoredBashScriptsList               []string `js:"StoredBashScriptsList"`
	StoredHIDScriptsList                []string `js:"StoredHIDScriptsList"`
	StoredUSBSettingsList               []string `js:"StoredUSBSettingsList"`
	StoredBluetoothSettingsList         []string `js:"StoredBluetoothSettingsList"`

	ConnectRetryCount            int  `js:"ConnectRetryCount"`
	EventListenerRunning         bool `js:"EventListenerRunning"`
	EventListenerShouldBeRunning bool `js:"EventListenerShouldBeRunning"`
	EventListenerCancelFunc      context.CancelFunc // not accessible from JS !!!
}

func createGlobalStateStruct() GlobalState {
	state := GlobalState{Object: O()}
	state.Title = "P4wnP1 by MaMe82"
	state.CurrentHIDScriptSource = initHIDScript
	state.CurrentGadgetSettings = NewUSBGadgetSettings()
	state.CurrentlyDeployingWifiSettings = false
	state.HidJobList = NewHIDJobStateList()
	state.TriggerActionList = NewTriggerActionSet()
	state.EventProcessor = NewEventProcessor(maxLogEntries, state.HidJobList)
	state.IsConnected = false
	state.IsModalEnabled = false
	state.FailedConnectionAttempts = 0

	state.StoredWifiSettingsList = []string{}
	state.StoredEthernetInterfaceSettingsList = []string{}
	state.StoredTriggerActionSetsList = []string{}
	state.StoredBashScriptsList = []string{}
	state.StoredHIDScriptsList = []string{}
	state.StoredUSBSettingsList = []string{}
	state.StoredBluetoothSettingsList = []string{}

	//Retrieve Interface settings
	state.InterfaceSettings = NewEthernetSettingsList()
	state.CurrentBluetoothControllerInformation = NewBluetoothControllerInformation()
	state.CurrentBluetoothAgentSettings = NewBluetoothAgentSettings()

	//state.WiFiSettings = NewWifiSettings()
	state.WiFiState = NewWiFiState()

	//Events
	state.EventListenerRunning = false
	state.EventListenerShouldBeRunning = false

	state.ConnectRetryCount = 0
	return state
}

func processEvent(evt *pb.Event, store *mvuex.Store, state *GlobalState) {
	println("New event", evt)

	typeName := common_web.EventTypeName[evt.Type]
	switch evt.Type {
	case common_web.EVT_NOTIFY_STATE_CHANGE:
		chgType := evt.Values[0].GetTint64()
		println("State change notify", common_web.EventTypeStateChangeName[chgType])
		switch common_web.EvtStateChangeType(chgType) {
		case common_web.STATE_CHANGE_EVT_TYPE_LED:
			println("Notify LED change, nothing to do")
		case common_web.STATE_CHANGE_EVT_TYPE_USB:
			store.Dispatch(VUEX_ACTION_UPDATE_CURRENT_USB_SETTINGS)
			store.Dispatch(VUEX_ACTION_UPDATE_RUNNING_HID_JOBS)
		case common_web.STATE_CHANGE_EVT_TYPE_NETWORK:
			store.Dispatch(VUEX_ACTION_UPDATE_ALL_ETHERNET_INTERFACE_SETTINGS)
		case common_web.STATE_CHANGE_EVT_TYPE_HID:
			store.Dispatch(VUEX_ACTION_UPDATE_RUNNING_HID_JOBS) // handled by dedicated listener
		case common_web.STATE_CHANGE_EVT_TYPE_WIFI:
			store.Dispatch(VUEX_ACTION_UPDATE_WIFI_STATE)
		case common_web.STATE_CHANGE_EVT_TYPE_TRIGGER_ACTIONS:
			store.Dispatch(VUEX_ACTION_UPDATE_CURRENT_TRIGGER_ACTIONS_FROM_SERVER)
		case common_web.STATE_CHANGE_EVT_TYPE_BLUETOOTH:
			store.Dispatch(VUEX_ACTION_UPDATE_CURRENT_BLUETOOTH_CONTROLLER_INFORMATION)
			store.Dispatch(VUEX_ACTION_UPDATE_CURRENT_BLUETOOTH_AGENT_SETTINGS)
		case common_web.STATE_CHANGE_EVT_TYPE_STORED_BASH_SCRIPTS_LIST:
			store.Dispatch(VUEX_ACTION_UPDATE_STORED_BASH_SCRIPTS_LIST)
		case common_web.STATE_CHANGE_EVT_TYPE_STORED_HID_SCRIPTS_LIST:
			store.Dispatch(VUEX_ACTION_UPDATE_STORED_HID_SCRIPTS_LIST)
		case common_web.STATE_CHANGE_EVT_TYPE_STORED_USB_SETTINGS_LIST:
			store.Dispatch(VUEX_ACTION_UPDATE_STORED_USB_SETTINGS_LIST)
		case common_web.STATE_CHANGE_EVT_TYPE_STORED_ETHERNET_INTERFACE_SETTINGS_LIST:
			store.Dispatch(VUEX_ACTION_UPDATE_STORED_ETHERNET_INTERFACE_SETTINGS_LIST)
		case common_web.STATE_CHANGE_EVT_TYPE_STORED_WIFI_SETTINGS_LIST:
			store.Dispatch(VUEX_ACTION_UPDATE_STORED_WIFI_SETTINGS_LIST)
		case common_web.STATE_CHANGE_EVT_TYPE_STORED_BLUETOOTH_SETTINGS_LIST:
			store.Dispatch(VUEX_ACTION_UPDATE_STORED_BLUETOOTH_SETTINGS_LIST)
		case common_web.STATE_CHANGE_EVT_TYPE_STORED_TRIGGER_ACTION_SETS_LIST:
			store.Dispatch(VUEX_ACTION_UPDATE_STORED_TRIGGER_ACTION_SETS_LIST)
		}
	default:
		// events which aren't of type "notify state change" are processed by the EventProcessor
		state.EventProcessor.HandleEvent(evt)
		println("Unhandled event of type ", typeName, evt)
	}
}

func actionUpdateAllStates(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	println("Updating all states")
	store.Dispatch(VUEX_ACTION_UPDATE_CURRENT_USB_SETTINGS)
	store.Dispatch(VUEX_ACTION_UPDATE_RUNNING_HID_JOBS)
	store.Dispatch(VUEX_ACTION_UPDATE_ALL_ETHERNET_INTERFACE_SETTINGS)
	store.Dispatch(VUEX_ACTION_UPDATE_WIFI_STATE)
	store.Dispatch(VUEX_ACTION_UPDATE_CURRENT_BLUETOOTH_CONTROLLER_INFORMATION)
	store.Dispatch(VUEX_ACTION_UPDATE_CURRENT_BLUETOOTH_AGENT_SETTINGS)
	store.Dispatch(VUEX_ACTION_UPDATE_CURRENT_TRIGGER_ACTIONS_FROM_SERVER)

	store.Dispatch(VUEX_ACTION_UPDATE_STORED_HID_SCRIPTS_LIST)
	store.Dispatch(VUEX_ACTION_UPDATE_STORED_USB_SETTINGS_LIST)
	store.Dispatch(VUEX_ACTION_UPDATE_STORED_ETHERNET_INTERFACE_SETTINGS_LIST)
	store.Dispatch(VUEX_ACTION_UPDATE_STORED_WIFI_SETTINGS_LIST)
	store.Dispatch(VUEX_ACTION_UPDATE_STORED_BLUETOOTH_SETTINGS_LIST)
	store.Dispatch(VUEX_ACTION_UPDATE_STORED_TRIGGER_ACTION_SETS_LIST)
	store.Dispatch(VUEX_ACTION_UPDATE_STORED_BASH_SCRIPTS_LIST)

}

func actionCancelHidJob(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, jobID *js.Object)  {


	go func() {
		id := uint32(jobID.Int())
		println("Cancel HIDScript job", id)
		//fetch deployed gadget settings
		err := RpcClient.CancelHIDScriptJob(defaultTimeout, id)
		if err != nil {
			println("Couldn't cancel HIDScript job", err)
			return
		}

		// ToDo: update HIDScriptJob list (should be done event based)
	}()


	return
}

func actionRemoveSucceededHidJobs(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState)  {
	vJobs := state.HidJobList.Jobs                        //vue object, no real array --> values have to be extracted to filter
	jobs := js.Global.Get("Object").Call("values", vJobs) //converted to native JS array (has filter method available
	filtered := jobs.Call("filter", func(job *jsHidJobState) bool {
		return job.HasSucceeded
	})
	for i:=0; i< filtered.Length(); i++ {
		job := &jsHidJobState{Object: filtered.Index(i)}
		store.Commit(VUEX_MUTATION_DELETE_HID_JOB_ID, job.Id)
	}
	return
}

func actionRemoveFailedHidJobs(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState)  {
	vJobs := state.HidJobList.Jobs                        //vue object, no real array --> values have to be extracted to filter
	jobs := js.Global.Get("Object").Call("values", vJobs) //converted to native JS array (has filter method available
	filtered := jobs.Call("filter", func(job *jsHidJobState) bool {
		return job.HasFailed
	})
	for i:=0; i< filtered.Length(); i++ {
		job := &jsHidJobState{Object: filtered.Index(i)}
		store.Commit(VUEX_MUTATION_DELETE_HID_JOB_ID, job.Id)
	}
	return
}

func actionStartEventListen(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	globalState.EventListenerShouldBeRunning = true
	go func() {
		for state.EventListenerShouldBeRunning {
			println("Starting store event listener")

			evStream, cancel, err := RpcClient.StartEventListening(time.Second * 2)
			if err != nil {
				println("Couldn't start event listener or connect to server ...")
				//return
				if globalState.EventListenerShouldBeRunning {
					// unintended disconnect, sleep, increase retry count, reconnect
					state.ConnectRetryCount++
					time.Sleep(time.Second * 2) // We add an addional timeout, to avoid immediate reconnect if `startEventListening` fails without consuming the timeout
					continue
				} else {
					// set retry count to zero
					state.ConnectRetryCount = 0
					break
				}
			}
			println(" ... Event listener started")

			defer cancel()
			defer evStream.CloseSend()

			if state.EventListenerCancelFunc != nil {
				globalState.EventListenerCancelFunc() // cancel old event listeners
			}
			globalState.EventListenerCancelFunc = cancel
			//state.EventListenerRunning = true  // should be done with mutation
			store.Commit(VUEX_MUTATION_SET_EVENT_LISTENER_RUNNING, true)
			store.Dispatch(VUEX_ACTION_UPDATE_ALL_STATES)

			// dummy, retrieve and print events
			for {
				newEvent, err := evStream.Recv() //Error if Websocket connection fails/aborts, but success is indicated only if stream data is received
				if err != nil {
					if err == io.EOF {
						println("Event listening aborted because end of event stream was reached")
					} else {
						println("Event listening aborted because of error", err)
					}
					break
				}

				// only print event
				processEvent(newEvent, store, state)

			}

			println("Stopped store event listener")

			//state.EventListenerRunning = false // should be done with mutation
			store.Commit(VUEX_MUTATION_SET_EVENT_LISTENER_RUNNING, false)
			globalState.EventListenerCancelFunc = nil

		}

	}()

	return
}

func actionStopEventListen(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	// Note: this action accesses `globalState` instead of `state`
	// both values point to the vuex state, but globalState holds fields only accessible from Go
	// (like the `EventListenerCancelFunc` which is used here).
	// Even if those fields would be exposed to JS by tagging them , they couldn't be externalized/internalized
	// correctly by gopherjs, as there's no JavaScript representation of the underlying Go objects.

	println("Stopping event listener")
	globalState.EventListenerShouldBeRunning = false
	if globalState.EventListenerCancelFunc != nil {
		println("Calling event listener cancel func")
		globalState.EventListenerCancelFunc()
	}
}

func actionUpdateStoredBluetoothSettingsList(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {
		println("Trying to fetch bluetooth settings list")
		//fetch deployed gadget settings
		btsList, err := RpcClient.GetStoredBluetoothSettingsList(defaultTimeout)
		if err != nil {
			println("Couldn't retrieve BluetoothSettings list")
			return
		}

		//commit to current

		context.Commit(VUEX_MUTATION_SET_STORED_BLUETOOTH_SETTINGS_LIST, btsList)
		//context.Commit(VUEX_MUTATION_SET_STORED_WIFI_SETTINGS_LIST, []string{"test1", "test2"})
	}()

	return
}

func actionDeployStoredBluetoothSettings(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, settingsName *js.Object) {
	go func() {
		println("Vuex dispatch load Bluetooth settings: ", settingsName.String())
		// convert to Go type
		goBluetoothStettings, err := RpcClient.DeployStoredBluetoothSettings(defaultTimeoutMid, &pb.StringMessage{Msg: settingsName.String()})
		if err != nil {
			QuasarNotifyError("Error deploying stored Bluetooth Settings", err.Error(), QUASAR_NOTIFICATION_POSITION_BOTTOM)
		}
		QuasarNotifySuccess("New Bluetooth settings deployed", "", QUASAR_NOTIFICATION_POSITION_TOP)
		jsBluetoothSettings := NewBluetoothSettings()
		jsBluetoothSettings.fromGo(goBluetoothStettings)
		println("New bluetooth settings", jsBluetoothSettings)
		context.Commit(VUEX_MUTATION_SET_CURRENT_BLUETOOTH_AGENT_SETTINGS, jsBluetoothSettings.As)
		context.Commit(VUEX_MUTATION_SET_CURRENT_BLUETOOTH_CONTROLLER_INFORMATION, jsBluetoothSettings.Ci)
	}()
}

func actionDeleteStoredBluetoothSettings(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, settingsName *js.Object) {
	go func() {
		println("Vuex dispatch delete Bluetooth settings: ", settingsName.String())
		// convert to Go type
		err := RpcClient.DeleteStoredBluetoothSettings(defaultTimeout, &pb.StringMessage{Msg: settingsName.String()})
		if err != nil {
			QuasarNotifyError("Error deleting stored Bluetooth Settings", err.Error(), QUASAR_NOTIFICATION_POSITION_BOTTOM)
		}
		QuasarNotifySuccess("Bluetooth settings deleted", "", QUASAR_NOTIFICATION_POSITION_TOP)
		actionUpdateStoredBluetoothSettingsList(store, context, state)
	}()
}

func actionStoreBluetoothSettings(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, req *jsBluetoothRequestSettingsStorage) {
	go func() {
		println("Vuex dispatch store Bluetooth settings: ", req.TemplateName)
		// convert to Go type
		err := RpcClient.StoreBluetoothSettings(defaultTimeout, req.toGo())
		if err != nil {
			QuasarNotifyError("Error storing Bluetooth Settings", err.Error(), QUASAR_NOTIFICATION_POSITION_BOTTOM)
		}
		QuasarNotifySuccess("New Bluetooth settings stored", "", QUASAR_NOTIFICATION_POSITION_TOP)
	}()
}

func actionDeleteStoredUSBSettings(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, settingsName *js.Object) {
	go func() {
		println("Vuex dispatch delete USB settings: ", settingsName.String())
		// convert to Go type
		err := RpcClient.DeleteStoredUSBSettings(defaultTimeout, &pb.StringMessage{Msg: settingsName.String()})
		if err != nil {
			QuasarNotifyError("Error deleting stored USB Settings", err.Error(), QUASAR_NOTIFICATION_POSITION_BOTTOM)
		}
		QuasarNotifySuccess("USB settings deleted", "", QUASAR_NOTIFICATION_POSITION_TOP)
		actionUpdateStoredUSBSettingsList(store, context, state)
	}()
}

func actionDeleteStoredTriggerActionSet(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, jsName *js.Object) {
	go func() {
		name := jsName.String()
		println("Vuex delete stored TriggerActionSet: ", name)

		// convert to Go type
		msg := &pb.StringMessage{Msg: name}

		err := RpcClient.DeleteStoredTriggerActionsSet(defaultTimeout, msg)
		if err != nil {
			QuasarNotifyError("Error deleting TriggerActionSet from store", err.Error(), QUASAR_NOTIFICATION_POSITION_BOTTOM)
			return
		}
		QuasarNotifySuccess("Deleted TriggerActionSet from store", name, QUASAR_NOTIFICATION_POSITION_TOP)

		actionUpdateStoredTriggerActionSetsList(store, context, state)
	}()
}

func actionDeleteStoredWifiSettings(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, settingsName *js.Object) {
	go func() {
		println("Vuex dispatch delete WiFi settings: ", settingsName.String())
		// convert to Go type
		err := RpcClient.DeleteStoredWifiSettings(defaultTimeout, &pb.StringMessage{Msg: settingsName.String()})
		if err != nil {
			QuasarNotifyError("Error deleting stored WiFi Settings", err.Error(), QUASAR_NOTIFICATION_POSITION_BOTTOM)
		}
		QuasarNotifySuccess("Stored WiFi settings deleted", "", QUASAR_NOTIFICATION_POSITION_TOP)
		actionUpdateStoredWifiSettingsList(store, context, state)
	}()
}

func actionDeleteStoredEthernetInterfaceSettings(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, settingsName *js.Object) {
	go func() {
		println("Vuex dispatch delete stored ethernet interface settings: ", settingsName.String())
		// convert to Go type
		err := RpcClient.DeleteStoredEthernetInterfaceSettings(defaultTimeoutMid, &pb.StringMessage{Msg: settingsName.String()})
		if err != nil {
			QuasarNotifyError("Error deleting stored ethernet interface Settings", err.Error(), QUASAR_NOTIFICATION_POSITION_BOTTOM)
		}
		QuasarNotifySuccess("Stored ethernet interface settings deleted", "", QUASAR_NOTIFICATION_POSITION_TOP)

		//we update all modelview settings of vuex, to reflect the changes
		actionUpdateStoredEthernetInterfaceSettingsList(store, context, state)
	}()
}

func actionUpdateCurrentBluetoothControllerInformation(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {
		println("Trying to fetch bluetooth controller information")
		//fetch deployed gadget settings
		res, err := RpcClient.GetBluetoothControllerInformation(defaultTimeout)
		if err != nil {
			println("Couldn't retrieve BluetoothControllerInformation")
			return
		}

		println("Bluetooth Controller Info: ", res)
		context.Commit(VUEX_MUTATION_SET_CURRENT_BLUETOOTH_CONTROLLER_INFORMATION, res)
	}()

	return
}

func actionDeployCurrentBluetoothControllerInformation(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {
		println("Trying to deploy bluetooth controller information: ", state.CurrentBluetoothControllerInformation)
		//fetch deployed gadget settings
		res, err := RpcClient.DeployBluetoothControllerInformation(defaultTimeout, state.CurrentBluetoothControllerInformation)
		if err != nil {
			println("Couldn't deploy BluetoothControllerInformation", err)
			actionUpdateCurrentBluetoothControllerInformation(store, context, state)
			return
		}

		println("Bluetooth Controller Info after deploy: ", res)
		context.Commit(VUEX_MUTATION_SET_CURRENT_BLUETOOTH_CONTROLLER_INFORMATION, res)
	}()

	return
}

func actionUpdateCurrentBluetoothAgentSettings(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {
		println("Trying to fetch bluetooth agent settings")
		//fetch deployed gadget settings
		res, err := RpcClient.GetBluetoothAgentSettings(defaultTimeout)
		if err != nil {
			println("Couldn't retrieve AgentSettings")
			return
		}

		println("Bluetooth Controller Info: ", res)
		context.Commit(VUEX_MUTATION_SET_CURRENT_BLUETOOTH_AGENT_SETTINGS, res)
	}()

	return
}

func actionDeployCurrentBluetoothAgentSettings(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {
		println("Trying to deploy bluetooth agent settings: ", state.CurrentBluetoothAgentSettings)
		//fetch deployed gadget settings
		res, err := RpcClient.DeployBluetoothAgentSettings(defaultTimeout, state.CurrentBluetoothAgentSettings)
		if err != nil {
			println("Couldn't deploy agent settings", err)
			actionUpdateCurrentBluetoothAgentSettings(store, context, state)
			return
		}

		println("Bluetooth agent settings after deploy: ", res)
		context.Commit(VUEX_MUTATION_SET_CURRENT_BLUETOOTH_AGENT_SETTINGS, res)
	}()

	return
}

func actionUpdateStoredUSBSettingsList(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {
		println("Trying to fetch stored USB settings list")
		//fetch deployed gadget settings
		usbsList, err := RpcClient.GetStoredUSBSettingsList(defaultTimeout)
		if err != nil {
			println("Couldn't retrieve USB SettingsList")
			return
		}

		//commit to current
		println(usbsList)
		context.Commit(VUEX_MUTATION_SET_STORED_USB_SETTINGS_LIST, usbsList)
		//context.Commit(VUEX_MUTATION_SET_STORED_WIFI_SETTINGS_LIST, []string{"test1", "test2"})
	}()

	return
}

func actionStoreUSBSettings(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, req *jsUSBRequestSettingsStorage) {
	go func() {
		println("Vuex dispatch store USB settings: ", req.TemplateName)
		// convert to Go type
		err := RpcClient.StoreUSBSettings(defaultTimeout, req.toGo())
		if err != nil {
			QuasarNotifyError("Error storing USB Settings", err.Error(), QUASAR_NOTIFICATION_POSITION_BOTTOM)
		}
		QuasarNotifySuccess("New USB settings stored", "", QUASAR_NOTIFICATION_POSITION_TOP)
	}()
}

func actionLoadUSBSettings(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, settingsName *js.Object) {
	go func() {
		println("Vuex dispatch load USB settings: ", settingsName.String())
		// convert to Go type
		settings, err := RpcClient.GetStoredUSBSettings(defaultTimeout, &pb.StringMessage{Msg: settingsName.String()})
		if err != nil {
			QuasarNotifyError("Error fetching stored USB Settings", err.Error(), QUASAR_NOTIFICATION_POSITION_BOTTOM)
		}

		jsSettings := NewUSBGadgetSettings()
		jsSettings.fromGo(settings)
		context.Commit(VUEX_MUTATION_SET_CURRENT_USB_SETTINGS, jsSettings)

		QuasarNotifySuccess("New USB settings loaded", "", QUASAR_NOTIFICATION_POSITION_TOP)
	}()
}

func actionDeployStoredUSBSettings(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, settingsName *js.Object) {
	go func() {
		println("Vuex dispatch load USB settings: ", settingsName.String())
		// convert to Go type
		goUSBstettings, err := RpcClient.DeployStoredUSBSettings(defaultTimeoutMid, &pb.StringMessage{Msg: settingsName.String()})
		if err != nil {
			QuasarNotifyError("Error deploying stored USB Settings", err.Error(), QUASAR_NOTIFICATION_POSITION_BOTTOM)
		}
		QuasarNotifySuccess("New USB settings deployed", "", QUASAR_NOTIFICATION_POSITION_TOP)
		jsUSBSettings := NewUSBGadgetSettings()
		jsUSBSettings.fromGo(goUSBstettings)
		context.Commit(VUEX_MUTATION_SET_CURRENT_USB_SETTINGS, jsUSBSettings)
	}()
}

func actionUpdateAllEthernetInterfaceSettings(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) interface{} {

	return NewPromise(func() (res interface{}, err error) {
		println("Trying to fetch all deployed Ethernet Interface settings")
		ifSettings, err := RpcClient.GetAllDeployedEthernetInterfaceSettings(time.Second * 5)
		if err != nil {
			err = errors.New("Couldn't retrieve interface settings")
			println(err.Error())
			return
		}

		//commit to current
		context.Commit(VUEX_MUTATION_SET_ALL_ETHERNET_INTERFACE_SETTINGS, ifSettings)
		//return nil,errors.New("gone wrong")
		return
	})

}

func actionStoreEthernetInterfaceSettings(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, req *jsEthernetRequestSettingsStorage) {
	go func() {
		println("Vuex dispatch store ethernet interface settings '", req.TemplateName, "' for interface: ", req.Settings.Name)
		// convert to Go type
		err := RpcClient.StoreEthernetInterfaceSettings(defaultTimeout, req.toGo())
		if err != nil {
			QuasarNotifyError("Error storing ethernet interface Settings", err.Error(), QUASAR_NOTIFICATION_POSITION_BOTTOM)
		}
		QuasarNotifySuccess("New ethernet interface settings stored", "", QUASAR_NOTIFICATION_POSITION_TOP)
	}()
}

func actionLoadEthernetInterfaceSettings(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, settingsName *js.Object) {
	go func() {
		println("Vuex dispatch load ethernet interface settings: ", settingsName.String())
		// retrieve GO type EthernetInterfaceSettings
		settings, err := RpcClient.GetStoredEthernetInterfaceSettings(defaultTimeout, &pb.StringMessage{Msg: settingsName.String()})
		if err != nil {
			QuasarNotifyError("Error fetching stored ethernet interface Settings", err.Error(), QUASAR_NOTIFICATION_POSITION_BOTTOM)
		}

		// convert to JS representation, in order to forward to mutation
		jsSettings := &jsEthernetInterfaceSettings{Object: O()}
		jsSettings.fromGo(settings)
		context.Commit(VUEX_MUTATION_SET_SINGLE_ETHERNET_INTERFACE_SETTINGS, jsSettings)

		QuasarNotifySuccess("New ethernet interface settings loaded", "Interface: "+jsSettings.Name, QUASAR_NOTIFICATION_POSITION_TOP)
	}()
}

func actionDeployStoredEthernetInterfaceSettings(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, settingsName *js.Object) {
	go func() {
		println("Vuex dispatch deploy stored ethernet interface settings: ", settingsName.String())
		// convert to Go type
		err := RpcClient.DeployStoredEthernetInterfaceSettings(defaultTimeoutMid, &pb.StringMessage{Msg: settingsName.String()})
		if err != nil {
			QuasarNotifyError("Error deploying stored ethernet interface Settings", err.Error(), QUASAR_NOTIFICATION_POSITION_BOTTOM)
		}
		QuasarNotifySuccess("New ethernet interface settings deployed", "", QUASAR_NOTIFICATION_POSITION_TOP)

		//we update all modelview settings of vuex, to reflect the changes
		actionUpdateAllEthernetInterfaceSettings(store, context, state)
	}()
}

func actionUpdateStoredEthernetInterfaceSettingsList(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {
		println("Trying to fetch EthernetInterfaceSettings list")
		//fetch deployed gadget settings
		eisList, err := RpcClient.GetStoredEthernetInterfaceSettingsList(defaultTimeout)
		if err != nil {
			println("Couldn't retrieve EthernetInterfaceSettings list")
			return
		}

		println("Fetched list: ", eisList)

		context.Commit(VUEX_MUTATION_SET_STORED_ETHERNET_INTERFACE_SETTINGS_LIST, eisList)
	}()

	return
}

func actionUpdateCurrentHidScriptSourceFromRemoteFile(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, req *jsLoadHidScriptSourceReq) {
	go func() {
		println("Trying to update current hid script source from remote file ", req.FileName)

		content, err := RpcClient.DownloadFileToBytes(defaultTimeoutMid, req.FileName, pb.AccessibleFolder_HID_SCRIPTS)
		if err != nil {
			QuasarNotifyError("Couldn't load HIDScript source "+req.FileName, err.Error(), QUASAR_NOTIFICATION_POSITION_TOP)
			//println("err", err)
			return
		}

		newSource := string(content)
		switch req.Mode {
		case HID_SCRIPT_SOURCE_LOAD_MODE_APPEND:
			newSource = state.CurrentHIDScriptSource + "\n" + newSource
		case HID_SCRIPT_SOURCE_LOAD_MODE_PREPEND:
			newSource = newSource + "\n" + state.CurrentHIDScriptSource
		case HID_SCRIPT_SOURCE_LOAD_MODE_REPLACE:
		default:
		}

		context.Commit(VUEX_MUTATION_SET_CURRENT_HID_SCRIPT_SOURCE_TO, newSource)
	}()

	return
}

func actionStoreCurrentHidScriptSourceToRemoteFile(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, filename *js.Object) {
	go func() {
		fname := filename.String()
		ext := filepath.Ext(fname)
		if lext := strings.ToLower(ext); lext != ".js" && lext != ".javascript" {
			fname = fname + ".js"
		}

		println("Trying to store current hid script source to remote file ", fname)

		content := []byte(state.CurrentHIDScriptSource)
		err := RpcClient.UploadBytesToFile(defaultTimeoutMid, fname, pb.AccessibleFolder_HID_SCRIPTS, content, true)
		if err != nil {
			QuasarNotifyError("Couldn't store HIDScript source "+fname, err.Error(), QUASAR_NOTIFICATION_POSITION_TOP)
			//println("err", err)
			return
		}

	}()

	return
}

func actionUpdateStoredBashScriptsList(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {
		println("Trying to fetch stored BashScripts list")
		//fetch deployed gadget settings
		bsList, err := RpcClient.GetStoredBashScriptsList(defaultTimeout)
		if err != nil {
			println("Couldn't retrieve stored BashScripts list")
			return
		}

		//commit to current
		println(bsList)
		context.Commit(VUEX_MUTATION_SET_STORED_BASH_SCRIPTS_LIST, bsList)
	}()

	return
}

func actionUpdateStoredHIDScriptsList(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {
		println("Trying to fetch stored HIDScripts list")
		//fetch deployed gadget settings
		hidsList, err := RpcClient.GetStoredHIDScriptsList(defaultTimeout)
		if err != nil {
			println("Couldn't retrieve  stored HIDScripts list")
			return
		}

		//commit to current
		println(hidsList)
		context.Commit(VUEX_MUTATION_SET_STORED_HID_SCRIPTS_LIST, hidsList)
	}()

	return
}

func actionUpdateGadgetSettingsFromDeployed(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {
		println("Trying to fetch deployed USB gadget settings")
		//fetch deployed gadget settings
		dGS, err := RpcClient.GetDeployedGadgetSettings(defaultTimeoutShort)
		if err != nil {
			println("Couldn't retrieve deployed gadget settings")
			return
		}
		//convert to JS version
		jsGS := &jsGadgetSettings{Object: O()}
		jsGS.fromGo(dGS)

		//commit to current
		context.Commit("setCurrentUSBSettings", jsGS)
	}()

	return
}

func actionUpdateWifiState(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {
		println("Trying to fetch deployed WiFi state list")
		//fetch WiFi state
		state, err := RpcClient.GetWifiState(defaultTimeout)
		if err != nil {
			println("Couldn't retrieve deployed WiFi settings")
			return
		}

		//commit to current
		context.Commit(VUEX_MUTATION_SET_WIFI_STATE, state)
	}()

	return
}

func actionUpdateStoredWifiSettingsList(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {
		println("Trying to fetch stored wifi settings list")
		//fetch deployed gadget settings
		wsList, err := RpcClient.GetStoredWifiSettingsList(defaultTimeout)
		if err != nil {
			println("Couldn't retrieve WifiSettingsList")
			return
		}

		//commit to current
		println(wsList)
		context.Commit(VUEX_MUTATION_SET_STORED_WIFI_SETTINGS_LIST, wsList)
		//context.Commit(VUEX_MUTATION_SET_STORED_WIFI_SETTINGS_LIST, []string{"test1", "test2"})
	}()

	return
}

func actionStoreWifiSettings(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, req *jsWifiRequestSettingsStorage) {
	go func() {
		println("Vuex dispatch store WiFi settings: ", req.TemplateName)
		// convert to Go type
		err := RpcClient.StoreWifiSettings(defaultTimeout, req.toGo())
		if err != nil {
			QuasarNotifyError("Error storing WiFi Settings", err.Error(), QUASAR_NOTIFICATION_POSITION_BOTTOM)
		}
		QuasarNotifySuccess("New WiFi settings stored", "", QUASAR_NOTIFICATION_POSITION_TOP)
	}()
}

func actionLoadWifiSettings(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, settingsName *js.Object) {
	go func() {
		println("Vuex dispatch load WiFi settings: ", settingsName.String())
		// convert to Go type
		settings, err := RpcClient.GetStoredWifiSettings(defaultTimeout, &pb.StringMessage{Msg: settingsName.String()})
		if err != nil {
			QuasarNotifyError("Error fetching stored WiFi Settings", err.Error(), QUASAR_NOTIFICATION_POSITION_BOTTOM)
		}

		jsSettings := NewWifiSettings()
		jsSettings.fromGo(settings)
		context.Commit(VUEX_MUTATION_SET_CURRENT_WIFI_SETTINGS, jsSettings)

		QuasarNotifySuccess("New WiFi settings loaded", "", QUASAR_NOTIFICATION_POSITION_TOP)
	}()
}

func actionDeployStoredWifiSettings(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, settingsName *js.Object) {
	go func() {
		println("Vuex dispatch deploy stored WiFi settings: ", settingsName.String())
		// convert to Go type
		wstate, err := RpcClient.DeployStoredWifiSettings(defaultTimeoutMid, &pb.StringMessage{Msg: settingsName.String()})
		if err != nil {
			QuasarNotifyError("Error deploying stored WiFi Settings", err.Error(), QUASAR_NOTIFICATION_POSITION_BOTTOM)
		}
		QuasarNotifySuccess("New WiFi settings deployed", "", QUASAR_NOTIFICATION_POSITION_TOP)
		state.WiFiState.fromGo(wstate)
	}()
}

func actionDeployWifiSettings(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, settings *jsWiFiSettings) {
	go func() {

		state.CurrentlyDeployingWifiSettings = true
		defer func() { state.CurrentlyDeployingWifiSettings = false }()

		println("Vuex dispatch deploy WiFi settings")
		// convert to Go type
		goSettings := settings.toGo()

		wstate, err := RpcClient.DeployWifiSettings(defaultTimeoutMid, goSettings)
		if err != nil {
			QuasarNotifyError("Error deploying WiFi Settings", err.Error(), QUASAR_NOTIFICATION_POSITION_BOTTOM)
		}
		QuasarNotifySuccess("New WiFi settings deployed", "", QUASAR_NOTIFICATION_POSITION_TOP)
		state.WiFiState.fromGo(wstate)
	}()
}

func actionUpdateRunningHidJobs(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {
		println("Trying to fetch data for running HIDScript jobs")
		//fetch deployed gadget settings
		jobstates, err := RpcClient.GetRunningHidJobStates(defaultTimeout)
		if err != nil {
			println("Couldn't retrieve stateof running HID jobs", err)
			return
		}

		state.HidJobList.Clear()

		for _, jobstate := range jobstates {
			println("updateing jobstate", jobstate)
			state.HidJobList.UpdateEntry(jobstate.Id, jobstate.VmId, false, false, "initial job state", "", time.Now().String(), jobstate.Source)
		}
	}()

	return
}

func actionUpdateStoredTriggerActionSetsList(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {
		println("Trying to fetch TriggerActionSetList")
		//fetch deployed gadget settings
		tasList, err := RpcClient.ListStoredTriggerActionSets(defaultTimeout)
		if err != nil {
			println("Couldn't retrieve TriggerActions list")
			return
		}

		//commit to current

		context.Commit(VUEX_MUTATION_SET_STORED_TRIGGER_ACTIONS_SETS_LIST, tasList)
		//context.Commit(VUEX_MUTATION_SET_STORED_WIFI_SETTINGS_LIST, []string{"test1", "test2"})
	}()

	return
}

func actionUpdateCurrentTriggerActionsFromServer(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {
		println("Trying to fetch current TriggerActions from server")
		tastate, err := RpcClient.GetDeployedTriggerActionSet(defaultTimeout)
		if err != nil {
			QuasarNotifyError("Error fetching deployed TriggerActions", err.Error(), QUASAR_NOTIFICATION_POSITION_TOP)
			return
		}

		// ToDo: Clear list berfore adding back elements
		state.TriggerActionList.Flush()

		for _, ta := range tastate.TriggerActions {

			jsTA := NewTriggerAction()
			jsTA.fromGo(ta)
			state.TriggerActionList.UpdateEntry(jsTA)
		}
	}()

	return
}

func actionAddNewTriggerAction(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {
		newTA := NewTriggerAction()
		newTA.IsActive = false // don't activate by default
		RpcClient.DeployTriggerActionsSetAdd(defaultTimeout, &pb.TriggerActionSet{TriggerActions: []*pb.TriggerAction{newTA.toGo()}})

		actionUpdateCurrentTriggerActionsFromServer(store, context, state)
	}()

	return
}

func actionRemoveTriggerActions(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, jsTas *jsTriggerActionSet) {
	go func() {
		RpcClient.DeployTriggerActionsSetRemove(defaultTimeout, jsTas.toGo())

		actionUpdateCurrentTriggerActionsFromServer(store, context, state)
	}()

	return
}

func actionStoreTriggerActionSet(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, req *jsTriggerActionSet) {
	go func() {
		println("Vuex dispatch store TriggerAction list: ", req.Name)
		tas := req.toGo()

		err := RpcClient.StoreTriggerActionSet(defaultTimeout, tas)
		if err != nil {
			QuasarNotifyError("Error storing TriggerActionSet", err.Error(), QUASAR_NOTIFICATION_POSITION_BOTTOM)
			return
		}
		QuasarNotifySuccess("TriggerActionSet stored", "", QUASAR_NOTIFICATION_POSITION_TOP)

	}()
}

func actionDeployTriggerActionSetReplace(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, tasToDeploy *jsTriggerActionSet) {
	go func() {
		tas := tasToDeploy.toGo()

		_, err := RpcClient.DeployTriggerActionsSetReplace(defaultTimeout, tas)
		if err != nil {
			QuasarNotifyError("Error replacing TriggerActionSet with given one", err.Error(), QUASAR_NOTIFICATION_POSITION_BOTTOM)
			return
		}
		QuasarNotifySuccess("Replaced TriggerActionSet with given one", "", QUASAR_NOTIFICATION_POSITION_TOP)

		actionUpdateCurrentTriggerActionsFromServer(store, context, state)
	}()
}

func actionDeployTriggerActionSetAdd(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, tasToDeploy *jsTriggerActionSet) {
	go func() {
		tas := tasToDeploy.toGo()
		_, err := RpcClient.DeployTriggerActionsSetAdd(defaultTimeout, tas)
		if err != nil {
			QuasarNotifyError("Error adding given TriggerActionSet to server", err.Error(), QUASAR_NOTIFICATION_POSITION_BOTTOM)
			return
		}
		QuasarNotifySuccess("Added TriggerActionSet to server", "", QUASAR_NOTIFICATION_POSITION_TOP)

		actionUpdateCurrentTriggerActionsFromServer(store, context, state)
	}()
}

func actionDeployStoredTriggerActionSetReplace(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, jsName *js.Object) {
	go func() {
		name := jsName.String()
		println("Vuex dispatch deploy stored TriggerActionSet as replacement: ", name)

		// convert to Go type
		msg := &pb.StringMessage{Msg: name}

		_, err := RpcClient.DeployStoredTriggerActionsSetReplace(defaultTimeout, msg)
		if err != nil {
			QuasarNotifyError("Error replacing TriggerActionSet with stored set", err.Error(), QUASAR_NOTIFICATION_POSITION_BOTTOM)
			return
		}
		QuasarNotifySuccess("Replaced TriggerActionSet by stored set", name, QUASAR_NOTIFICATION_POSITION_TOP)

		actionUpdateCurrentTriggerActionsFromServer(store, context, state)
	}()
}

func actionDeployStoredTriggerActionSetAdd(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, jsName *js.Object) {
	go func() {
		name := jsName.String()
		println("Vuex dispatch deploy stored TriggerActionSet as addition: ", name)

		// convert to Go type
		msg := &pb.StringMessage{Msg: name}

		_, err := RpcClient.DeployStoredTriggerActionsSetAdd(defaultTimeout, msg)
		if err != nil {
			QuasarNotifyError("Error adding TriggerActionSet from store", err.Error(), QUASAR_NOTIFICATION_POSITION_BOTTOM)
			return
		}
		QuasarNotifySuccess("Added TriggerActionSet from store", name, QUASAR_NOTIFICATION_POSITION_TOP)

		actionUpdateCurrentTriggerActionsFromServer(store, context, state)
	}()
}

func actionDeployCurrentGadgetSettings(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {

		state.CurrentlyDeployingGadgetSettings = true
		defer func() { state.CurrentlyDeployingGadgetSettings = false }()

		//get current GadgetSettings
		curGS := state.CurrentGadgetSettings.toGo()

		//try to set them via gRPC (the server holds an internal state, setting != deploying)
		err := RpcClient.SetRemoteGadgetSettings(curGS, defaultTimeoutShort)
		if err != nil {
			QuasarNotifyError("Error in pre-check of new USB gadget settings", err.Error(), QUASAR_NOTIFICATION_POSITION_TOP)
			return
		}

		//try to deploy the, now set, remote GadgetSettings via gRPC
		_, err = RpcClient.DeployRemoteGadgetSettings(defaultTimeout)
		if err != nil {
			QuasarNotifyError("Error while deploying new USB gadget settings", err.Error(), QUASAR_NOTIFICATION_POSITION_TOP)
			return
		}

		notification := &QuasarNotification{Object: O()}
		notification.Message = "New Gadget Settings deployed successfully"
		notification.Position = QUASAR_NOTIFICATION_POSITION_TOP
		notification.Type = QUASAR_NOTIFICATION_TYPE_POSITIVE
		notification.Timeout = 2000
		QuasarNotify(notification)

	}()

	return
}

func actionDeployEthernetInterfaceSettings(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, settings *jsEthernetInterfaceSettings) {
	go func() {
		println("Vuex dispatch deploy ethernet interface settings")
		// convert to Go type
		goSettings := settings.toGo()

		err := RpcClient.DeployedEthernetInterfaceSettings(defaultTimeoutShort, goSettings)
		if err != nil {
			Alert(err)
		}
	}()
}

func initMVuex() *mvuex.Store {
	state := createGlobalStateStruct()
	globalState = &state //make accessible through global var (things like mutexes aren't accessible from JS and thus not externalized/internalized)
	store := mvuex.NewStore(
		mvuex.State(state),
		mvuex.Mutation("setModalEnabled", func(store *mvuex.Store, state *GlobalState, enabled bool) {
			state.IsModalEnabled = enabled
			return
		}),
		mvuex.Mutation(VUEX_MUTATION_SET_CURRENT_HID_SCRIPT_SOURCE_TO, func(store *mvuex.Store, state *GlobalState, newText string) {
			state.CurrentHIDScriptSource = newText
			return
		}),
		mvuex.Mutation(VUEX_MUTATION_SET_CURRENT_USB_SETTINGS, func(store *mvuex.Store, state *GlobalState, settings *jsGadgetSettings) {
			state.CurrentGadgetSettings = settings
			return
		}),
		mvuex.Mutation(VUEX_MUTATION_SET_WIFI_STATE, func(store *mvuex.Store, state *GlobalState, wifiState *jsWiFiState) {
			state.WiFiState = wifiState
			return
		}),
		mvuex.Mutation(VUEX_MUTATION_SET_CURRENT_WIFI_SETTINGS, func(store *mvuex.Store, state *GlobalState, wifiSettings *jsWiFiSettings) {
			state.WiFiState.CurrentSettings = wifiSettings
			return
		}),
		mvuex.Mutation(VUEX_MUTATION_SET_STORED_WIFI_SETTINGS_LIST, func(store *mvuex.Store, state *GlobalState, wsList []interface{}) {
			println("New ws list", wsList)
			hvue.Set(state, "StoredWifiSettingsList", wsList)
		}),
		mvuex.Mutation(VUEX_MUTATION_SET_STORED_BLUETOOTH_SETTINGS_LIST, func(store *mvuex.Store, state *GlobalState, btsList []interface{}) {
			println("New Bluetooth list", btsList)
			hvue.Set(state, "StoredBluetoothSettingsList", btsList)
		}),
		mvuex.Mutation(VUEX_MUTATION_SET_STORED_USB_SETTINGS_LIST, func(store *mvuex.Store, state *GlobalState, usbList []interface{}) {
			println("New USB settings list", usbList)
			hvue.Set(state, "StoredUSBSettingsList", usbList)
		}),
		mvuex.Mutation(VUEX_MUTATION_SET_STORED_BASH_SCRIPTS_LIST, func(store *mvuex.Store, state *GlobalState, bsList []interface{}) {
			hvue.Set(state, "StoredBashScriptsList", bsList)
		}),
		mvuex.Mutation(VUEX_MUTATION_SET_STORED_HID_SCRIPTS_LIST, func(store *mvuex.Store, state *GlobalState, hidsList []interface{}) {
			hvue.Set(state, "StoredHIDScriptsList", hidsList)
		}),
		mvuex.Mutation(VUEX_MUTATION_SET_STORED_TRIGGER_ACTIONS_SETS_LIST, func(store *mvuex.Store, state *GlobalState, tasList []interface{}) {
			hvue.Set(state, "StoredTriggerActionSetsList", tasList)
		}),
		mvuex.Mutation(VUEX_MUTATION_SET_STORED_ETHERNET_INTERFACE_SETTINGS_LIST, func(store *mvuex.Store, state *GlobalState, eisList []interface{}) {
			hvue.Set(state, "StoredEthernetInterfaceSettingsList", eisList)
		}),
		mvuex.Mutation(VUEX_MUTATION_SET_ALL_ETHERNET_INTERFACE_SETTINGS, func(store *mvuex.Store, state *GlobalState, eifSettings *jsEthernetSettingsArray) {
			println("Updating all ethernet interface settings: ", eifSettings)
			state.InterfaceSettings = eifSettings
		}),
		mvuex.Mutation(VUEX_MUTATION_SET_SINGLE_ETHERNET_INTERFACE_SETTINGS, func(store *mvuex.Store, state *GlobalState, ifSettings *jsEthernetInterfaceSettings) {
			println("Updating ethernet interface settings for ", ifSettings.Name)
			state.InterfaceSettings.updateSingleInterface(ifSettings)
		}),
		mvuex.Mutation(VUEX_MUTATION_SET_CURRENT_BLUETOOTH_CONTROLLER_INFORMATION, func(store *mvuex.Store, state *GlobalState, btCtlInfo *jsBluetoothControllerInformation) {
			println("Updating bluetooth controller information for ", btCtlInfo.Name)
			state.CurrentBluetoothControllerInformation = btCtlInfo
		}),
		mvuex.Mutation(VUEX_MUTATION_SET_CURRENT_BLUETOOTH_AGENT_SETTINGS, func(store *mvuex.Store, state *GlobalState, agentSettings *jsBluetoothAgentSettings) {
			println("Updating bluetooth agent settings for ", agentSettings)
			state.CurrentBluetoothAgentSettings = agentSettings
		}),
		mvuex.Mutation(VUEX_MUTATION_SET_EVENT_LISTENER_RUNNING, func(store *mvuex.Store, state *GlobalState, running *js.Object) {
			state.EventListenerRunning = running.Bool()
		}),
		mvuex.Mutation(VUEX_MUTATION_DELETE_HID_JOB_ID, func(store *mvuex.Store, state *GlobalState, jobID *js.Object) {
			id := jobID.Int()
			state.HidJobList.DeleteEntry(int64(id))

		}),


		mvuex.Action(VUEX_ACTION_UPDATE_ALL_STATES, actionUpdateAllStates),

		mvuex.Action(VUEX_ACTION_UPDATE_CURRENT_BLUETOOTH_CONTROLLER_INFORMATION, actionUpdateCurrentBluetoothControllerInformation),
		mvuex.Action(VUEX_ACTION_DEPLOY_CURRENT_BLUETOOTH_CONTROLLER_INFORMATION, actionDeployCurrentBluetoothControllerInformation),
		mvuex.Action(VUEX_ACTION_UPDATE_CURRENT_BLUETOOTH_AGENT_SETTINGS, actionUpdateCurrentBluetoothAgentSettings),
		mvuex.Action(VUEX_ACTION_DEPLOY_CURRENT_BLUETOOTH_AGENT_SETTINGS, actionDeployCurrentBluetoothAgentSettings),
		mvuex.Action(VUEX_ACTION_STORE_BLUETOOTH_SETTINGS, actionStoreBluetoothSettings),
		mvuex.Action(VUEX_ACTION_DELETE_STORED_BLUETOOTH_SETTINGS, actionDeleteStoredBluetoothSettings),
		mvuex.Action(VUEX_ACTION_DEPLOY_STORED_BLUETOOTH_SETTINGS, actionDeployStoredBluetoothSettings),
		mvuex.Action(VUEX_ACTION_UPDATE_STORED_BLUETOOTH_SETTINGS_LIST, actionUpdateStoredBluetoothSettingsList),


		mvuex.Action(VUEX_ACTION_UPDATE_CURRENT_USB_SETTINGS, actionUpdateGadgetSettingsFromDeployed),
		mvuex.Action(VUEX_ACTION_DEPLOY_CURRENT_USB_SETTINGS, actionDeployCurrentGadgetSettings),
		mvuex.Action(VUEX_ACTION_UPDATE_RUNNING_HID_JOBS, actionUpdateRunningHidJobs),
		mvuex.Action(VUEX_ACTION_STORE_USB_SETTINGS, actionStoreUSBSettings),
		mvuex.Action(VUEX_ACTION_LOAD_USB_SETTINGS, actionLoadUSBSettings),
		mvuex.Action(VUEX_ACTION_DEPLOY_STORED_USB_SETTINGS, actionDeployStoredUSBSettings),
		mvuex.Action(VUEX_ACTION_DELETE_STORED_USB_SETTINGS, actionDeleteStoredUSBSettings),
		mvuex.Action(VUEX_ACTION_UPDATE_STORED_USB_SETTINGS_LIST, actionUpdateStoredUSBSettingsList),

		mvuex.Action(VUEX_ACTION_DEPLOY_ETHERNET_INTERFACE_SETTINGS, actionDeployEthernetInterfaceSettings),
		mvuex.Action(VUEX_ACTION_UPDATE_ALL_ETHERNET_INTERFACE_SETTINGS, actionUpdateAllEthernetInterfaceSettings),

		mvuex.Action(VUEX_ACTION_UPDATE_WIFI_STATE, actionUpdateWifiState),
		mvuex.Action(VUEX_ACTION_DEPLOY_WIFI_SETTINGS, actionDeployWifiSettings),
		mvuex.Action(VUEX_ACTION_UPDATE_STORED_WIFI_SETTINGS_LIST, actionUpdateStoredWifiSettingsList),

		mvuex.Action(VUEX_ACTION_UPDATE_CURRENT_TRIGGER_ACTIONS_FROM_SERVER, actionUpdateCurrentTriggerActionsFromServer),
		mvuex.Action(VUEX_ACTION_ADD_NEW_TRIGGER_ACTION, actionAddNewTriggerAction),
		mvuex.Action(VUEX_ACTION_REMOVE_TRIGGER_ACTIONS, actionRemoveTriggerActions),
		mvuex.Action(VUEX_ACTION_STORE_TRIGGER_ACTION_SET, actionStoreTriggerActionSet),
		mvuex.Action(VUEX_ACTION_UPDATE_STORED_TRIGGER_ACTION_SETS_LIST, actionUpdateStoredTriggerActionSetsList),
		mvuex.Action(VUEX_ACTION_DEPLOY_STORED_TRIGGER_ACTION_SET_REPLACE, actionDeployStoredTriggerActionSetReplace),
		mvuex.Action(VUEX_ACTION_DEPLOY_STORED_TRIGGER_ACTION_SET_ADD, actionDeployStoredTriggerActionSetAdd),
		mvuex.Action(VUEX_ACTION_DELETE_STORED_TRIGGER_ACTION_SET, actionDeleteStoredTriggerActionSet),
		mvuex.Action(VUEX_ACTION_DEPLOY_TRIGGER_ACTION_SET_REPLACE, actionDeployTriggerActionSetReplace),
		mvuex.Action(VUEX_ACTION_DEPLOY_TRIGGER_ACTION_SET_ADD, actionDeployTriggerActionSetAdd),

		mvuex.Action(VUEX_ACTION_UPDATE_STORED_BASH_SCRIPTS_LIST, actionUpdateStoredBashScriptsList),
		mvuex.Action(VUEX_ACTION_UPDATE_STORED_HID_SCRIPTS_LIST, actionUpdateStoredHIDScriptsList),
		mvuex.Action(VUEX_ACTION_UPDATE_STORED_ETHERNET_INTERFACE_SETTINGS_LIST, actionUpdateStoredEthernetInterfaceSettingsList),
		mvuex.Action(VUEX_ACTION_UPDATE_CURRENT_HID_SCRIPT_SOURCE_FROM_REMOTE_FILE, actionUpdateCurrentHidScriptSourceFromRemoteFile),
		mvuex.Action(VUEX_ACTION_STORE_CURRENT_HID_SCRIPT_SOURCE_TO_REMOTE_FILE, actionStoreCurrentHidScriptSourceToRemoteFile),

		mvuex.Action(VUEX_ACTION_STORE_WIFI_SETTINGS, actionStoreWifiSettings),
		mvuex.Action(VUEX_ACTION_LOAD_WIFI_SETTINGS, actionLoadWifiSettings),
		mvuex.Action(VUEX_ACTION_DEPLOY_STORED_WIFI_SETTINGS, actionDeployStoredWifiSettings),
		mvuex.Action(VUEX_ACTION_DELETE_STORED_WIFI_SETTINGS, actionDeleteStoredWifiSettings),

		mvuex.Action(VUEX_ACTION_STORE_ETHERNET_INTERFACE_SETTINGS, actionStoreEthernetInterfaceSettings),
		mvuex.Action(VUEX_ACTION_LOAD_ETHERNET_INTERFACE_SETTINGS, actionLoadEthernetInterfaceSettings),
		mvuex.Action(VUEX_ACTION_DEPLOY_STORED_ETHERNET_INTERFACE_SETTINGS, actionDeployStoredEthernetInterfaceSettings),
		mvuex.Action(VUEX_ACTION_DELETE_STORED_ETHERNET_INTERFACE_SETTINGS, actionDeleteStoredEthernetInterfaceSettings),


		mvuex.Action(VUEX_ACTION_START_EVENT_LISTEN, actionStartEventListen),
		mvuex.Action(VUEX_ACTION_STOP_EVENT_LISTEN, actionStopEventListen),

		mvuex.Action(VUEX_ACTION_REMOVE_SUCCEEDED_HID_JOBS, actionRemoveSucceededHidJobs),
		mvuex.Action(VUEX_ACTION_REMOVE_FAILED_HID_JOBS, actionRemoveFailedHidJobs),
		mvuex.Action(VUEX_ACTION_CANCEL_HID_JOB, actionCancelHidJob),


		mvuex.Getter("triggerActions", func(state *GlobalState) interface{} {
			return state.TriggerActionList.TriggerActions
		}),
		mvuex.Getter("hidjobs", func(state *GlobalState) interface{} {
			return state.HidJobList.Jobs
		}),

		mvuex.Getter("hidjobsRunning", func(state *GlobalState) interface{} {
			vJobs := state.HidJobList.Jobs                        //vue object, no real array --> values have to be extracted to filter
			jobs := js.Global.Get("Object").Call("values", vJobs) //converted to native JS array (has filter method available
			filtered := jobs.Call("filter", func(job *jsHidJobState) bool {
				return !(job.HasSucceeded || job.HasFailed)
			})
			return filtered
		}),
		mvuex.Getter("isConnected", func(state *GlobalState) interface{} {
			return state.EventListenerRunning
		}),

		mvuex.Getter("hidjobsFailed", func(state *GlobalState) interface{} {
			vJobs := state.HidJobList.Jobs                        //vue object, no real array --> values have to be extracted to filter
			jobs := js.Global.Get("Object").Call("values", vJobs) //converted to native JS array (has filter method available
			filtered := jobs.Call("filter", func(job *jsHidJobState) bool {
				return job.HasFailed
			})
			return filtered
		}),

		mvuex.Getter("hidjobsSucceeded", func(state *GlobalState) interface{} {
			println("Getter HID JOBS SUCCEEDED")
			vJobs := state.HidJobList.Jobs                        //vue object, no real array --> values have to be extracted to filter
			jobs := js.Global.Get("Object").Call("values", vJobs) //converted to native JS array (has filter method available
			filtered := jobs.Call("filter", func(job *jsHidJobState) bool {
				return job.HasSucceeded
			})
			return filtered
		}),


		mvuex.Getter("storedWifiSettingsSelect", func(state *GlobalState) interface{} {
			selectWS := js.Global.Get("Array").New()
			for _, curS := range state.StoredWifiSettingsList {
				option := struct {
					*js.Object
					Label string `js:"label"`
					Value string `js:"value"`
				}{Object: O()}
				option.Label = curS
				option.Value = curS
				selectWS.Call("push", option)
			}
			return selectWS
		}),
	)

	// fetch deployed gadget settings
	store.Dispatch(VUEX_ACTION_UPDATE_CURRENT_USB_SETTINGS)

	// Update already running HID jobs
	store.Dispatch(VUEX_ACTION_UPDATE_RUNNING_HID_JOBS)

	// Update WiFi state
	store.Dispatch(VUEX_ACTION_UPDATE_WIFI_STATE)

	// propagate Vuex store to global scope to allow injecting it to Vue by setting the "store" option
	js.Global.Set("store", store)


	return store
}

func InitGlobalState() *mvuex.Store {
	return initMVuex()
}
