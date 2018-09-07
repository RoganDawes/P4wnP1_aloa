package service

import (
	pb "github.com/mame82/P4wnP1_go/proto"
	"log"
	"github.com/mame82/P4wnP1_go/netlink"
	"github.com/mame82/P4wnP1_go/service/util"
	"net"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"os"
	"io/ioutil"
	"syscall"
	"time"
	"sync"
	"regexp"
	"strconv"
)

const (
	wifi_if_name string = "wlan0"
)

//ToDo: big to do ... move all the shitty command tool line wrapping/parsing (iw, hostapd, wpa_supplicant etc.) to dedicated netlink/nl80211 implementation
//VERY LOW PRIORITY, as this basically means reimplementing the whole toolset for a way too small benefit

type WiFiAuthMode int

const (
	WiFiAuthMode_OPEN WiFiAuthMode = iota
	//WiFiAuthMode_WEP
	WiFiAuthMode_WPA_PSK
	WiFiAuthMode_WPA2_PSK
	WiFiAuthMode_UNSUPPORTED
)

const (
	WPA_SUPPLICANT_CONNECT_TIMEOUT = time.Second * 20
)

type WifiState struct {
	mutexSettings           *sync.Mutex
	Settings                *pb.WiFiSettings
	CmdWpaSupplicant        *exec.Cmd
	mutexWpaSupplicant      *sync.Mutex
	CmdHostapd              *exec.Cmd
	mutexHostapd            *sync.Mutex
	IfaceName               string
	PathWpaSupplicantConf   string
	PathHostapdConf         string
	LoggerHostapd           *util.TeeLogger
	LoggerWpaSupplicant     *util.TeeLogger
	OutMonitorWpaSupplicant *wpaSupplicantOutMonitor
}

type wpaSupplicantOutMonitor struct {
	resultReceived *util.Signal
	result         bool
	*sync.Mutex
}

func (m *wpaSupplicantOutMonitor) Write(p []byte) (n int, err error) {
	// if result already received, the write could exit (early out)
	if m.resultReceived.IsSet() {
		return n, nil
	}

	// check if buffer contains relevant strings (assume write is called line wise by the hosted process
	// otherwise we'd need to utilize an io.Reader
	line := string(p)

	switch {
	case strings.Contains(line, "WRONG_KEY"):
		log.Printf("Seems the provided PSK doesn't match\n")
		m.Lock()
		defer m.Unlock()
		m.result = false
		m.resultReceived.Set()
	case strings.Contains(line, "CTRL-EVENT-CONNECTED"):
		log.Printf("Connected to target network\n")
		m.Lock()
		defer m.Unlock()
		m.result = true
		m.resultReceived.Set()
	}
	return len(p), nil
}

func (m *wpaSupplicantOutMonitor) WaitConnectResultOnce(timeout time.Duration) (connected bool, err error) {
	err = m.resultReceived.WaitTimeout(timeout)
	if err != nil {
		return false, errors.New("Couldn't retrieve wpa_supplicant connection state before timeout")
	}

	m.Lock()
	defer m.Unlock()
	connected = m.result
	m.resultReceived.Reset() //Disable result received, for next use
	return
}

func NewWpaSupplicantOutMonitor() *wpaSupplicantOutMonitor {
	return &wpaSupplicantOutMonitor{
		resultReceived: util.NewSignal(false, false),
		Mutex:          &sync.Mutex{},
	}
}

func NewWifiState(startupSettings *pb.WiFiSettings, ifName string) (res *WifiState) {
	if !binaryAvailable("wpa_supplicant") {
		panic("wpa_supplicant seems to be missing, please install it")
	}
	// to create wpa_supplicant.conf
	if !binaryAvailable("wpa_passphrase") {
		panic("wpa_passphrase seems to be missing, please install it")
	}
	if !binaryAvailable("hostapd") {
		panic("hostapd seems to be missing, please install it")
	}
	// for wifiScan
	if !binaryAvailable("iw") {
		panic("The tool 'iw' seems to be missing, please install it")
	}

	res = &WifiState{
		mutexSettings:         &sync.Mutex{},
		IfaceName:             ifName,
		Settings:              startupSettings,
		CmdWpaSupplicant:      nil,
		mutexWpaSupplicant:    &sync.Mutex{},
		CmdHostapd:            nil,
		mutexHostapd:          &sync.Mutex{},
		PathWpaSupplicantConf: fmt.Sprintf("/tmp/wpa_supplicant_%s.conf", ifName),
		PathHostapdConf:       fmt.Sprintf("/tmp/hostapd_%s.conf", ifName),
	}

	res.OutMonitorWpaSupplicant = NewWpaSupplicantOutMonitor()
	res.LoggerHostapd = util.NewTeeLogger(true)
	res.LoggerHostapd.SetPrefix("hostapd: ")
	res.LoggerWpaSupplicant = util.NewTeeLogger(true)
	res.LoggerWpaSupplicant.SetPrefix("wpa_supplicant: ")
	res.LoggerWpaSupplicant.AddOutput(res.OutMonitorWpaSupplicant) // add watcher too tee'ed output writers

	return
}

