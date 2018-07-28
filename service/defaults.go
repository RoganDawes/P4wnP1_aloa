package service

import (
	pb "../proto"
)

const (
	DEFAULT_CDC_ECM_HOST_ADDR = "42:63:66:12:34:56"
	DEFAULT_CDC_ECM_DEV_ADDR  = "42:63:66:56:34:12"
	DEFAULT_RNDIS_HOST_ADDR   = "42:63:65:12:34:56"
	DEFAULT_RNDIS_DEV_ADDR    = "42:63:65:56:34:12"
	USB_ETHERNET_BRIDGE_MAC   = "24:22:26:12:14:16"
	USB_ETHERNET_BRIDGE_NAME  = "usbeth"
)

func GetDefaultNetworkSettingsUSB() (*pb.EthernetInterfaceSettings) {
	//configure 172.24.0.1/255.255.255.252 for usbeth
	ifSettings := &pb.EthernetInterfaceSettings {
		Enabled:            true,
		Name:               USB_ETHERNET_BRIDGE_NAME,
		IpAddress4:         "172.16.0.1",
		Netmask4:           "255.255.255.252",
		Mode:               pb.EthernetInterfaceSettings_DHCP_SERVER,
		DhcpServerSettings: GetDefaultDHCPConfigUSB(),
	}
	return ifSettings
}

func GetDefaultNetworkSettingsWiFi() (*pb.EthernetInterfaceSettings) {
	ifSettings := &pb.EthernetInterfaceSettings {
		Enabled:            true,
		Name:               "wlan0",
		Mode:               pb.EthernetInterfaceSettings_DHCP_SERVER,
		IpAddress4:         "172.24.0.1",
		Netmask4:           "255.255.255.0",
		DhcpServerSettings: GetDefaultDHCPConfigWiFi(),
	}
	return ifSettings
}

func GetDefaultDHCPConfigUSB() (settings *pb.DHCPServerSettings) {
	settings = &pb.DHCPServerSettings{
		//CallbackScript:     "/bin/evilscript",
		DoNotBindInterface: false, //only bind to given interface
		ListenInterface:    USB_ETHERNET_BRIDGE_NAME,
		LeaseFile:          "/tmp/dnsmasq_" + USB_ETHERNET_BRIDGE_NAME + ".leases",
		ListenPort:         0,     //No DNS, DHCP only
		NotAuthoritative:   false, //be authoritative
		Ranges: []*pb.DHCPServerRange{
			&pb.DHCPServerRange{RangeLower: "172.16.0.2", RangeUpper: "172.16.0.2", LeaseTime: "5m"},
			//&pb.DHCPServerRange{RangeLower: "172.16.0.5", RangeUpper: "172.16.0.6", LeaseTime: "2m"},
		},
		Options: map[uint32]string{
			//Note: Options 1 (Netmask), 12 (Hostname) and 28 (Broadcast Address) are still enabled
			3:   "", //Disable option: Router
			6:   "", //Disable option: DNS
			//252: "http://172.16.0.1/wpad.dat",
		},
	}
	return
}

func GetDefaultDHCPConfigWiFi() (settings *pb.DHCPServerSettings) {
	settings = &pb.DHCPServerSettings{
		//CallbackScript:     "/bin/evilscript",
		DoNotBindInterface: false, //only bind to given interface
		ListenInterface:    "wlan0",
		LeaseFile:          "/tmp/dnsmasq_wlan0.leases",
		ListenPort:         0,     //No DNS, DHCP only
		NotAuthoritative:   false, //be authoritative
		Ranges: []*pb.DHCPServerRange{
			&pb.DHCPServerRange{RangeLower: "172.24.0.2", RangeUpper: "172.24.0.20", LeaseTime: "5m"},
		},
		Options: map[uint32]string{
			3:   "", //Disable option: Router
			6:   "", //Disable option: DNS
		},
	}
	return
}

func GetDefaultLEDSettings() (res *pb.LEDSettings) {
	return &pb.LEDSettings{
		BlinkCount: 254,
	}
}

// Note: If no single function is enabled, the gadget mustn't be enabled itself in order to be deployable
func GetDefaultGadgetSettings() (res pb.GadgetSettings) {
	res = pb.GadgetSettings{
		Enabled:          false,
		Vid:              "0x1d6b",
		Pid:              "0x1347",
		Manufacturer:     "MaMe82",
		Product:          "P4wnP1 by MaMe82",
		Serial:           "deadbeef1337",
		Use_CDC_ECM:      false,
		Use_RNDIS:        false,
		Use_HID_KEYBOARD: false,
		Use_HID_MOUSE:    false,
		Use_HID_RAW:      false,
		Use_UMS:          false,
		Use_SERIAL:       false,
		RndisSettings: &pb.GadgetSettingsEthernet{
			HostAddr: DEFAULT_RNDIS_HOST_ADDR,
			DevAddr:  DEFAULT_RNDIS_DEV_ADDR,
		},
		CdcEcmSettings: &pb.GadgetSettingsEthernet{
			HostAddr: DEFAULT_CDC_ECM_HOST_ADDR,
			DevAddr:  DEFAULT_CDC_ECM_DEV_ADDR,
		},
		UmsSettings: &pb.GadgetSettingsUMS{
			File:"", //we don't supply an image file, which is no problem as it could be applied later on (removable media)
			Cdrom:false, //By default we don't emulate a CD drive, but a flashdrive
		},
	}

	return res
}

func GetDefaultWiFiSettings() (res *pb.WiFiSettings) {
	res = &pb.WiFiSettings{
		Mode: pb.WiFiSettings_AP,
		AuthMode: pb.WiFiSettings_WPA2_PSK,
		Disabled: false,
		Reg: "US",
		ApChannel: 6,
		ApHideSsid: false,
		BssCfgAP: &pb.BSSCfg{
			SSID: "P4wnP1",
			PSK: "MaMe82-P4wnP1",
		},
		DisableNexmon: true,
		BssCfgClient: nil, //not needed
	}
	return
}
