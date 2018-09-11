package service

import (
	pb "github.com/mame82/P4wnP1_go/proto"
	"sync"
	"os/exec"
	"github.com/mame82/P4wnP1_go/service/util"
	"errors"
	"fmt"
	"strings"
	"log"
	"time"
	"syscall"
	"net"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
)

const (
	wifi_if_name string = "wlan0"
	WPA_SUPPLICANT_CONNECT_TIMEOUT = time.Second * 20
)


func wifiCheckExternalBinaries() error {
	if !binaryAvailable("wpa_supplicant") {
		return errors.New("wpa_supplicant seems to be missing, please install it")
	}
	// to create wpa_supplicant.conf
	if !binaryAvailable("wpa_passphrase") {
		return errors.New("wpa_passphrase seems to be missing, please install it")
	}
	if !binaryAvailable("hostapd") {
		return errors.New("hostapd seems to be missing, please install it")
	}
	// for wifiScan
	if !binaryAvailable("iw") {
		return errors.New("The tool 'iw' seems to be missing, please install it")
	}

	return nil
}

type WiFiService struct {
	State *pb.WiFi2State
	Settings *pb.WiFi2Settings

	mutexSettings           *sync.Mutex // Lock settings on change
	CmdWpaSupplicant        *exec.Cmd //Manages wpa-supplicant process
	mutexWpaSupplicant      *sync.Mutex //mutex for wpa-supplicant proc
	CmdHostapd              *exec.Cmd //Manages hostapd process
	mutexHostapd            *sync.Mutex //hostapd proc lock
	IfaceName               string //Name of WiFi interface
	PathWpaSupplicantConf   string // path to config file for wpa-supplicant
	PathHostapdConf         string // path to config file for hostapd
	LoggerHostapd           *util.TeeLogger //logger for hostapd
	LoggerWpaSupplicant     *util.TeeLogger //logger for WPA supplicant
	OutMonitorWpaSupplicant *wpaSupplicantOutMonitor //Monitors wpa_supplicant output and sets signals where needed
}


func (wSvc *WiFiService) StartHostapd() (err error) {
	log.Printf("Starting hostapd for interface '%s'...\n", wSvc.IfaceName)

	wSvc.mutexHostapd.Lock()
	defer wSvc.mutexHostapd.Unlock()

	//stop hostapd if already running
	if wSvc.CmdHostapd != nil {
		// avoid deadlock
		wSvc.mutexHostapd.Unlock()
		wSvc.StopHostapd()
		wSvc.mutexHostapd.Lock()
	}

	wSvc.CmdHostapd = exec.Command("/usr/sbin/hostapd", wSvc.PathHostapdConf)
	wSvc.CmdHostapd.Stdout = wSvc.LoggerHostapd.LogWriter
	wSvc.CmdHostapd.Stderr = wSvc.LoggerHostapd.LogWriter
	err = wSvc.CmdHostapd.Start()
	if err != nil {
		wSvc.CmdHostapd.Wait()
		return errors.New(fmt.Sprintf("Error starting hostapd '%v'", err))
	}
	log.Printf("... hostapd for interface '%s' started\n", wSvc.IfaceName)
	return nil
}

func (wSvc *WiFiService) StopHostapd() (err error) {
	eSuccess := fmt.Sprintf("... hostapd for interface '%s' stopped", wSvc.IfaceName)
	eCantStop := fmt.Sprintf("... couldn't terminate hostapd for interface '%s'", wSvc.IfaceName)

	wSvc.mutexHostapd.Lock()
	defer wSvc.mutexHostapd.Unlock()

	if wSvc.CmdHostapd == nil {
		log.Printf("... hostapd for interface '%s' isn't running, no need to stop it\n", wSvc.IfaceName)
		return nil
	}

	wSvc.CmdHostapd.Process.Signal(syscall.SIGTERM)
	wSvc.CmdHostapd.Wait()
	if !wSvc.CmdHostapd.ProcessState.Exited() {
		log.Printf("... hostapd didn't react on SIGTERM for interface '%s', trying SIGKILL\n", wSvc.IfaceName)
		wSvc.CmdHostapd.Process.Kill()

		time.Sleep(500 * time.Millisecond)
		if wSvc.CmdHostapd.ProcessState.Exited() {
			wSvc.CmdHostapd = nil
			log.Println(eSuccess)
			return nil
		} else {
			log.Println(eCantStop)
			return errors.New(eCantStop)
		}
	}

	wSvc.CmdHostapd = nil
	log.Println(eSuccess)
	return nil
}

