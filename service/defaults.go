package service

import (
	pb "../proto"
)

func GetDefaultNetworkSettingsUSB() (*pb.EthernetInterfaceSettings) {
	//configure 172.24.0.1/255.255.255.252 for usbeth
	ifSettings := &pb.EthernetInterfaceSettings {
		Enabled:            false,
		Name:               USB_ETHERNET_BRIDGE_NAME,
		IpAddress4:         "172.16.0.1",
		Netmask4:           "255.255.255.252",
		Mode:               pb.EthernetInterfaceSettings_MANUAL,
		DhcpServerSettings: GetDefaultDHCPConfigUSB(),
	}
	return ifSettings
}

func GetDefaultDHCPConfigUSB() (settings *pb.DHCPServerSettings) {
	settings = &pb.DHCPServerSettings{
		CallbackScript:     "/bin/evilscript",
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

func GetDefaultLEDSettings() (res pb.LEDSettings) {
	return pb.LEDSettings{
		BlinkCount: 254,
	}
}

func GetDefaultGadgetSettings() (res pb.GadgetSettings) {
	res = pb.GadgetSettings{
		Enabled:          false,
		Vid:              "0x1d6b",
		Pid:              "0x1337",
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
			HostAddr: "42:63:65:12:34:56",
			DevAddr:  "42:63:65:56:34:12",
		},
		CdcEcmSettings: &pb.GadgetSettingsEthernet{
			HostAddr: "42:63:66:12:34:56",
			DevAddr:  "42:63:66:56:34:12",
		},
		UmsSettings: &pb.GadgetSettingsUMS{
			File:"", //we don't supply an image file, which is no problem as it could be applied later on (removable media)
			Cdrom:false, //By default we don't emulate a CD drive, but a flashdrive
		},
	}

	return res
}