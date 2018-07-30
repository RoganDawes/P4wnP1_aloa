// +build linux,arm

package main

import (
	"github.com/mame82/P4wnP1_go/service"
	"log"
	pb "github.com/mame82/P4wnP1_go/proto"
	"fmt"
	"time"
)

func main() {
	ap_settings := &pb.WiFiSettings{
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

	sta_settings := &pb.WiFiSettings{
		Mode: pb.WiFiSettings_STA,
		AuthMode: pb.WiFiSettings_WPA2_PSK,
		Disabled: false,
		Reg: "DE",
		BssCfgClient: &pb.BSSCfg{
			SSID: "WLAN-579086",
			PSK: "5824989790864470",
		},
		DisableNexmon: true,
		BssCfgAP: nil, //not needed
	}

	err := service.DeployWifiSettings(ap_settings)
	if err != nil { log.Println(err)}

	err = service.DeployWifiSettings(sta_settings)
	if err != nil { log.Println(err)}

	fmt.Println("Sleeping 10 seconds")
	time.Sleep(time.Second * 10)
}
