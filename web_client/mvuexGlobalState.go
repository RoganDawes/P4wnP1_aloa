// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"time"
	"github.com/mame82/mvuex"
)

var globalState *GlobalState

const (
	maxLogEntries = 500

	VUEX_ACTION_UPDATE_RUNNING_HID_JOBS              = "updateRunningHidJobs"
	VUEX_ACTION_DEPLOY_CURRENT_GADGET_SETTINGS       = "deployCurrentGadgetSettings"
	VUEX_ACTION_UPDATE_GADGET_SETTINGS_FROM_DEPLOYED = "updateCurrentGadgetSettingsFromDeployed"
	VUEX_ACTION_DEPLOY_ETHERNET_INTERFACE_SETTINGS   = "deployEthernetInterfaceSettings"
	VUEX_ACTION_UPDATE_WIFI_SETTINGS_FROM_DEPLOYED   = "updateCurrentWifiSettingsFromDeployed"
	VUEX_ACTION_DEPLOY_WIFI_SETTINGS                 = "deployWifiSettings"

	VUEX_MUTATION_SET_CURRENT_GADGET_SETTINGS_TO   = "setCurrentGadgetSettings"
	VUEX_MUTATION_SET_CURRENT_WIFI_SETTINGS        = "setCurrentWifiSettings"
	VUEX_MUTATION_SET_CURRENT_HID_SCRIPT_SOURCE_TO = "setCurrentHIDScriptSource"

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
	EventReceiver                    *jsEventReceiver        `js:"eventReceiver"`
	HidJobList                       *jsHidJobStateList      `js:"hidJobList"`
	IsModalEnabled                   bool                    `js:"isModalEnabled"`
	IsConnected                      bool                    `js:"isConnected"`
	FailedConnectionAttempts         int                     `js:"failedConnectionAttempts"`
	InterfaceSettings                *jsEthernetSettingsList `js:"InterfaceSettings"`
	WiFiSettings                     *jsWiFiSettings         `js:"wifiSettings"`
	WiFiState                        *jsWiFiConnectionState  `js:"wifiConnectionState"`
}

func createGlobalStateStruct() GlobalState {
	state := GlobalState{Object: O()}
	state.Title = "P4wnP1 by MaMe82"
	state.CurrentHIDScriptSource = initHIDScript
	state.CurrentGadgetSettings = NewUSBGadgetSettings()
	state.CurrentlyDeployingGadgetSettings = false
	//UpdateGadgetSettingsFromDeployed(state.CurrentGadgetSettings)
	state.HidJobList = NewHIDJobStateList()
	state.EventReceiver = NewEventReceiver(maxLogEntries, state.HidJobList)
	state.IsConnected = false
	state.IsModalEnabled = false
	state.FailedConnectionAttempts = 0
	//Retrieve Interface settings
	// ToDo: Replace panics by default values
	ifSettings, err := RpcClient.GetAllDeployedEthernetInterfaceSettings(time.Second * 5)
	if err != nil {
		panic("Couldn't retrieve interface settings")
	}
	state.InterfaceSettings = ifSettings
	wifiSettings, err := RpcClient.GetDeployedWiFiSettings(time.Second * 5)
	if err != nil {
		panic("Couldn't retrieve WiFi settings")
	}
	state.WiFiSettings = wifiSettings
	state.WiFiState = NewWiFiConnectionState()
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

func actionUpdateWifiSettingsFromDeployed(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {
		//fetch deployed gadget settings
		dWS, err := RpcClient.GetDeployedWiFiSettings(time.Second * 5)
		if err != nil {
			println("Couldn't retrieve deployed WiFi settings")
			return
		}

		//commit to current
		context.Commit(VUEX_MUTATION_SET_CURRENT_WIFI_SETTINGS, dWS)
	}()

	return
}

/*
func actionDeployWifiSettings(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, settings *jsWiFiSettings) {
	go func() {
		println("Vuex dispatch deploy WiFi settings")
		// convert to Go type
		goSettings := settings.toGo()

		wstate, err := RpcClient.DeployWifiSettings(time.Second*20, goSettings)
		if err != nil {
			QuasarNotifyError("Error deploying WiFi Settings", err.Error(), QUASAR_NOTIFICATION_POSITION_BOTTOM)
		}
		state.WiFiState.fromGo(wstate)
	}()
}
*/
func actionDeployWifiSettings(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState, settings *jsWiFiSettings) {
	go func() {
		println("Vuex dispatch deploy WiFi settings")
		// convert to Go type
		goSettings := settings.toGo2()

		wstate, err := RpcClient.DeployWifiSettings2(time.Second*20, goSettings)
		if err != nil {
			QuasarNotifyError("Error deploying WiFi Settings", err.Error(), QUASAR_NOTIFICATION_POSITION_BOTTOM)
		}
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

func actionDeployCurrentGadgetSettings(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {

		// ToDo: Indicate deployment process via global state
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
		mvuex.Mutation(VUEX_MUTATION_SET_CURRENT_WIFI_SETTINGS, func(store *mvuex.Store, state *GlobalState, settings *jsWiFiSettings) {
			state.WiFiSettings = settings
			return
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
		mvuex.Action(VUEX_ACTION_UPDATE_WIFI_SETTINGS_FROM_DEPLOYED, actionUpdateWifiSettingsFromDeployed),
		mvuex.Action(VUEX_ACTION_DEPLOY_WIFI_SETTINGS, actionDeployWifiSettings),

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


	)

	// fetch deployed gadget settings
	store.Dispatch(VUEX_ACTION_UPDATE_GADGET_SETTINGS_FROM_DEPLOYED)

	// Update already running HID jobs
	store.Dispatch(VUEX_ACTION_UPDATE_RUNNING_HID_JOBS)

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
