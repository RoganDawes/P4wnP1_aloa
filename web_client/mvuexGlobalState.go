// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	pb "github.com/mame82/P4wnP1_go/proto/gopherjs"
	"github.com/mame82/hvue"
	"github.com/mame82/mvuex"
	"github.com/pkg/errors"
	"path/filepath"
	"strings"
	"time"
)

var globalState *GlobalState

const (
	maxLogEntries = 500

	//Bluetooth
	VUEX_ACTION_UPDATE_CURRENT_BLUETOOTH_CONTROLLER_INFORMATION = "updateCurrentBluetoothControllerInformation"
	VUEX_ACTION_DEPLOY_CURRENT_BLUETOOTH_CONTROLLER_INFORMATION = "deployCurrentBluetoothControllerInformation"

	VUEX_MUTATION_SET_CURRENT_BLUETOOTH_CONTROLLER_INFORMATION = "setCurrentBluetoothControllerInformation"

	//HIDScripts and jobs
	VUEX_ACTION_UPDATE_RUNNING_HID_JOBS                           = "updateRunningHidJobs"
	VUEX_ACTION_UPDATE_STORED_HID_SCRIPTS_LIST                    = "updateStoredHIDScriptsList"
	VUEX_ACTION_UPDATE_CURRENT_HID_SCRIPT_SOURCE_FROM_REMOTE_FILE = "updateCurrentHidScriptSourceFromRemoteFile"
	VUEX_ACTION_STORE_CURRENT_HID_SCRIPT_SOURCE_TO_REMOTE_FILE    = "storeCurrentHidScriptSourceToRemoteFile"

	VUEX_MUTATION_SET_CURRENT_HID_SCRIPT_SOURCE_TO = "setCurrentHIDScriptSource"
	VUEX_MUTATION_SET_STORED_HID_SCRIPTS_LIST      = "setStoredHIDScriptsList"

	//USBGadget
	VUEX_ACTION_DEPLOY_CURRENT_USB_SETTINGS     = "deployCurrentUSBSettings"
	VUEX_ACTION_UPDATE_CURRENT_USB_SETTINGS     = "updateCurrentUSBSettings"
	VUEX_ACTION_STORE_USB_SETTINGS              = "storeUSBSettings"
	VUEX_ACTION_LOAD_USB_SETTINGS               = "loadUSBSettings"
	VUEX_ACTION_DEPLOY_STORED_USB_SETTINGS      = "deployStoredUSBSettings"
	VUEX_ACTION_UPDATE_STORED_USB_SETTINGS_LIST = "updateStoredUSBSettingsList"

	VUEX_MUTATION_SET_CURRENT_USB_SETTINGS     = "setCurrentUSBSettings"
	VUEX_MUTATION_SET_STORED_USB_SETTINGS_LIST = "setStoredUSBSettingsList"

	// Ethernet
	VUEX_ACTION_UPDATE_ALL_ETHERNET_INTERFACE_SETTINGS         = "updateAllEthernetInterfaceSettings"
	VUEX_ACTION_DEPLOY_ETHERNET_INTERFACE_SETTINGS             = "deployEthernetInterfaceSettings"
	VUEX_ACTION_STORE_ETHERNET_INTERFACE_SETTINGS              = "storeEthernetInterfaceSettings"
	VUEX_ACTION_LOAD_ETHERNET_INTERFACE_SETTINGS               = "loadEthernetInterfaceSettings"
	VUEX_ACTION_DEPLOY_STORED_ETHERNET_INTERFACE_SETTINGS      = "deployStoredEthernetInterfaceSettings"
	VUEX_ACTION_UPDATE_STORED_ETHERNET_INTERFACE_SETTINGS_LIST = "updateStoredEthernetInterfaceSettingsList"

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
	Title                                 string                            `js:"title"`
	CurrentHIDScriptSource                string                            `js:"currentHIDScriptSource"`
	CurrentGadgetSettings                 *jsGadgetSettings                 `js:"currentGadgetSettings"`
	CurrentlyDeployingGadgetSettings      bool                              `js:"deployingGadgetSettings"`
	CurrentlyDeployingWifiSettings        bool                              `js:"deployingWifiSettings"`
	EventReceiver                         *jsEventReceiver                  `js:"eventReceiver"`
	HidJobList                            *jsHidJobStateList                `js:"hidJobList"`
	TriggerActionList                     *jsTriggerActionSet               `js:"triggerActionList"`
	IsModalEnabled                        bool                              `js:"isModalEnabled"`
	IsConnected                           bool                              `js:"isConnected"`
	FailedConnectionAttempts              int                               `js:"failedConnectionAttempts"`
	InterfaceSettings                     *jsEthernetSettingsList           `js:"InterfaceSettings"`
	WiFiState                             *jsWiFiState                      `js:"wifiState"`
	CurrentBluetoothControllerInformation *jsBluetoothControllerInformation `js:"CurrentBluetoothControllerInformation"`

	StoredWifiSettingsList                []string                          `js:"StoredWifiSettingsList"`
	StoredEthernetInterfaceSettingsList   []string                          `js:"StoredEthernetInterfaceSettingsList"`
	StoredTriggerActionSetsList           []string                          `js:"StoredTriggerActionSetsList"`
	StoredBashScriptsList                 []string                          `js:"StoredBashScriptsList"`
	StoredHIDScriptsList                  []string                          `js:"StoredHIDScriptsList"`
	StoredUSBSettingsList                 []string                          `js:"StoredUSBSettingsList"`
}

