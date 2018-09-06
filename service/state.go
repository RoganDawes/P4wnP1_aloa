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
	Wifi *WifiState
}

func InitGlobalServiceState() (err error) {
	state := &GlobalServiceState{}
	ServiceState = state // store state in global variable

	state.StoredNetworkSetting = make(map[string]*pb.EthernetInterfaceSettings)

	/*
	state.StoredNetworkSetting[USB_ETHERNET_BRIDGE_NAME] = GetDefaultNetworkSettingsUSB()
	state.StoredNetworkSetting["wlan0"] = GetDefaultNetworkSettingsWiFi()
	*/
	//pre initialize Default settings for "wlan0" and USB_ETHERNET_BRIDGE_NAME ("usbeth")
	state.StoredNetworkSetting[USB_ETHERNET_BRIDGE_NAME] = &pb.EthernetInterfaceSettings{
		Name: USB_ETHERNET_BRIDGE_NAME,
		Enabled: false,
		Mode: pb.EthernetInterfaceSettings_MANUAL,
		IpAddress4:         "172.16.0.1",
		Netmask4:           "255.255.255.252",
	}
	state.StoredNetworkSetting["wlan0"] = &pb.EthernetInterfaceSettings{
		Name: "wlan0",
		Enabled: false,
		Mode: pb.EthernetInterfaceSettings_MANUAL,
		IpAddress4:         "172.24.0.1",
		Netmask4:           "255.255.255.0",
	}
	state.Wifi = NewWifiState(GetDefaultWiFiSettings(), wifi_if_name)

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