func (wSvc *WiFiService) StopWpaSupplicant() (err error) {
	eSuccess := fmt.Sprintf("... wpa_supplicant for interface '%s' stopped", wSvc.IfaceName)
	eCantStop := fmt.Sprintf("... couldn't terminate wpa_supplicant for interface '%s'", wSvc.IfaceName)

	log.Printf("... stop running wpa_supplicant processes for interface '%s'\n", wSvc.IfaceName)

	wSvc.mutexWpaSupplicant.Lock()
	defer wSvc.mutexWpaSupplicant.Unlock()

	if wSvc.CmdWpaSupplicant == nil {
		log.Printf("... wpa_supplicant for interface '%s' wasn't running, no need to stop it\n", wSvc.IfaceName)
		return nil
	}

	log.Printf("... sending SIGTERM for wpa_supplicant on interface '%s' with PID\n", wSvc.IfaceName, wSvc.CmdWpaSupplicant.Process.Pid)
	wSvc.CmdWpaSupplicant.Process.Signal(syscall.SIGTERM)
	wSvc.CmdWpaSupplicant.Wait()
	if !wSvc.CmdWpaSupplicant.ProcessState.Exited() {
		log.Printf("... wpa_supplicant didn't react on SIGTERM for interface '%s', trying SIGKILL\n", wSvc.IfaceName)
		wSvc.CmdWpaSupplicant.Process.Kill()

		time.Sleep(500 * time.Millisecond)
		if wSvc.CmdWpaSupplicant.ProcessState.Exited() {
			wSvc.CmdWpaSupplicant = nil
			log.Println(eSuccess)
			return nil
		} else {
			log.Println(eCantStop)
			return errors.New(eCantStop)
		}
	}

	wSvc.CmdWpaSupplicant = nil
	log.Println(eSuccess)
	return nil
}

func (wSvc *WiFiService) StartWpaSupplicant(timeout time.Duration) (err error) {
	log.Printf("Starting wpa_supplicant for interface '%s'...\n", wSvc.IfaceName)

	wSvc.mutexWpaSupplicant.Lock()
	defer wSvc.mutexWpaSupplicant.Unlock()

	//stop wpa_supplicant if already running
	if wSvc.CmdWpaSupplicant != nil {
		// avoid dead lock
		wSvc.mutexWpaSupplicant.Unlock()
		wSvc.StopWpaSupplicant()
		wSvc.mutexWpaSupplicant.Lock()
	}

	//we monitor output of wpa_supplicant till we are connected, fail due to wrong PSK or timeout is reached
	//Note: PID file creation doesn't work when not started as daemon, so we do it manually, later on
	wSvc.CmdWpaSupplicant = exec.Command("/sbin/wpa_supplicant", "-c", wSvc.PathWpaSupplicantConf, "-i", wSvc.IfaceName)
	wSvc.CmdWpaSupplicant.Stdout = wSvc.LoggerWpaSupplicant.LogWriter

	err = wSvc.CmdWpaSupplicant.Start()
	if err != nil {
		return err
	}

	//wait for result in output
	connected, errcon := wSvc.OutMonitorWpaSupplicant.WaitConnectResultOnce(timeout)
	if errcon != nil {
		log.Printf("... wpa_supplicant reached timeout of '%v' without beeing able to connect to given network\n", timeout)
		log.Println("... killing wpa_supplicant")
		// avoid dead lock
		wSvc.mutexWpaSupplicant.Unlock()
		wSvc.StopWpaSupplicant()
		wSvc.mutexWpaSupplicant.Lock()
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
		wSvc.mutexWpaSupplicant.Unlock()
		wSvc.StopWpaSupplicant()
		wSvc.mutexWpaSupplicant.Lock()
		return errors.New("Wrong PSK")
	}

	return nil
}

