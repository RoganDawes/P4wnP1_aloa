// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/mame82/hvue"
	"github.com/mame82/mvuex"
	"time"
)

var globalState *GlobalState

const (
	maxLogEntries = 500

	VUEX_ACTION_UPDATE_RUNNING_HID_JOBS              = "updateRunningHidJobs"
	VUEX_ACTION_UPDATE_TRIGGER_ACTIONS              = "updateTriggerActions"
	VUEX_ACTION_DEPLOY_CURRENT_GADGET_SETTINGS       = "deployCurrentGadgetSettings"
	VUEX_ACTION_UPDATE_GADGET_SETTINGS_FROM_DEPLOYED = "updateCurrentGadgetSettingsFromDeployed"
	VUEX_ACTION_DEPLOY_ETHERNET_INTERFACE_SETTINGS   = "deployEthernetInterfaceSettings"
	VUEX_ACTION_UPDATE_WIFI_STATE                    = "updateCurrentWifiSettingsFromDeployed"
	VUEX_ACTION_DEPLOY_WIFI_SETTINGS                 = "deployWifiSettings"

	VUEX_ACTION_UPDATE_STORED_WIFI_SETTINGS_LIST = "updateStoredWifiSettingsList"
	VUEX_ACTION_STORE_WIFI_SETTINGS              = "storeWifiSettings"
	VUEX_ACTION_LOAD_WIFI_SETTINGS               = "storeWifiSettings"

	VUEX_MUTATION_SET_CURRENT_GADGET_SETTINGS_TO   = "setCurrentGadgetSettings"
	VUEX_MUTATION_SET_WIFI_STATE                   = "setCurrentWifiSettings"
	VUEX_MUTATION_SET_CURRENT_HID_SCRIPT_SOURCE_TO = "setCurrentHIDScriptSource"
	VUEX_MUTATION_SET_STORED_WIFI_SETTINGS_LIST    = "setStoredWifiSettingsList"

	initHIDScript = `layout('us');			// US keyboard layout
typingSpeed(100,150)	// Wait 100ms between key strokes + an additional random value between 0ms and 150ms (natural)

waitLEDRepeat(NUM);		// Wait till NUM LED of target changes frequently multiple times (doesn't work on OSX)
press("GUI r");
delay(500);
type("notepad\n")
delay(1000);
for (var i = 0; i < 3; i++) {
  type("Hello from P4wnP1 run " + i + " !\n");
  type("Moving mouse right ...");
  moveStepped(500,0);
  type("and left\n");
  moveStepped(-500,0);
}
type("Let's type fast !!!!!!!!!!!!!!!\n")
typingSpeed(0,0);
for (var i = 3; i < 10; i++) {
  type("Hello from P4wnP1 run " + i + " !\n");
  type("Moving mouse right ...");
  moveStepped(500,0);
  type("and left\n");
  moveStepped(-500,0);
}`
)

type GlobalState struct {
	*js.Object
	Title                            string                  `js:"title"`
	CurrentHIDScriptSource           string                  `js:"currentHIDScriptSource"`
	CurrentGadgetSettings            *jsGadgetSettings       `js:"currentGadgetSettings"`
	CurrentlyDeployingGadgetSettings bool                    `js:"deployingGadgetSettings"`
	CurrentlyDeployingWifiSettings   bool                    `js:"deployingWifiSettings"`
	EventReceiver                    *jsEventReceiver        `js:"eventReceiver"`
	HidJobList                       *jsHidJobStateList      `js:"hidJobList"`
	TriggerActionList *jsTriggerActionList `js:"triggerActionList"`
	IsModalEnabled                   bool                    `js:"isModalEnabled"`
	IsConnected                      bool                    `js:"isConnected"`
	FailedConnectionAttempts         int                     `js:"failedConnectionAttempts"`
	InterfaceSettings                *jsEthernetSettingsList `js:"InterfaceSettings"`
	//WiFiSettings                     *jsWiFiSettings         `js:"wifiSettings"`
	WiFiState *jsWiFiState `js:"wifiState"`

	StoredWifiSettingsList []string `js:"StoredWifiSettingsList"`
}