func (wifiState *WifiState) StartHostapd() (err error) {
	log.Printf("Starting hostapd for interface '%s'...\n", wifiState.IfaceName)

	wifiState.mutexHostapd.Lock()
	defer wifiState.mutexHostapd.Unlock()

	//check if interface is valid
	if_exists, _ := CheckInterfaceExistence(wifiState.IfaceName)
	if !if_exists {
		return errors.New(fmt.Sprintf("The given interface '%s' doesn't exist", wifiState.IfaceName))
	}

	//stop hostapd if already running
	if wifiState.CmdHostapd != nil {
		// avoid deadlock
		wifiState.mutexHostapd.Unlock()
		wifiState.StopHostapd()
		wifiState.mutexHostapd.Lock()
	}

	//We use the run command and allow hostapd to daemonize
	//wifiState.CmdHostapd = exec.Command("/usr/sbin/hostapd", "-f", logFileHostapd(wifiState.IfaceName), wifiState.PathHostapdConf)
	wifiState.CmdHostapd = exec.Command("/usr/sbin/hostapd", wifiState.PathHostapdConf)
	wifiState.CmdHostapd.Stdout = wifiState.LoggerHostapd.LogWriter
	wifiState.CmdHostapd.Stderr = wifiState.LoggerHostapd.LogWriter
	err = wifiState.CmdHostapd.Start()
	if err != nil {
		//bytes, _ := wifiState.CmdHostapd.CombinedOutput()
		//println(string(bytes))
		wifiState.CmdHostapd.Wait()
		return errors.New(fmt.Sprintf("Error starting hostapd '%v'", err))
	}
	log.Printf("... hostapd for interface '%s' started\n", wifiState.IfaceName)
	return nil
}

func (wifiState *WifiState) StopHostapd() (err error) {
	eSuccess := fmt.Sprintf("... hostapd for interface '%s' stopped", wifiState.IfaceName)
	eCantStop := fmt.Sprintf("... couldn't terminate hostapd for interface '%s'", wifiState.IfaceName)

	wifiState.mutexHostapd.Lock()
	defer wifiState.mutexHostapd.Unlock()

	if wifiState.CmdHostapd == nil {
		log.Printf("... hostapd for interface '%s' isn't running, no need to stop it\n", wifiState.IfaceName)
		return nil
	}

	wifiState.CmdHostapd.Process.Signal(syscall.SIGTERM)
	wifiState.CmdHostapd.Wait()
	if !wifiState.CmdHostapd.ProcessState.Exited() {
		log.Printf("... hostapd didn't react on SIGTERM for interface '%s', trying SIGKILL\n", wifiState.IfaceName)
		wifiState.CmdHostapd.Process.Kill()

		time.Sleep(500 * time.Millisecond)
		if wifiState.CmdHostapd.ProcessState.Exited() {
			wifiState.CmdHostapd = nil
			log.Println(eSuccess)
			return nil
		} else {
			log.Println(eCantStop)
			return errors.New(eCantStop)
		}
	}

	wifiState.CmdHostapd = nil
	log.Println(eSuccess)
	return nil
}