func (wSvc *WiFiService) GetState() pb.WiFi2State {
	return *wSvc.State
}

func MatchGivenBBSToScanResult(scanRes []BSS, targets []*pb.WiFi2BSSCfg) (matches []*pb.WiFi2BSSCfg) {
	for _,bssCfgTarget := range targets {
		for _,bssCfgScan := range scanRes {
			if bssCfgScan.SSID == bssCfgTarget.SSID {
				// SSID match, possible candidate
				matches = append(matches, bssCfgTarget)
			}
		}
	}
	return
}

func (wSvc *WiFiService) runStaMode(newWifiSettings *pb.WiFi2Settings) (err error) {
	if len(newWifiSettings.Client_BSSList) == 0 {
		return errors.New("Error: WiFi mode set to station (STA) but no BSS configurations provided")
	}


	//scan for provided wifi
	scanres, err := WifiScan(wSvc.IfaceName)
	if err != nil {
		return errors.New(fmt.Sprintf("Scanning for existing WiFi networks failed: %v", err))
	}

	matchingBssList := MatchGivenBBSToScanResult(scanres, newWifiSettings.Client_BSSList)
	if len(matchingBssList) == 0 {
		return errors.New(fmt.Sprintf("Non of the given SSIDs found during scan\n"))
	}

	// Create config for the remaining networks
	confstr, err := wifiCreateWpaSupplicantConfStringList(matchingBssList)
	if err != nil { return err }
	// store config to file
	log.Printf("Creating wpa_supplicant configuration file at '%s'\n", wSvc.PathWpaSupplicantConf)
	err = ioutil.WriteFile(wSvc.PathWpaSupplicantConf, []byte(confstr), os.ModePerm)
	if err != nil { return err }

	//ToDo: proper error handling, in case connection not possible
	err = wSvc.StartWpaSupplicant(WPA_SUPPLICANT_CONNECT_TIMEOUT)
	if err != nil { return err }

	wSvc.State.Bss.SSID = "unknown SSID"
	wSvc.State.Channel = newWifiSettings.Channel
	wSvc.State.Regulatory = newWifiSettings.Regulatory
	wSvc.State.HideSsid = newWifiSettings.HideSsid
	wSvc.State.WorkingMode = pb.WiFi2WorkingMode_STA
	wSvc.State.Disabled = false



	return nil
}

// ToDo: Output monitor for AP-ENABLED (same approach as for wpa_supplicant)
func (wSvc *WiFiService) runAPMode(newWifiSettings *pb.WiFi2Settings) (err error) {
	//generate hostapd.conf (overwrite old one)
	hostapdCreateConfigFile2(newWifiSettings, wSvc.PathHostapdConf)

	//start hostapd
	err = wSvc.StartHostapd()
	if err != nil {
		wSvc.State.WorkingMode = pb.WiFi2WorkingMode_UNKNOWN
		return err
	}

	// update Connection wSvc
	wSvc.State.Bss.SSID = newWifiSettings.Ap_BSS.SSID
	wSvc.State.Channel = newWifiSettings.Channel
	wSvc.State.Regulatory = newWifiSettings.Regulatory
	wSvc.State.HideSsid = newWifiSettings.HideSsid
	wSvc.State.WorkingMode = pb.WiFi2WorkingMode_AP
	wSvc.State.Disabled = false
	return nil
}