func createGlobalStateStruct() GlobalState {
	state := GlobalState{Object: O()}
	state.Title = "P4wnP1 by MaMe82"
	state.CurrentHIDScriptSource = initHIDScript
	state.CurrentGadgetSettings = NewUSBGadgetSettings()
	state.CurrentlyDeployingWifiSettings = false
	state.HidJobList = NewHIDJobStateList()
	state.TriggerActionList = NewTriggerActionSet()
	state.EventReceiver = NewEventReceiver(maxLogEntries, state.HidJobList)
	state.IsConnected = false
	state.IsModalEnabled = false
	state.FailedConnectionAttempts = 0

	state.StoredWifiSettingsList = []string{}
	state.StoredEthernetInterfaceSettingsList = []string{}
	state.StoredTriggerActionSetsList = []string{}
	state.StoredBashScriptsList = []string{}
	state.StoredHIDScriptsList = []string{}
	state.StoredUSBSettingsList = []string{}
	//Retrieve Interface settings
	state.InterfaceSettings = NewEthernetSettingsList()
	state.CurrentBluetoothControllerInformation = NewBluetoothControllerInformation()

	/*
	wifiSettings, err := RpcClient.GetWifiState(time.Second * 5)
	if err != nil {
		panic("Couldn't retrieve WiFi settings")
	}
	*/
	//state.WiFiSettings = NewWifiSettings()
	state.WiFiState = NewWiFiState()
	return state
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
		println("Trying to deploy bluetooth controller information: ", state.CurrentBluetoothControllerInformation.CurrentSettings.Powered)
		//fetch deployed gadget settings
		res, err := RpcClient.DeployBluetoothControllerInformation(defaultTimeout, state.CurrentBluetoothControllerInformation)
		if err != nil {
			println("Couldn't deploy BluetoothControllerInformation", err)
			return
		}

		println("Bluetooth Controller Info after deploy: ", res.CurrentSettings.Powered)
		context.Commit(VUEX_MUTATION_SET_CURRENT_BLUETOOTH_CONTROLLER_INFORMATION, res)
	}()

	return
}

func actionUpdateStoredUSBSettingsList(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {
		println("Trying to fetch wsList")
		//fetch deployed gadget settings
		wsList, err := RpcClient.GetStoredUSBSettingsList(defaultTimeout)
		if err != nil {
			println("Couldn't retrieve WifiSettingsList")
			return
		}

		//commit to current
		println(wsList)
		context.Commit(VUEX_MUTATION_SET_STORED_USB_SETTINGS_LIST, wsList)
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
		println("Vuex dispatch load ethernet interface settings: ", settingsName.String())
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
		wsList, err := RpcClient.GetStoredBashScriptsList(defaultTimeout)
		if err != nil {
			println("Couldn't retrieve stored BashScripts list")
			return
		}

		//commit to current
		println(wsList)
		context.Commit(VUEX_MUTATION_SET_STORED_BASH_SCRIPTS_LIST, wsList)
	}()

	return
}