func (wifiState *WifiState) StopWpaSupplicant() (err error) {
	eSuccess := fmt.Sprintf("... wpa_supplicant for interface '%s' stopped", wifiState.IfaceName)
	eCantStop := fmt.Sprintf("... couldn't terminate wpa_supplicant for interface '%s'", wifiState.IfaceName)

	log.Printf("... stop running wpa_supplicant processes for interface '%s'\n", wifiState.IfaceName)

	wifiState.mutexWpaSupplicant.Lock()
	defer wifiState.mutexWpaSupplicant.Unlock()

	if wifiState.CmdWpaSupplicant == nil {
		log.Printf("... wpa_supplicant for interface '%s' wasn't running, no need to stop it\n", wifiState.IfaceName)
		return nil
	}

	log.Printf("... sending SIGTERM for wpa_supplicant on interface '%s' with PID\n", wifiState.IfaceName, wifiState.CmdWpaSupplicant.Process.Pid)
	wifiState.CmdWpaSupplicant.Process.Signal(syscall.SIGTERM)
	wifiState.CmdWpaSupplicant.Wait()
	if !wifiState.CmdWpaSupplicant.ProcessState.Exited() {
		log.Printf("... wpa_supplicant didn't react on SIGTERM for interface '%s', trying SIGKILL\n", wifiState.IfaceName)
		wifiState.CmdWpaSupplicant.Process.Kill()

		time.Sleep(500 * time.Millisecond)
		if wifiState.CmdWpaSupplicant.ProcessState.Exited() {
			wifiState.CmdWpaSupplicant = nil
			log.Println(eSuccess)
			return nil
		} else {
			log.Println(eCantStop)
			return errors.New(eCantStop)
		}
	}

	wifiState.CmdWpaSupplicant = nil
	log.Println(eSuccess)
	return nil
}

func (wifiState *WifiState) StartWpaSupplicant(timeout time.Duration) (err error) {
	log.Printf("Starting wpa_supplicant for interface '%s'...\n", wifiState.IfaceName)

	wifiState.mutexWpaSupplicant.Lock()
	defer wifiState.mutexWpaSupplicant.Unlock()

	//check if interface is valid
	if_exists, _ := CheckInterfaceExistence(wifiState.IfaceName)
	if !if_exists {
		return errors.New(fmt.Sprintf("The given interface '%s' doesn't exist", wifiState.IfaceName))
	}

	//stop wpa_supplicant if already running
	if wifiState.CmdWpaSupplicant != nil {
		// avoid dead lock
		wifiState.mutexWpaSupplicant.Unlock()
		wifiState.StopWpaSupplicant()
		wifiState.mutexWpaSupplicant.Lock()
	}

	//we monitor output of wpa_supplicant till we are connected, fail due to wrong PSK or timeout is reached
	//Note: PID file creation doesn't work when not started as daemon, so we do it manually, later on
	wifiState.CmdWpaSupplicant = exec.Command("/sbin/wpa_supplicant", "-c", wifiState.PathWpaSupplicantConf, "-i", wifiState.IfaceName)
	wifiState.CmdWpaSupplicant.Stdout = wifiState.LoggerWpaSupplicant.LogWriter

	err = wifiState.CmdWpaSupplicant.Start()
	if err != nil {
		return err
	}

	//wait for result in output
	connected, errcon := wifiState.OutMonitorWpaSupplicant.WaitConnectResultOnce(timeout)
	if errcon != nil {
		log.Printf("... wpa_supplicant reached timeout of '%v' without beeing able to connect to given network\n", timeout)
		log.Println("... killing wpa_supplicant")
		// avoid dead lock
		wifiState.mutexWpaSupplicant.Unlock()
		wifiState.StopWpaSupplicant()
		wifiState.mutexWpaSupplicant.Lock()
		return errors.New("TIMEOUT REACHED")
	}
	if connected {
		//We could return success and keep wpa_supplicant running
		log.Println("... connected to given WiFi network, wpa_supplicant running")
		return nil
	} else {
		//we stop wpa_supplicant and return err
		log.Println("... seems the wrong PSK was provided for the given WiFi network, stopping wpa_supplicant ...")
		//wifiStopWpaSupplicant(nameIface)
		log.Println("... killing wpa_supplicant")
		// avoid dead lock
		wifiState.mutexWpaSupplicant.Unlock()
		wifiState.StopWpaSupplicant()
		wifiState.mutexWpaSupplicant.Lock()
		return errors.New("Wrong PSK")
	}

	return nil
}

type BSS struct {
	SSID           string
	BSSID          net.HardwareAddr
	Frequency      int
	BeaconInterval time.Duration //carefull, on IE level beacon interval isn't meassured in milliseconds
	AuthMode       WiFiAuthMode
	Signal         float32 //Signal strength in dBm
}

