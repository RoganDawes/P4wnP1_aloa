package service

import (
	pb "github.com/mame82/P4wnP1_go/proto"
)

var ServiceState *GlobalServiceState

type GlobalServiceState struct {
	EvMgr    *EventManager
	UsbGM 	*UsbGadgetManager
	Led 	*LedState
	HidDevPath map[string]string //stores device path for HID devices
	StoredNetworkSetting map[string]*pb.EthernetInterfaceSettings
}

func InitGlobalServiceState() (err error) {
	state := &GlobalServiceState{}
	ServiceState = state // store state in global variable

	state.StoredNetworkSetting = make(map[string]*pb.EthernetInterfaceSettings)
	//preinitialize Default settings for "wlan0" and USB_ETHERNET_BRIDGE_NAME ("usbeth")
	state.StoredNetworkSetting[USB_ETHERNET_BRIDGE_NAME] = GetDefaultNetworkSettingsUSB()
	state.StoredNetworkSetting["wlan0"] = GetDefaultNetworkSettingsWiFi()


	state.HidDevPath  = make(map[string]string) //should be initialized BEFORE UsbGadgetManager uses it
	state.EvMgr = NewEventManager(20)
	state.UsbGM,err = NewUSBGadgetManager()
	if err != nil { return }
	ledState, err := NewLed(false)
	if err != nil { return }
	state.Led = ledState

	return nil
}


func (state *GlobalServiceState) StartService() {
	state.EvMgr.Start()
}

func (state *GlobalServiceState) StopService() {
	state.EvMgr.Stop()
}