func actionUpdateStoredHIDScriptsList(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {
		println("Trying to fetch stored HIDScripts list")
		//fetch deployed gadget settings
		wsList, err := RpcClient.GetStoredHIDScriptsList(defaultTimeout)
		if err != nil {
			println("Couldn't retrieve  stored HIDScripts list")
			return
		}

		//commit to current
		println(wsList)
		context.Commit(VUEX_MUTATION_SET_STORED_HID_SCRIPTS_LIST, wsList)
	}()

	return
}

func actionUpdateGadgetSettingsFromDeployed(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {
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
		//fetch deployed gadget settings
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
		println("Trying to fetch wsList")
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
		println("Vuex dispatch load WiFi settings: ", settingsName.String())
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
		//fetch deployed gadget settings
		jobstates, err := RpcClient.GetRunningHidJobStates(defaultTimeout)
		if err != nil {
			println("Couldn't retrieve stateof running HID jobs", err)
			return
		}

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
			println("Couldn't retrieve WifiSettingsList")
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
	globalState = &state //make accessible through global var
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
		mvuex.Mutation(VUEX_MUTATION_SET_ALL_ETHERNET_INTERFACE_SETTINGS, func(store *mvuex.Store, state *GlobalState, eifSettings *jsEthernetSettingsList) {
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
		/*
		mvuex.Mutation("startLogListening", func (store *mvuex.Store, state *GlobalState) {
			state.EventReceiver.StartListening()
			return
		}),
		mvuex.Mutation("stopLogListening", func (store *mvuex.Store, state *GlobalState) {
			state.EventReceiver.StopListening()
			return
		}),
		*/
		mvuex.Action(VUEX_ACTION_UPDATE_CURRENT_BLUETOOTH_CONTROLLER_INFORMATION, actionUpdateCurrentBluetoothControllerInformation),
		mvuex.Action(VUEX_ACTION_DEPLOY_CURRENT_BLUETOOTH_CONTROLLER_INFORMATION, actionDeployCurrentBluetoothControllerInformation),

		mvuex.Action(VUEX_ACTION_UPDATE_CURRENT_USB_SETTINGS, actionUpdateGadgetSettingsFromDeployed),
		mvuex.Action(VUEX_ACTION_DEPLOY_CURRENT_USB_SETTINGS, actionDeployCurrentGadgetSettings),
		mvuex.Action(VUEX_ACTION_UPDATE_RUNNING_HID_JOBS, actionUpdateRunningHidJobs),
		mvuex.Action(VUEX_ACTION_STORE_USB_SETTINGS, actionStoreUSBSettings),
		mvuex.Action(VUEX_ACTION_LOAD_USB_SETTINGS, actionLoadUSBSettings),
		mvuex.Action(VUEX_ACTION_DEPLOY_STORED_USB_SETTINGS, actionDeployStoredUSBSettings),
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

		mvuex.Action(VUEX_ACTION_STORE_ETHERNET_INTERFACE_SETTINGS, actionStoreEthernetInterfaceSettings),
		mvuex.Action(VUEX_ACTION_LOAD_ETHERNET_INTERFACE_SETTINGS, actionLoadEthernetInterfaceSettings),
		mvuex.Action(VUEX_ACTION_DEPLOY_STORED_ETHERNET_INTERFACE_SETTINGS, actionDeployStoredEthernetInterfaceSettings),


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

		mvuex.Getter("hidjobsFailed", func(state *GlobalState) interface{} {
			vJobs := state.HidJobList.Jobs                        //vue object, no real array --> values have to be extracted to filter
			jobs := js.Global.Get("Object").Call("values", vJobs) //converted to native JS array (has filter method available
			filtered := jobs.Call("filter", func(job *jsHidJobState) bool {
				return job.HasFailed
			})
			return filtered
		}),

		mvuex.Getter("hidjobsSucceeded", func(state *GlobalState) interface{} {
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

	/*
	// Update ethernet interface state (done by component)
	store.Dispatch(VUEX_ACTION_UPDATE_ALL_ETHERNET_INTERFACE_SETTINGS)
	*/

	// propagate Vuex store to global scope to allow injecting it to Vue by setting the "store" option
	js.Global.Set("store", store)

	/*
	// Start Event Listening
	state.EventReceiver.StartListening()
	*/

	return store
}

func InitGlobalState() *mvuex.Store {
	return initMVuex()
}