func (wSvc *WiFiService) DeploySettings(newWifiSettings *pb.WiFi2Settings) (wstate *pb.WiFi2State, err error) {
	log.Println("Deploying new WiFi settings...")
	log.Printf("Settings: %+v\n", newWifiSettings)

	wSvc.mutexSettings.Lock()
	defer wSvc.mutexSettings.Unlock()

	// Reset wSvc to unknown, if something goes wrong, there's no wpa_supplicant or hostapd
	wSvc.State.WorkingMode = pb.WiFi2WorkingMode_UNKNOWN

	//ToDo: Dis/Enable nexmon if needed

	//stop wpa_supplicant if needed
	err = wSvc.StopWpaSupplicant()
	if err != nil { return wSvc.State, err}
	//kill hostapd in case it is still running
	err = wSvc.StopHostapd()
	if err != nil { return wSvc.State, err}
	wSvc.State.Disabled = true

	if !newWifiSettings.Disabled {
		switch newWifiSettings.WorkingMode {
		case pb.WiFi2WorkingMode_AP:
			err = wSvc.runAPMode(newWifiSettings)
		case pb.WiFi2WorkingMode_STA, pb.WiFi2WorkingMode_STA_FAILOVER_AP:
			errSta := wSvc.runStaMode(newWifiSettings)
			if errSta != nil {
				//in failover mode, we try to enable AP first
				if newWifiSettings.WorkingMode == pb.WiFi2WorkingMode_STA_FAILOVER_AP {
					log.Println(errSta)
					log.Printf("Trying to fail over to Access Point Mode...")
					err = wSvc.runAPMode(newWifiSettings)
				} else {
					err = errSta
				}
			}
		default:
			// None allowed working mode, so we leave wSvc as UNKNOWN
			err = errors.New("Unknown working mode")
		}
	}

	if err == nil {
		log.Printf("... WiFi settings deployed successfully\n")
	} else {
		log.Printf("... deploying WiFi settings failed: %s\n", err.Error())
	}


	// update settings (wSvc is updated by runAPMode/runStaMode)
	wSvc.Settings = newWifiSettings

	return wSvc.State, nil
}

