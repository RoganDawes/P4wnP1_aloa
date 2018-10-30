// +build js

package main

import (
	//	"honnef.co/go/js/dom"
	"github.com/gopherjs/gopherjs/js"
	"github.com/mame82/hvue"
	"github.com/mame82/mvuex"
)

var (
	serverAddr = GetBaseURL()
	//RpcClient     = NewRpcClient(serverAddr + ":80")
	RpcClient     = NewRpcClient(serverAddr)

)

func GetBaseURL() string {
	document := js.Global.Get("window").Get("document")
	location := document.Get("location")
	port := location.Get("port").String()
	url := location.Get("protocol").String() + "//" + location.Get("hostname").String()
	if len(port) > 0 {
		url = url + ":" + port
	}
	return url
}

type appController struct {
	*js.Object
}

func Store(store *mvuex.Store) hvue.ComponentOption {
	return func(config *hvue.Config) {
		config.Set("store", store)
	}
}

func Router(router *js.Object) hvue.ComponentOption {
	return func(config *hvue.Config) {
		config.Set("router", router)
	}
}


func main() {
	println(GetBaseURL())

	store := InitGlobalState() //sets Vuex store in JS window.store

//	RpcClient.StartListening() //Start event listening after global state is initiated (contains the event handlers)

	// ToDo: delete because debug
//	RpcClient.GetAllDeployedEthernetInterfaceSettings(time.Second*10)

	router := NewVueRouter(
		"/usb",
		// route below could be used for an easter egg
		//VueRouterRoute("/","", "<usb-settings></usb-settings>"),
		VueRouterRoute("/usb","", "<usb-settings></usb-settings>"),
		VueRouterRoute("/hid","", "<hid-script></hid-script>"),
		VueRouterRoute("/hidjobs","", "<hid-job-event-overview></hid-job-event-overview>"),
		VueRouterRoute("/logger","", "<logger :max-entries='7'></logger>"),
		VueRouterRoute("/network","", "<network></network>"),
		VueRouterRoute("/wifi","", "<wifi></wifi>"),
		VueRouterRoute("/triggeractions","", "<triggeraction-manager></triggeraction-manager>"),
		VueRouterRoute("/bluetooth","", "<bluetooth></bluetooth>"),
	)


	InitComponentsDialog()
	InitCompHIDJobs()
	InitCompHIDEvents()
	InitCompModal()
	InitCompEthernetAddresses2()
	InitCompToggleSwitch()
	InitCompUSBSettings()
	InitComponentsHIDScript()
	InitCompLogger()
	InitCompState()
	InitComponentsNetwork()
	InitComponentsWiFi()
	InitComponentsTriggerActions()
	InitComponentsBluetooth()

	vm := hvue.NewVM(
		hvue.El("#app"),
		hvue.Template(templateMainApp),
		//add console to app as computed property, to allow debug output on vue events
		hvue.Computed("state", func(vm *hvue.VM) interface{} {
			return vm.Get("$store").Get("state") //works only with Vuex store option added
		}),
		hvue.BeforeMount(func(vm *hvue.VM) {
			vm.Get("$q").Get("addressbarColor").Call("set", "#027be3")
		}),
		Store(store), //include Vuex store in global scope, using own hvue fork, see here: https://github.com/HuckRidgeSW/hvue/pull/6
		Router(router),
	)
	// ToDo: remove next lines, debug code
	js.Global.Set("vm",vm)
}

const templateMainApp = `
    <q-layout view="lHh Lpr fFf">
        <q-layout-header :reveal="!$q.platform.is.desktop">
            <q-toolbar>
                <q-toolbar-title>
                    P4wnP1 web-frontend ({{ $store.getters.isConnected }})
					<span slot="subtitle" class="mobile-only">by MaMe82</span>
                </q-toolbar-title>
            </q-toolbar>
            <q-tabs>
                <q-route-tab default slot="title" to="usb" name="tab-usb" icon="usb" label="USB Settings"></q-route-tab>
                <q-route-tab slot="title" to="wifi" name="tab-wifi" icon="wifi" label="WiFi settings"></q-route-tab>
                <q-route-tab slot="title" to="bluetooth" name="tab-bluetooth" icon="bluetooth" label="Bluetooth"></q-route-tab>
                <q-route-tab slot="title" to="network" name="tab-network" icon="settings_ethernet" label="Network settings"></q-route-tab>
                <q-route-tab slot="title" to="triggeractions" name="tab-triggeraction" icon="whatshot" label="Trigger Actions"></q-route-tab>
                <q-route-tab slot="title" to="hid" name="tab-hid-script" icon="keyboard" label="HIDScript"></q-route-tab>
 <!--               <q-route-tab slot="title" to="hidjobs" name="tab-hid-jobs" icon="schedule" label="HID Events"></q-route-tab> -->
                <q-route-tab slot="title" to="logger" name="tab-logger" icon="message" label="Event Log"></q-route-tab>
            </q-tabs>
        </q-layout-header>


        <q-layout-footer class="desktop-only">
            <q-toolbar>
                <q-toolbar-title>
                    <div slot="subtitle">by MaMe82</div>
                </q-toolbar-title>
            </q-toolbar>
        </q-layout-footer>

        <q-page-container>
            <router-view></router-view>

			<disconnect-modal :value="!$store.getters.isConnected"></disconnect-modal>
        </q-page-container>


    </q-layout>
`