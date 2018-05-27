/*
All the ugly stuff which not only depends on Linux (nothing is platform independent here), but
uses external binaries and depends on them (dnsmasq, dhclient, wpa_supplicant, hostapd etc.) ... or even
worse, the external binaries are glued together with /bin/bash tricks.
 */

package service

import (
	"os/exec"
	"strings"
	"fmt"
	"regexp"
	"errors"
	"net"
	"strconv"
	"time"
)

func binaryAvailable(binname string) bool {
	cmd := exec.Command("which", binname)
	out,err := cmd.CombinedOutput()
	if err != nil { return false}
	if len(out) == 0 { return false }

	if strings.Contains(string(out), binname) {
		return true
	}
	return false
}

//ToDo: Create netlink based implementation (not relying on 'iw'): low priority
func ParseIwScan(scanresult string) (bsslist []BSS, err error) {
	//fmt.Printf("Parsing:\n%s\n", scanresult)

	//split into BSS sections
	rp := regexp.MustCompile("(?msU)^BSS.*")
	strBSSList := rp.Split(scanresult, -1)
	if len(strBSSList) < 1 {
		return nil, errors.New("Error parsing iw scan result") //splitting should always result in one element at least
	}

	bsslist = []BSS{}

	rp_bssid := regexp.MustCompile("[0-9a-f]{2}:[0-9a-f]{2}:[0-9a-f]{2}:[0-9a-f]{2}:[0-9a-f]{2}:[0-9a-f]{2}")
	rp_freq := regexp.MustCompile("(?m)freq:\\s*([0-9]{4})")
	rp_ssid := regexp.MustCompile("(?m)SSID:\\s*(.*)\n")
	rp_beacon_intv := regexp.MustCompile("(?m)beacon interval:\\s*([0-9]*)TU")

	rp_WEP := regexp.MustCompile("(?m)WEP:")
	rp_WPA := regexp.MustCompile("(?m)WPA:")
	rp_WPA2 := regexp.MustCompile("(?m)RSN:")
	rp_PSK := regexp.MustCompile("(?m)Authentication suites: PSK") //ToDo: check if PSK occurs under respective IE (RSN or WPA, when either is chosen)

	//signal: -75.00 dBm
	rp_signal := regexp.MustCompile("(?m)signal:\\s*(-?[0-9]*\\.[0-9]*)")
	for idx, strBSS := range strBSSList[1:] {
		currentBSS := BSS{}
		//fmt.Printf("BSS %d\n================\n%s\n", idx, strBSS)
		fmt.Printf("BSS %d\n================\n", idx)

		//BSSID (should be in first line)
		strBSSID := rp_bssid.FindString(strBSS)
		fmt.Printf("BSSID: %s\n", strBSSID)
		currentBSS.BSSID, err = net.ParseMAC(strBSSID)
		if err != nil { return nil,err}

		//freq
		strFreq_sub := rp_freq.FindStringSubmatch(strBSS)
		strFreq := "0"
		if len(strFreq_sub) > 1 { strFreq = strFreq_sub[1]}
		fmt.Printf("Freq: %s\n", strFreq)
		tmpI64, err := strconv.ParseInt(strFreq, 10,32)
		if err != nil { return nil, err }
		currentBSS.Frequency = int(tmpI64)

		//ssid
		strSsid_sub := rp_ssid.FindStringSubmatch(strBSS)
		strSSID := ""
		if len(strSsid_sub) > 1 { strSSID = strSsid_sub[1]}
		fmt.Printf("SSID: '%s'\n", strSSID)
		currentBSS.SSID = strSSID

		//beacon interval
		strBI_sub := rp_beacon_intv.FindStringSubmatch(strBSS)
		strBI := "100"
		if len(strBI_sub) > 1 { strBI = strBI_sub[1]}
		fmt.Printf("Beacon Interval: %s\n", strBI)
		tmpI64, err = strconv.ParseInt(strBI, 10,32)
		if err != nil { return nil, err }
		currentBSS.BeaconInterval = time.Microsecond * time.Duration(tmpI64 * 1024) //1TU = 1024 microseconds (not 1000)

		//auth type
		//assume OPEN
		//if "WEP: is present assume UNSUPPORTED
		//if "WPA:" is present assume WPA (overwrite WEP/UNSUPPORTED)
		//if "RSN:" is present assume WPA2 (overwrite WPA/UNSUPPORTED)
		//in case of WPA/WPA2 check for presence of "Authentication suites: PSK" to assure PSK support, otherwise assume unsupported (no EAP/CHAP support for now)
		currentBSS.AuthMode = WiFiAuthMode_OPEN
		if rp_WEP.MatchString(strBSS) {currentBSS.AuthMode = WiFiAuthMode_UNSUPPORTED}
		if rp_WPA.MatchString(strBSS) {currentBSS.AuthMode = WiFiAuthMode_WPA_PSK}
		if rp_WPA2.MatchString(strBSS) {currentBSS.AuthMode = WiFiAuthMode_WPA2_PSK}
		if currentBSS.AuthMode == WiFiAuthMode_WPA_PSK || currentBSS.AuthMode == WiFiAuthMode_WPA2_PSK {
			if !rp_PSK.MatchString(strBSS) {currentBSS.AuthMode = WiFiAuthMode_UNSUPPORTED}
		}
		switch currentBSS.AuthMode {
		case WiFiAuthMode_UNSUPPORTED:
			fmt.Println("AuthMode: UNSUPPORTED")
		case WiFiAuthMode_OPEN:
			fmt.Println("AuthMode: OPEN")
		case WiFiAuthMode_WPA_PSK:
			fmt.Println("AuthMode: WPA PSK")
		case WiFiAuthMode_WPA2_PSK:
			fmt.Println("AuthMode: WPA2 PSK")
		}

		//signal
		strSignal_sub := rp_signal.FindStringSubmatch(strBSS)
		strSignal := "0.0"
		if len(strSignal_sub) > 1 { strSignal = strSignal_sub[1]}
		tmpFloat, err := strconv.ParseFloat(strSignal, 32)
		if err != nil { return nil, err }
		currentBSS.Signal = float32(tmpFloat)
		fmt.Printf("Signal: %s dBm\n", strSignal)

		bsslist = append(bsslist, currentBSS)
	}

	return bsslist,nil
}