func (state WifiState) GetDeployWifiSettings() (ws *pb.WiFiSettings, err error) {
	return state.Settings, nil
}

func (state *WifiState) DeployWifiSettings(newWifiSettings *pb.WiFiSettings) (err error) {
	log.Printf("Trying to deploy WiFi settings:\n%v\n", newWifiSettings)
	ifName := wifi_if_name

	state.mutexSettings.Lock()
	defer state.mutexSettings.Unlock()

	//Get Interface
	iface, err := net.InterfaceByName(ifName)
	if err != nil {
		return errors.New(fmt.Sprintf("No WiFi interface present: %v\n", err))
	}

	firmwareChange := false
	if newWifiSettings.DisableNexmon {
		//load legacy driver + firmware
		if wifiIsNexmonLoaded() {
			err = wifiLoadLegacy()
			if err != nil {
				return
			}
			firmwareChange = true
		}
	} else {
		//load nexmon driver + firmware
		if !wifiIsNexmonLoaded() {
			err = wifiLoadNexmon()
			if err != nil {
				return
			}
			firmwareChange = true
		}
	}

	if firmwareChange {
		ReInitNetworkInterface(ifName)
	}

	linkStateChange := false
	currentlyEnabled, errstate := netlink.NetworkLinkGetStateUp(iface)
	if errstate != nil {
		linkStateChange = true
	} // current link state couldn't be retireved, regard as changed
	if currentlyEnabled == newWifiSettings.Disabled {
		linkStateChange = true
	} //Is disabled and should be enabled, or the other way around
	if linkStateChange || firmwareChange { // Enable/Disable if only if needed
		// ToDo: the new interface state isn't reflected to respective ethernet settings
		if newWifiSettings.Disabled {
			log.Printf("Setting WiFi interface %s to DOWN\n", iface.Name)
			err = netlink.NetworkLinkDown(iface)
		} else {
			log.Printf("Setting WiFi interface %s to UP\n", iface.Name)
			err = netlink.NetworkLinkUp(iface)
		}
	}

	//set proper regulatory dom
	err = wifiSetReg(newWifiSettings.Reg)
	if err != nil {
		log.Printf("Error setting WiFi regulatory domain '%s': %v\n", newWifiSettings.Reg, err) //we don't abort on error here
	}

	//stop wpa_supplicant if needed
	state.StopWpaSupplicant()
	//kill hostapd in case it is still running
	err = state.StopHostapd()
	if err != nil {
		return err // ToDo: returning at this point is a bit harsh
	}

	switch newWifiSettings.Mode {
	case pb.WiFiSettings_AP:
		//generate hostapd.conf (overwrite old one)
		hostapdCreateConfigFile(newWifiSettings, state.PathHostapdConf)

		//start hostapd
		err = state.StartHostapd()
		if err != nil {
			return err
		}
	case pb.WiFiSettings_STA:
		if newWifiSettings.BssCfgClient == nil {
			return errors.New("Error: WiFi mode set to station (STA) but no BSS configuration for target WiFi provided")
		}
		if len(newWifiSettings.BssCfgClient.SSID) == 0 {
			return errors.New("Error: WiFi mode set to station (STA) but no SSID provided to identify BSS to join")
		}

		//scan for provided wifi
		scanres, err := WifiScan(ifName)
		if err != nil {
			return errors.New(fmt.Sprintf("Scanning for existing WiFi networks failed: %v", err))
		}
		var matchingBss *BSS = nil
		for _, bss := range scanres {
			if bss.SSID == newWifiSettings.BssCfgClient.SSID {
				matchingBss = &bss
				break
			}
		}
		if matchingBss == nil {
			return errors.New(fmt.Sprintf("SSID not found during scan: '%s'", newWifiSettings.BssCfgClient.SSID))
		}

		if len(newWifiSettings.BssCfgClient.PSK) == 0 && matchingBss.AuthMode != WiFiAuthMode_OPEN {
			//seems we try to connect an OPEN AUTHENTICATION network, but the existing BSS isn't OPEN AUTH
			return errors.New(fmt.Sprintf("WiFi SSID '%s' found during scan, but authentication mode isn't OPEN and no PSK was provided", newWifiSettings.BssCfgClient.SSID))
		} else {
			err = WifiCreateWpaSupplicantConfigFile(newWifiSettings.BssCfgClient.SSID, newWifiSettings.BssCfgClient.PSK, state.PathWpaSupplicantConf)
			if err != nil {
				return err
			}
			//ToDo: proper error handling, in case connection not possible
			err = state.StartWpaSupplicant(WPA_SUPPLICANT_CONNECT_TIMEOUT)
			if err != nil {
				return err
			}
		}

	}

	log.Printf("... WiFi settings deployed successfully, checking for stored interface configuration...\n")

	// store new state
	state.Settings = newWifiSettings

	return nil
}

