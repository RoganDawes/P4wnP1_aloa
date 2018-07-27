package main

import (
	"./mvuex"
	"github.com/gopherjs/gopherjs/js"
	"time"
	"context"
	pb "../proto/gopherjs"
)

const (
	maxLogEntries = 500
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

	Counter int `js:"count"`
	Text string `js:"text"`
}

func UpdateGadgetSettingsFromDeployed(jsGS *jsGadgetSettings) {
	//gs := vue.GetVM(c).Get("gadgetSettings")
	println("UpdateGadgetSettingsFromDeployed called")

	ctx,cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()


	deployedGs, err := Client.Client.GetDeployedGadgetSetting(ctx, &pb.Empty{})
	if err != nil { println(err); return } // ToDo: change to alert with parsed status

	jsGS.fromGS(deployedGs)
	return
}

func createGlobalStateStruct() GlobalState {
	state := GlobalState{Object:O()}
	state.Title = "P4wnP1 by MaMe82"
	state.CurrentHIDScriptSource = initHIDScript
	state.CurrentGadgetSettings = NewUSBGadgetSettings()
	UpdateGadgetSettingsFromDeployed(state.CurrentGadgetSettings)
	state.EventLog = NewLogger(maxLogEntries)

	state.Counter = 1337
	state.Text = "Hi there says MaMe82"
	return state
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
		mvuex.Mutation("setCurrentHIDScriptSource", func (store *mvuex.Store, state *GlobalState, newText string) {
			state.CurrentHIDScriptSource = newText
			return
		}),
		mvuex.Mutation("setCurrentGadgetSettings", func (store *mvuex.Store, state *GlobalState, newSettings *jsGadgetSettings) {
			state.CurrentGadgetSettings = newSettings
			return
		}),
		mvuex.Mutation("setCurrentGadgetSettingsFromDeployed", func (store *mvuex.Store, state *GlobalState) {
			//ToDo: check if this is valid for synchronous run, has to be dispatched to action otherwise
			println("Store: commit setCurrentGadgetSettingsFromDeployed")
			go UpdateGadgetSettingsFromDeployed(state.CurrentGadgetSettings)
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
	)


	// propagate Vuex store to global scope to allow injecting it to Vue by setting the "store" option
	js.Global.Set("store", store)

	// Start Event Listening
	state.EventLog.StartListening()

}

func InitGlobalState() {
	initMVuex()

}