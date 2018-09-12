// +build js

package main

import (
//	"honnef.co/go/js/dom"
	"github.com/gopherjs/gopherjs/js"
	"github.com/HuckRidgeSW/hvue"
	"time"
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
	RpcClient.StartListening() //Start event listening after global state is initiated (contains the event handlers)

	// ToDo: delete because debug
	RpcClient.GetAllDeployedEthernetInterfaceSettings(time.Second*10)

	router := NewVueRouter(
		VueRouterRoute("/usb","", "<usb-settings></usb-settings>"),
		VueRouterRoute("/","", "<usb-settings></usb-settings>"),
		VueRouterRoute("/hid","", "<hid-script></hid-script>"),
		VueRouterRoute("/hidjobs","", "<hid-job-event-overview></hid-job-event-overview>"),
		VueRouterRoute("/logger","", "<logger :max-entries='7'></logger>"),
		VueRouterRoute("/network","", "<network></network>"),
		VueRouterRoute("/wifi","", "<wifi></wifi>"),
	)


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
	vm := hvue.NewVM(
		hvue.El("#app"),
		//add "testString" to data
		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := struct{
				*js.Object
				TestString string `js:"testString"`
				SelectedTab string `js:"selectedTab"`
			}{Object: O()}
			data.SelectedTab = "USB"
			data.TestString = "type('hello');"
			return &data
		}),
		//add console to app as computed property, to allow debug output on vue events
		hvue.Computed(
			"console",
			func(vm *hvue.VM) interface{} {
			return js.Global.Get("console")
		}),
		hvue.Computed("state", func(vm *hvue.VM) interface{} {
			return vm.Get("$store").Get("state") //works only with Vuex store option added
		}),
		Store(store), //include Vuex store in global scope, using own hvue fork, see here: https://github.com/HuckRidgeSW/hvue/pull/6
		Router(router),
	)
	// ToDo: remove next line, debug code
	js.Global.Set("vm",vm)

}