//check if nexmon driver + firmware is active is loaded
func wifiIsNexmonLoaded() bool {
	return true
}

func wifiLoadNexmon() error {
	log.Println("Loading nexmon WiFi firmware")
	return nil
}

func wifiLoadLegacy() error {
	log.Println("Loading leagcy WiFi firmware")
	return nil
}

func wifiSetReg(reg string) (err error) {
	if len(reg) == 0 {
		reg = "US" //default
		log.Printf("No ISO/IEC 3166-1 alpha2 regulatory domain provided, defaulting to '%s'\n", reg)
	}

	reg = strings.ToUpper(reg)

	proc := exec.Command("/sbin/iw", "reg", "set", reg)
	err = proc.Run()
	if err != nil {
		return err
	}

	log.Printf("Notified kernel to use ISO/IEC 3166-1 alpha2 regulatory domain '%s'\n", reg)
	return nil
}

func WifiScan(ifName string) (result []BSS, err error) {
	proc := exec.Command("/sbin/iw", ifName, "scan")
	res, err := proc.CombinedOutput()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error running scan: '%s'\niw outpur: %s", err, res))
	}

	result, err = ParseIwScan(string(res))

	return
}

func WifiCreateWpaSupplicantConfigFile(ssid string, psk string, filename string) (err error) {
	log.Printf("Creating wpa_suuplicant configuration file at '%s'\n", filename)
	fileContent, err := wifiCreateWpaSupplicantConfString(ssid, psk)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(filename, []byte(fileContent), os.ModePerm)
	return
}

func wifiCreateWpaSupplicantConfString(ssid string, psk string) (config string, err error) {
	// if a PSK is provided, we assume it is needed, otherwise we assum OPEN AUTHENTICATION
	if len(psk) > 0 {
		fmt.Println("Connecting WiFi with PSK")
		proc := exec.Command("/usr/bin/wpa_passphrase", ssid, psk)
		cres, err := proc.CombinedOutput()

		if err != nil {
			return "", errors.New(fmt.Sprintf("Error craeting wpa_supplicant.conf for SSID '%s' with PSK '%s': %s", ssid, psk, string(cres)))
		}
		config = string(cres)
	} else {
		fmt.Println("Connecting WiFi with OPEN AUTH")
		config = fmt.Sprintf(
			`network={
			ssid="%s"
			key_mgmt=NONE
		}`, ssid)

	}

	config = "ctrl_interface=/run/wpa_supplicant\n" + config

	return
}

func wifiCreateHostapdConfString(ws *pb.WiFiSettings) (config string, err error) {
	if ws.Mode != pb.WiFiSettings_STA_FAILOVER_AP && ws.Mode != pb.WiFiSettings_AP {
		return "", errors.New("WiFiSettings don't use an AP")
	}

	if ws.BssCfgAP == nil {
		return "", errors.New("WiFiSettings don't contain a BSS configuration for an AP")
	}

	config = fmt.Sprintf("interface=%s\n", wifi_if_name)

	config += fmt.Sprintf("driver=nl80211\n")                            //netlink capable driver
	config += fmt.Sprintf("hw_mode=g\n")                                 //Use 2.4GHz band
	config += fmt.Sprintf("ieee80211n=1\n")                              //Enable 802.111n
	config += fmt.Sprintf("wmm_enabled=1\n")                             //Enable WMM
	config += fmt.Sprintf("ht_capab=[HT40][SHORT-GI-20][DSSS_CCK-40]\n") // 40MHz channels with 20ns guard interval
	config += fmt.Sprintf("macaddr_acl=0\n")                             //Accept all MAC addresses

	config += fmt.Sprintf("ssid=%s\n", ws.BssCfgAP.SSID)
	config += fmt.Sprintf("channel=%d\n", ws.ApChannel)

	if ws.AuthMode == pb.WiFiSettings_WPA2_PSK {
		config += fmt.Sprintf("auth_algs=1\n") //Use WPA authentication
		config += fmt.Sprintf("wpa=2\n")       //Use WPA2
		//ToDo: check if PSK could be provided encrypted
		config += fmt.Sprintf("wpa_key_mgmt=WPA-PSK\n") //Use a pre-shared key

		config += fmt.Sprintf("wpa_passphrase=%s\n", ws.BssCfgAP.PSK) //Set PSK
		config += fmt.Sprintf("rsn_pairwise=CCMP\n")                  //Use Use AES, instead of TKIP
	} else {
		config += fmt.Sprintf("auth_algs=3\n") //Both, open and shared auth
	}

	if ws.ApHideSsid {
		config += fmt.Sprintf("ignore_broadcast_ssid=2\n") //Require clients to know the SSID
	} else {
		config += fmt.Sprintf("ignore_broadcast_ssid=0\n") //Send beacons + probes
	}

	return
}