func NewWifiService() (res *WiFiService) {
	ifName := wifi_if_name
	err := wifiCheckExternalBinaries()
	if err != nil { panic(err) }

	//Check interface existence
	if exists,_ := CheckInterfaceExistence(ifName); !exists {
		panic(errors.New(fmt.Sprintf("WiFi interface '%s' not present")))
	}

	res = &WiFiService{

		mutexSettings:         &sync.Mutex{},
		CmdWpaSupplicant:      nil,
		mutexWpaSupplicant:    &sync.Mutex{},
		CmdHostapd:            nil,
		mutexHostapd:          &sync.Mutex{},
		IfaceName:             ifName,
		PathWpaSupplicantConf: fmt.Sprintf("/tmp/wpa_supplicant_%s.conf", ifName),
		PathHostapdConf:       fmt.Sprintf("/tmp/hostapd_%s.conf", ifName),
	}

	res.OutMonitorWpaSupplicant = NewWpaSupplicantOutMonitor()

	res.LoggerHostapd = util.NewTeeLogger(true)
	res.LoggerHostapd.SetPrefix("hostapd: ")

	res.LoggerWpaSupplicant = util.NewTeeLogger(true)
	res.LoggerWpaSupplicant.SetPrefix("wpa_supplicant: ")
	res.LoggerWpaSupplicant.AddOutput(res.OutMonitorWpaSupplicant) // add watcher too tee'ed output writers

	// Initial settings and state on service start
	res.Settings = &pb.WiFi2Settings{
		Disabled: false,
		WorkingMode: pb.WiFi2WorkingMode_AP,
		Client_BSSList: []*pb.WiFi2BSSCfg{},
		Ap_BSS: &pb.WiFi2BSSCfg{},
	}
	res.State = &pb.WiFi2State{
		Disabled: true,
		WorkingMode: pb.WiFi2WorkingMode_UNKNOWN,
		Bss: &pb.WiFi2BSSCfg{},
	}

	return res
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


type WiFiAuthMode int

const (
	WiFiAuthMode_OPEN WiFiAuthMode = iota
	//WiFiAuthMode_WEP
	WiFiAuthMode_WPA_PSK
	WiFiAuthMode_WPA2_PSK
	WiFiAuthMode_UNSUPPORTED
)

type BSS struct {
	SSID           string
	BSSID          net.HardwareAddr
	Frequency      int
	BeaconInterval time.Duration //carefull, on IE level beacon interval isn't measured in milliseconds
	AuthMode       WiFiAuthMode
	Signal         float32 //Signal strength in dBm
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

func wifiCreateWpaSupplicantConfStringList(bsslist []*pb.WiFi2BSSCfg) (config string, err error) {
	// if a PSK is provided, we assume it is needed, otherwise we assume OPEN AUTHENTICATION
	for _,bss := range bsslist {
		ssid := bss.SSID
		psk := bss.PSK
		if len(psk) > 0 {
			fmt.Println("Connecting WiFi with PSK")
			proc := exec.Command("/usr/bin/wpa_passphrase", ssid, psk)
			cres, err := proc.CombinedOutput()

			if err != nil {
				return "", errors.New(fmt.Sprintf("Error craeting wpa_supplicant.conf for SSID '%s' with PSK '%s': %s", ssid, psk, string(cres)))
			}
			config += string(cres)
		} else {
			fmt.Println("Connecting WiFi with OPEN AUTH")
			config += fmt.Sprintf(
				`network={
			ssid="%s"
			key_mgmt=NONE
		}
`, ssid)

		}

	}

	config = "ctrl_interface=/run/wpa_supplicant\n" + config

	return
}


func wifiCreateHostapdConfString2(ws *pb.WiFi2Settings) (config string, err error) {
	if ws.WorkingMode != pb.WiFi2WorkingMode_AP && ws.WorkingMode != pb.WiFi2WorkingMode_STA_FAILOVER_AP {
		return "", errors.New("Couldn't create hostapd configuration, the settings don't include an AP")
	}

	if ws.Ap_BSS == nil {
		return "", errors.New("WiFiSettings don't contain a BSS configuration for an AP")
	}

	config = fmt.Sprintf("interface=%s\n", wifi_if_name)

	config += fmt.Sprintf("driver=nl80211\n")                            //netlink capable driver
	config += fmt.Sprintf("hw_mode=g\n")                                 //Use 2.4GHz band
	config += fmt.Sprintf("ieee80211n=1\n")                              //Enable 802.111n
	config += fmt.Sprintf("wmm_enabled=1\n")                             //Enable WMM
	config += fmt.Sprintf("ht_capab=[HT40][SHORT-GI-20][DSSS_CCK-40]\n") // 40MHz channels with 20ns guard interval
	config += fmt.Sprintf("macaddr_acl=0\n")                             //Accept all MAC addresses

	config += fmt.Sprintf("ssid=%s\n", ws.Ap_BSS.SSID)
	config += fmt.Sprintf("channel=%d\n", ws.Channel)

	if ws.AuthMode == pb.WiFi2AuthMode_WPA2_PSK {
		config += fmt.Sprintf("auth_algs=1\n") //Use WPA authentication
		config += fmt.Sprintf("wpa=2\n")       //Use WPA2
		//ToDo: check if PSK could be provided encrypted
		config += fmt.Sprintf("wpa_key_mgmt=WPA-PSK\n") //Use a pre-shared key

		config += fmt.Sprintf("wpa_passphrase=%s\n", ws.Ap_BSS.PSK) //Set PSK
		config += fmt.Sprintf("rsn_pairwise=CCMP\n")                  //Use Use AES, instead of TKIP
	} else {
		config += fmt.Sprintf("auth_algs=3\n") //Both, open and shared auth
	}

	if ws.HideSsid {
		config += fmt.Sprintf("ignore_broadcast_ssid=2\n") //Require clients to know the SSID
	} else {
		config += fmt.Sprintf("ignore_broadcast_ssid=0\n") //Send beacons + probes
	}

	return
}

func hostapdCreateConfigFile2(s *pb.WiFi2Settings, filename string) (err error) {
	log.Printf("Creating hostapd configuration file at '%s'\n", filename)
	fileContent, err := wifiCreateHostapdConfString2(s)
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