func createGlobalStateStruct() GlobalState {
	state := GlobalState{Object: O()}
	state.Title = "P4wnP1 by MaMe82"
	state.CurrentHIDScriptSource = initHIDScript
	state.CurrentGadgetSettings = NewUSBGadgetSettings()
	state.CurrentlyDeployingWifiSettings = false
	state.HidJobList = NewHIDJobStateList()
	state.TriggerActionList = NewTriggerActionList()
	state.EventReceiver = NewEventReceiver(maxLogEntries, state.HidJobList)
	state.IsConnected = false
	state.IsModalEnabled = false
	state.FailedConnectionAttempts = 0

	state.StoredWifiSettingsList = []string{}
	//Retrieve Interface settings
	// ToDo: Replace panics by default values
	ifSettings, err := RpcClient.GetAllDeployedEthernetInterfaceSettings(time.Second * 5)
	if err != nil {
		panic("Couldn't retrieve interface settings")
	}
	state.InterfaceSettings = ifSettings

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

func actionUpdateGadgetSettingsFromDeployed(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {
		//fetch deployed gadget settings
		dGS, err := RpcClient.RpcGetDeployedGadgetSettings(time.Second * 5)
		if err != nil {
			println("Couldn't retrieve deployed gadget settings")
			return
		}
		//convert to JS version
		jsGS := &jsGadgetSettings{Object: O()}
		jsGS.fromGS(dGS)

		//commit to current
		context.Commit("setCurrentGadgetSettings", jsGS)
	}()

	return
}

func actionUpdateWifiState(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {
		//fetch deployed gadget settings
		state, err := RpcClient.GetWifiState(time.Second * 5)
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
		wsList, err := RpcClient.GetStoredWifiSettingsList(time.Second * 10)
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
		err := RpcClient.StoreWifiSettings(time.Second*30, req.toGo())
		if err != nil {
			QuasarNotifyError("Error storing WiFi Settings", err.Error(), QUASAR_NOTIFICATION_POSITION_BOTTOM)
		}
		QuasarNotifySuccess("New WiFi settings stored", "", QUASAR_NOTIFICATION_POSITION_TOP)
	}()
}

func actionDeployWifiSettings(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, settings *jsWiFiSettings) {
	go func() {

		state.CurrentlyDeployingWifiSettings = true
		defer func() { state.CurrentlyDeployingWifiSettings = false }()

		println("Vuex dispatch deploy WiFi settings")
		// convert to Go type
		goSettings := settings.toGo()

		wstate, err := RpcClient.DeployWifiSettings(time.Second*30, goSettings)
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
		jobstates, err := RpcClient.RpcGetRunningHidJobStates(time.Second * 10)
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

func actionUpdateTriggerActions(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {
		// ToDo: replace by with RPC, dummy function

		//Test Actions
		// Generate an array of TriggerActions, which is provided from the Vuex store once the components ViewModel structures are finalized
		taList := []*jsVMTriggerAction{
			NewTriggerAction(),
			NewTriggerAction(),
			NewTriggerAction(),
			NewTriggerAction(),
			NewTriggerAction(),
			NewTriggerAction(),
			NewTriggerAction(),
			NewTriggerAction(),
			NewTriggerAction(),
			NewTriggerAction(),
		}
		taList[1].Immutable = true
		taList[2].IsActive = false
		for idx,ta := range taList {
			ta.Id = uint32(idx)
			state.TriggerActionList.UpdateEntry(ta)
		}
	}()

	return
}

func actionDeployCurrentGadgetSettings(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {

		state.CurrentlyDeployingGadgetSettings = true
		defer func() { state.CurrentlyDeployingGadgetSettings = false }()

		//get current GadgetSettings
		curGS := state.CurrentGadgetSettings.toGS()

		//try to set them via gRPC (the server holds an internal state, setting != deploying)
		err := RpcClient.RpcSetRemoteGadgetSettings(curGS, time.Second)
		if err != nil {
			QuasarNotifyError("Error in pre-check of new USB gadget settings", err.Error(), QUASAR_NOTIFICATION_POSITION_TOP)
			return
		}

		//try to deploy the, now set, remote GadgetSettings via gRPC
		_, err = RpcClient.RpcDeployRemoteGadgetSettings(time.Second * 10)
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

		err := RpcClient.DeployedEthernetInterfaceSettings(time.Second*3, goSettings)
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
		mvuex.Mutation(VUEX_MUTATION_SET_CURRENT_GADGET_SETTINGS_TO, func(store *mvuex.Store, state *GlobalState, settings *jsGadgetSettings) {
			state.CurrentGadgetSettings = settings
			return
		}),
		mvuex.Mutation(VUEX_MUTATION_SET_WIFI_STATE, func(store *mvuex.Store, state *GlobalState, wifiState *jsWiFiState) {
			state.WiFiState = wifiState
			return
		}),
		mvuex.Mutation(VUEX_MUTATION_SET_STORED_WIFI_SETTINGS_LIST, func(store *mvuex.Store, state *GlobalState, wsList []interface{}) {
			println("New ws list", wsList)
			hvue.Set(state, "StoredWifiSettingsList", wsList)
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
		mvuex.Action(VUEX_ACTION_UPDATE_GADGET_SETTINGS_FROM_DEPLOYED, actionUpdateGadgetSettingsFromDeployed),
		mvuex.Action(VUEX_ACTION_DEPLOY_CURRENT_GADGET_SETTINGS, actionDeployCurrentGadgetSettings),
		mvuex.Action(VUEX_ACTION_UPDATE_RUNNING_HID_JOBS, actionUpdateRunningHidJobs),
		mvuex.Action(VUEX_ACTION_DEPLOY_ETHERNET_INTERFACE_SETTINGS, actionDeployEthernetInterfaceSettings),
		mvuex.Action(VUEX_ACTION_UPDATE_WIFI_STATE, actionUpdateWifiState),
		mvuex.Action(VUEX_ACTION_DEPLOY_WIFI_SETTINGS, actionDeployWifiSettings),
		mvuex.Action(VUEX_ACTION_UPDATE_STORED_WIFI_SETTINGS_LIST, actionUpdateStoredWifiSettingsList),
		mvuex.Action(VUEX_ACTION_STORE_WIFI_SETTINGS, actionStoreWifiSettings),
		mvuex.Action(VUEX_ACTION_UPDATE_TRIGGER_ACTIONS, actionUpdateTriggerActions),

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
			for _,curS := range state.StoredWifiSettingsList  {
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
	store.Dispatch(VUEX_ACTION_UPDATE_GADGET_SETTINGS_FROM_DEPLOYED)

	// Update already running HID jobs
	store.Dispatch(VUEX_ACTION_UPDATE_RUNNING_HID_JOBS)

	// Update WiFi state
	store.Dispatch(VUEX_ACTION_UPDATE_WIFI_STATE)

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
