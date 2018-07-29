package main

import (
	"github.com/gopherjs/gopherjs/js"
	"time"
	"github.com/mame82/mvuex"
)

const (
	maxLogEntries = 500

	VUEX_ACTION_DEPLOY_CURRENT_GADGET_SETTINGS       = "deployCurrentGadgetSettings"
	VUEX_ACTION_UPDATE_GADGET_SETTINGS_FROM_DEPLOYED = "updateCurrentGadgetSettingsFromDeployed"
	VUEX_MUTATION_SET_CURRENT_GADGET_SETTINGS_TO     = "setCurrentGadgetSettings"
	VUEX_MUTATION_SET_CURRENT_HID_SCRIPT_SOURCE_TO   = "setCurrentHIDScriptSource"

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
	Title string `js:"title"`
	CurrentHIDScriptSource string `js:"currentHIDScriptSource"`
	CurrentGadgetSettings *jsGadgetSettings `js:"currentGadgetSettings"`
	EventLog *jsLoggerData `js:"eventLog"`
	IsModalEnabled bool `js:"isModalEnabled"`

	Counter int `js:"count"`
	Text string `js:"text"`
}


func createGlobalStateStruct() GlobalState {
	state := GlobalState{Object:O()}
	state.Title = "P4wnP1 by MaMe82"
	state.CurrentHIDScriptSource = initHIDScript
	state.CurrentGadgetSettings = NewUSBGadgetSettings()
	//UpdateGadgetSettingsFromDeployed(state.CurrentGadgetSettings)
	state.EventLog = NewLogger(maxLogEntries)
	state.IsModalEnabled = true

	state.Counter = 1337
	state.Text = "Hi there says MaMe82"
	return state
}

func actionUpdateGadgetSettingsFromDeployed(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {
		//fetch deployed gadget settings
		dGS,err := RpcGetDeployedGadgetSettings(time.Second * 3)
		if err != nil {
			println("Couldn't retrieve deployed gadget settings")
			return
		}
		//convert to JS version
		jsGS := &jsGadgetSettings{Object:O()}
		jsGS.fromGS(dGS)

		//commit to current
		context.Commit("setCurrentGadgetSettings", jsGS)
	}()

	return
}

func actionDeployCurrentGadgetSettings(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
	go func() {
		// ToDo: Indicate deployment process via global state

		//get current GadgetSettings
		curGS := state.CurrentGadgetSettings.toGS()

		//try to set them via gRPC (the server holds an internal state, setting != deploying)
		err := RpcSetRemoteGadgetSettings(curGS, time.Second)
		if err != nil {
			//ToDo: use global store to return something, or allow actions to return promises (latter is too much JavaScript)
			Alert(err.Error())
			return
		}

		//try to deploy the, now set, remote GadgetSettings via gRPC
		_,err = RpcDeployRemoteGadgetSettings(time.Second*10)
		if err != nil {
			//ToDo: use global store to return something, or allow actions to return promises (latter is too much JavaScript)
			Alert(err.Error())
			return
		}



		//ToDo: If we're here, we succeeded and should indicate this via global state
		Alert("GadgetSettings deployed successfully")

	}()

	return
}

func initMVuex() {
	state := createGlobalStateStruct()
	store := mvuex.NewStore(
		mvuex.State(state),
		mvuex.Action("actiontest", func(store *mvuex.Store, context *mvuex.ActionContext, state *GlobalState) {
			go func() {
				for i:=0; i<10; i++ {
					println(state.Counter)
					time.Sleep(1*time.Second)
					context.Commit("increment",5)
				}

			}()

		}),
		mvuex.Mutation("setModalEnabled", func (store *mvuex.Store, state *GlobalState, enabled bool) {
			state.IsModalEnabled = enabled
			return
		}),
		mvuex.Mutation("increment", func (store *mvuex.Store, state *GlobalState, add int) {
			state.Counter += add
			return
		}),
		mvuex.Mutation("decrement", func (store *mvuex.Store, state *GlobalState) {
			state.Counter--
			return
		}),
		mvuex.Mutation("setText", func (store *mvuex.Store, state *GlobalState, newText string) {
			state.Text = newText
			return
		}),
		mvuex.Mutation(VUEX_MUTATION_SET_CURRENT_HID_SCRIPT_SOURCE_TO, func (store *mvuex.Store, state *GlobalState, newText string) {
			state.CurrentHIDScriptSource = newText
			return
		}),
		mvuex.Mutation(VUEX_MUTATION_SET_CURRENT_GADGET_SETTINGS_TO, func (store *mvuex.Store, state *GlobalState, settings *jsGadgetSettings) {
			state.CurrentGadgetSettings = settings
			return
		}),
		mvuex.Mutation("startLogListening", func (store *mvuex.Store, state *GlobalState) {
			state.EventLog.StartListening()
			return
		}),
		mvuex.Mutation("stopLogListening", func (store *mvuex.Store, state *GlobalState) {
			state.EventLog.StopListening()
			return
		}),
		mvuex.Action(VUEX_ACTION_UPDATE_GADGET_SETTINGS_FROM_DEPLOYED, actionUpdateGadgetSettingsFromDeployed),
		mvuex.Action(VUEX_ACTION_DEPLOY_CURRENT_GADGET_SETTINGS, actionDeployCurrentGadgetSettings),
	)

	// fetch deployed gadget settings
	store.Dispatch("updateCurrentGadgetSettingsFromDeployed")

	// propagate Vuex store to global scope to allow injecting it to Vue by setting the "store" option
	js.Global.Set("store", store)

	// Start Event Listening
	state.EventLog.StartListening()

}

func InitGlobalState() {
	initMVuex()

}