func hostapdCreateConfigFile(s *pb.WiFiSettings, filename string) (err error) {
	log.Printf("Creating hostapd configuration file at '%s'\n", filename)
	fileContent, err := wifiCreateHostapdConfString(s)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(filename, []byte(fileContent), os.ModePerm)
	return
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
		if err != nil {
			return nil, err
		}

		//freq
		strFreq_sub := rp_freq.FindStringSubmatch(strBSS)
		strFreq := "0"
		if len(strFreq_sub) > 1 {
			strFreq = strFreq_sub[1]
		}
		fmt.Printf("Freq: %s\n", strFreq)
		tmpI64, err := strconv.ParseInt(strFreq, 10, 32)
		if err != nil {
			return nil, err
		}
		currentBSS.Frequency = int(tmpI64)

		//ssid
		strSsid_sub := rp_ssid.FindStringSubmatch(strBSS)
		strSSID := ""
		if len(strSsid_sub) > 1 {
			strSSID = strSsid_sub[1]
		}
		fmt.Printf("SSID: '%s'\n", strSSID)
		currentBSS.SSID = strSSID

		//beacon interval
		strBI_sub := rp_beacon_intv.FindStringSubmatch(strBSS)
		strBI := "100"
		if len(strBI_sub) > 1 {
			strBI = strBI_sub[1]
		}
		fmt.Printf("Beacon Interval: %s\n", strBI)
		tmpI64, err = strconv.ParseInt(strBI, 10, 32)
		if err != nil {
			return nil, err
		}
		currentBSS.BeaconInterval = time.Microsecond * time.Duration(tmpI64*1024) //1TU = 1024 microseconds (not 1000)

		//auth type
		//assume OPEN
		//if "WEP: is present assume UNSUPPORTED
		//if "WPA:" is present assume WPA (overwrite WEP/UNSUPPORTED)
		//if "RSN:" is present assume WPA2 (overwrite WPA/UNSUPPORTED)
		//in case of WPA/WPA2 check for presence of "Authentication suites: PSK" to assure PSK support, otherwise assume unsupported (no EAP/CHAP support for now)
		currentBSS.AuthMode = WiFiAuthMode_OPEN
		if rp_WEP.MatchString(strBSS) {
			currentBSS.AuthMode = WiFiAuthMode_UNSUPPORTED
		}
		if rp_WPA.MatchString(strBSS) {
			currentBSS.AuthMode = WiFiAuthMode_WPA_PSK
		}
		if rp_WPA2.MatchString(strBSS) {
			currentBSS.AuthMode = WiFiAuthMode_WPA2_PSK
		}
		if currentBSS.AuthMode == WiFiAuthMode_WPA_PSK || currentBSS.AuthMode == WiFiAuthMode_WPA2_PSK {
			if !rp_PSK.MatchString(strBSS) {
				currentBSS.AuthMode = WiFiAuthMode_UNSUPPORTED
			}
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
		if len(strSignal_sub) > 1 {
			strSignal = strSignal_sub[1]
		}
		tmpFloat, err := strconv.ParseFloat(strSignal, 32)
		if err != nil {
			return nil, err
		}
		currentBSS.Signal = float32(tmpFloat)
		fmt.Printf("Signal: %s dBm\n", strSignal)

		bsslist = append(bsslist, currentBSS)
	}

	return bsslist, nil
}
