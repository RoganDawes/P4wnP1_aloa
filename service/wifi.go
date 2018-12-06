// +build linux

package service

import (
	"github.com/mame82/P4wnP1_aloa/common_web"
	"github.com/mame82/P4wnP1_aloa/netlink"
	pb "github.com/mame82/P4wnP1_aloa/proto"
	"sync"
	"os/exec"
	"github.com/mame82/P4wnP1_aloa/service/util"
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
	wifi_if_name                   string = "wlan0"
	WPA_SUPPLICANT_CONNECT_TIMEOUT        = time.Second * 20
	HOSTAPD_WAIT_AP_UP_TIMEOUT        = time.Second * 8
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
	RootSvc *Service

	State *pb.WiFiState
	//Settings *pb.WiFi2Settings

	mutexSettings           *sync.Mutex              // Lock settings on change
	CmdWpaSupplicant        *exec.Cmd                //Manages wpa-supplicant process
	mutexWpaSupplicant      *sync.Mutex              //mutex for wpa-supplicant proc
	CmdHostapd              *exec.Cmd                //Manages hostapd process
	mutexHostapd            *sync.Mutex              //hostapd proc lock
	IfaceName               string                   //Name of WiFi interface
	PathWpaSupplicantConf   string                   // path to config file for wpa-supplicant
	PathHostapdConf         string                   // path to config file for hostapd
	LoggerHostapd           *util.TeeLogger          //logger for hostapd
	LoggerWpaSupplicant     *util.TeeLogger          //logger for WPA supplicant
	OutMonitorWpaSupplicant *wpaSupplicantOutMonitor //Monitors wpa_supplicant output and sets signals where needed
	OutMonitorHostapd       *hostapdOutMonitor //Monitors hostapd output and sets signals where needed
}

func (wSvc *WiFiService) StartHostapd(timeout time.Duration) (err error) {
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
	// the logger is the signal generator for OutMonitorHostapd, so we reset the signal before applying the logger
	wSvc.OutMonitorHostapd.resultReceived.Reset()
	wSvc.CmdHostapd.Stdout = wSvc.LoggerHostapd.LogWriter
	wSvc.CmdHostapd.Stderr = wSvc.LoggerHostapd.LogWriter
	err = wSvc.CmdHostapd.Start()
	if err != nil {
		wSvc.CmdHostapd.Wait()
		return errors.New(fmt.Sprintf("Error starting hostapd '%v'", err))
	}

	//wait for result in output
	apUp, errcon := wSvc.OutMonitorHostapd.WaitConnectResultOnce(timeout)
	if errcon != nil {
		log.Printf("... hostapd reached timeout of '%v' without beeing able to bring up an Access Point\n", timeout)
		// avoid dead lock
		wSvc.mutexHostapd.Unlock()
		wSvc.StopHostapd()
		wSvc.mutexHostapd.Lock()
		return errors.New("TIMEOUT REACHED")
	}

	if apUp {
		//We could return success and keep wpa_supplicant running
		log.Printf("... hostapd AP for interface '%s' started\n", wSvc.IfaceName)
		return nil
	} else {
		log.Println("... hostapd failed to bring up an Access Point, stopping ...")
		//wifiStopWpaSupplicant(nameIface)
		// avoid dead lock
		wSvc.mutexHostapd.Unlock()
		wSvc.StopHostapd()
		wSvc.mutexHostapd.Lock()
		log.Println("... hostapd terminated")
		return errors.New("Hostapd failed to bring up Access Point")
	}

	return nil
}

func (wSvc *WiFiService) StopHostapd() (err error) {
	eSuccess := fmt.Sprintf("... hostapd for interface '%s' stopped", wSvc.IfaceName)
	eCantStop := fmt.Sprintf("... couldn't terminate hostapd for interface '%s'", wSvc.IfaceName)

	log.Println("... killing hostapd")
	wSvc.mutexHostapd.Lock()
	defer wSvc.mutexHostapd.Unlock()

	if wSvc.CmdHostapd == nil {
		log.Printf("... hostapd for interface '%s' isn't running, no need to stop it\n", wSvc.IfaceName)
		return nil
	}

	/*
	wSvc.CmdHostapd.Process.Signal(syscall.SIGTERM)
	time.Sleep(time.Millisecond * 500)
	if wSvc.CmdHostapd.ProcessState == nil || !wSvc.CmdHostapd.ProcessState.Exited() {
		wSvc.CmdHostapd.Process.Kill()
		wSvc.CmdHostapd.Wait()

		//wSvc.CmdHostapd.Process.Kill()
		if wSvc.CmdHostapd.ProcessState.Exited() {
			wSvc.CmdHostapd = nil
			log.Println(eSuccess)
			return nil
		} else {
			log.Println(eCantStop)
			return errors.New(eCantStop)
		}

	}
	*/
	err = ProcSoftKill(wSvc.CmdHostapd, time.Second)
	if err != nil { return errors.New(eCantStop) }


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

	/*
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
	*/
	log.Printf("... stopping wpa_supplicant\n", wSvc.IfaceName, wSvc.CmdWpaSupplicant.Process.Pid)
	err = ProcSoftKill(wSvc.CmdWpaSupplicant, time.Second*2)
	if err != nil { return errors.New(eCantStop) }

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
	// the logger is the signal generator for OutMonitorWpaSupplicant, so we reset the signal before apllying the logger
	wSvc.OutMonitorWpaSupplicant.resultReceived.Reset()
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


func MatchGivenBBSToScanResult(scanRes []BSS, targets []*pb.WiFiBSSCfg) (matches []*pb.WiFiBSSCfg) {
	for _, bssCfgTarget := range targets {
		for _, bssCfgScan := range scanRes {
			if bssCfgScan.SSID == bssCfgTarget.SSID {
				log.Printf("Found SSID '%s'\n", bssCfgScan.SSID)
				if len(bssCfgTarget.PSK) == 0 && bssCfgScan.AuthMode != WiFiAuthMode_OPEN {
					log.Printf("No PSK provided for '%s', but authentication mode isn't OPEN. Ignoring this network ...\n", bssCfgScan.SSID)
				} else {
					// SSID match, possible candidate
					matches = append(matches, bssCfgTarget)
				}

			}
		}
	}
	return
}

func (wSvc *WiFiService) runStaMode(newWifiSettings *pb.WiFiSettings) (err error) {
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
	if err != nil {
		return err
	}
	// store config to file
	log.Printf("Creating wpa_supplicant configuration file at '%s'\n", wSvc.PathWpaSupplicantConf)
	err = ioutil.WriteFile(wSvc.PathWpaSupplicantConf, []byte(confstr), os.ModePerm)
	if err != nil {
		return err
	}

	//ToDo: proper error handling, in case connection not possible
	err = wSvc.StartWpaSupplicant(WPA_SUPPLICANT_CONNECT_TIMEOUT)
	if err != nil {
		return err
	}


	return nil
}

// ToDo: Output monitor for AP-ENABLED (same approach as for wpa_supplicant)
func (wSvc *WiFiService) runAPMode(newWifiSettings *pb.WiFiSettings) (err error) {
	//generate hostapd.conf (overwrite old one)
	hostapdCreateConfigFile2(newWifiSettings, wSvc.PathHostapdConf)

	//start hostapd
	err = wSvc.StartHostapd(HOSTAPD_WAIT_AP_UP_TIMEOUT)
	if err != nil {
		fmt.Println("Wait 2 seconds and retry to to start hostapd once...")
		time.Sleep(2 * time.Second)
		err = wSvc.StartHostapd(HOSTAPD_WAIT_AP_UP_TIMEOUT)
		if err != nil {
			return err
		}
	}

	return nil
}

func (wSvc *WiFiService) DeploySettings(newWifiSettings *pb.WiFiSettings) (wstate *pb.WiFiState, err error) {
	log.Println("Deploying new WiFi settings...")
	log.Printf("Settings: %+v\n", newWifiSettings)

	wSvc.mutexSettings.Lock()
	defer wSvc.mutexSettings.Unlock()


	//ToDo: Dis/Enable nexmon if needed

	//stop wpa_supplicant if needed
	err = wSvc.StopWpaSupplicant()
	if err != nil {
		return wSvc.State, err
	}
	//kill hostapd in case it is still running
	err = wSvc.StopHostapd()
	if err != nil {
		return wSvc.State, err
	}

	// Note: When hostapd is killed, the WiFi interface is shut down. If the firmware is reloade (toggeling
	// nexmon) the whole interface gets destroyed. Both have influence on processes depending on the network
	// configuration of the WiFi interface (e.g. DHCP server / client). In order to avoid errors, we reconfigure the
	// WiFi network interface after changing the WiFi state.
	// We do this after setting the respective WiFi mode (STA / AP / FAILOVER), but in order to make this modes
	// work, we enable the interface first - regardless of its 'enabled' state.
	iface, tmperr := net.InterfaceByName(wSvc.IfaceName)
	if tmperr == nil {
		netlink.NetworkLinkUp(iface)
	}

	// Set proper regulatory domain
	errReg := wifiSetReg(newWifiSettings.Regulatory)
	if errReg != nil {
		log.Printf("Error setting WiFi regulatory domain '%s': %v\n", newWifiSettings.Regulatory	, err) //we don't abort on error here
	}

	var triggerEvent *pb.Event = nil
	if !newWifiSettings.Disabled {
		switch newWifiSettings.WorkingMode {
		case pb.WiFiWorkingMode_AP:
			err = wSvc.runAPMode(newWifiSettings)
			// emit Trigger event if AP is Up
			if err == nil {
				triggerEvent = ConstructEventTrigger(common_web.TRIGGER_EVT_TYPE_WIFI_AP_STARTED)
			}
		case pb.WiFiWorkingMode_STA, pb.WiFiWorkingMode_STA_FAILOVER_AP:
			errSta := wSvc.runStaMode(newWifiSettings)
			if errSta == nil {
				triggerEvent = ConstructEventTrigger(common_web.TRIGGER_EVT_TYPE_WIFI_CONNECTED_AS_STA)
			} else {
				//in failover mode, we try to enable AP first
				if newWifiSettings.WorkingMode == pb.WiFiWorkingMode_STA_FAILOVER_AP {
					log.Println(errSta)
					log.Printf("Trying to fail over to Access Point Mode...")
					err = wSvc.runAPMode(newWifiSettings)
					if err == nil {
						triggerEvent = ConstructEventTrigger(common_web.TRIGGER_EVT_TYPE_WIFI_AP_STARTED)
					}
				} else {
					err = errSta
				}
			}
		default:
			// None allowed working mode, so we leave wSvc as UNKNOWN
			err = errors.New("Unknown working mode")
		}
	}
	//fmt.Println("HOSTAPD ERR CHECK\n==============\n-->", err)
	if err == nil {
		log.Printf("... WiFi settings deployed successfully\n")
	} else {
		log.Printf("... deploying WiFi settings failed: %s\n", err.Error())
		return wSvc.State, err
	}

	// At this point, we reestablish the interface settings
	//ReInitNetworkInterface(wSvc.IfaceName)
	if nim,err := wSvc.RootSvc.SubSysNetwork.GetManagedInterface(wSvc.IfaceName); err == nil {
		nim.ReDeploy()
	}



	// update settings (wSvc is updated by runAPMode/runStaMode)
	wSvc.State.CurrentSettings = newWifiSettings

	// update rest of state
	if serr := wSvc.UpdateStateFromIw(); serr != nil {
		log.Println("Couldn't update internal WiFi state:", serr)
	}

	// Fire the event after everything is done, especially after redeployment of the network interface settings
	// to allow an ActionTrigger which deploys another ethernet settings template (without changing the settings
	// in parallel)
	// ToDo: check if it makes sense to lock the methods responsible for deploying ethernet interface settings and WifI settings (both long-running)
	if triggerEvent != nil {
		wSvc.RootSvc.SubSysEvent.Emit(triggerEvent)
	}


	return wSvc.State, nil
}

func NewWifiService(rootSvc *Service) (res *WiFiService) {
	ifName := wifi_if_name
	err := wifiCheckExternalBinaries()
	if err != nil {
		panic(err)
	}

	//Check interface existence
	if exists := CheckInterfaceExistence(ifName); !exists {
		panic(errors.New(fmt.Sprintf("WiFi interface '%s' not present")))
	}

	res = &WiFiService{
		RootSvc: rootSvc,
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
	res.OutMonitorHostapd = NewHostapdOutMonitor()

	res.LoggerHostapd = util.NewTeeLogger(true)
	res.LoggerHostapd.SetPrefix("hostapd: ")
	res.LoggerHostapd.AddOutput(res.OutMonitorHostapd)

	res.LoggerWpaSupplicant = util.NewTeeLogger(true)
	res.LoggerWpaSupplicant.SetPrefix("wpa_supplicant: ")
	res.LoggerWpaSupplicant.AddOutput(res.OutMonitorWpaSupplicant) // add watcher too tee'ed output writers

	// Initial settings and state on service start

	res.State = &pb.WiFiState{
		Mode: pb.WiFiStateMode_STA_NOT_CONNECTED,
		Channel: 0,
		Ssid: "",
	}

	res.State.CurrentSettings = &pb.WiFiSettings{
		Disabled:       false,
		WorkingMode:    pb.WiFiWorkingMode_AP,
		Client_BSSList: []*pb.WiFiBSSCfg{&pb.WiFiBSSCfg{SSID:"", PSK:""}},
		Ap_BSS:         &pb.WiFiBSSCfg{},
	}
	return res
}

// io.Writer firing a signal if predefined output arrives
type hostapdOutMonitor struct {
	resultReceived *util.Signal
	result         bool
	*sync.Mutex
}

func (m *hostapdOutMonitor) Write(p []byte) (n int, err error) {
	// if result already received, the write could exit (early out)
	if m.resultReceived.IsSet() {
		return n, nil
	}

	// check if buffer contains relevant strings (assume write is called line wise by the hosted process
	// otherwise we'd need to utilize an io.Reader
	line := string(p)

	switch {
	case strings.Contains(line, "AP-DISABLED"):
		log.Printf("Starting Access Point failed\n")
		m.Lock()
		defer m.Unlock()
		m.result = false
		m.resultReceived.Set()
	case strings.Contains(line, "AP-ENABLED"):
		log.Printf("Access point is up\n")
		m.Lock()
		defer m.Unlock()
		m.result = true
		m.resultReceived.Set()
	}
	return len(p), nil
}

func (m *hostapdOutMonitor) WaitConnectResultOnce(timeout time.Duration) (connected bool, err error) {
	err = m.resultReceived.WaitTimeout(timeout)
	if err != nil {
		return false, errors.New("Couldn't retrieve hostapd connection state before timeout")
	}

	m.Lock()
	defer m.Unlock()
	connected = m.result
	m.resultReceived.Reset() //Disable result received, for next use
	return
}

func NewHostapdOutMonitor() *hostapdOutMonitor {
	return &hostapdOutMonitor{
		resultReceived: util.NewSignal(false, false),
		Mutex:          &sync.Mutex{},
		result: false,
	}
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
		// CTRL-EVENT-CONNECTED - Connection to 4e:66:41:a0:5b:35 completed
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
		return nil, errors.New(fmt.Sprintf("Error running scan: '%s'\niw output: %s", err, res))
	}

	result, err = ParseIwScan(string(res))

	return
}

func wifiCreateWpaSupplicantConfStringList(bsslist []*pb.WiFiBSSCfg) (config string, err error) {
	// if a PSK is provided, we assume it is needed, otherwise we assume OPEN AUTHENTICATION
	for _, bss := range bsslist {
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

func wifiCreateHostapdConfString(ws *pb.WiFiSettings) (config string, err error) {
	if ws.WorkingMode != pb.WiFiWorkingMode_AP && ws.WorkingMode != pb.WiFiWorkingMode_STA_FAILOVER_AP {
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

	if ws.AuthMode == pb.WiFiAuthMode_WPA2_PSK {
		config += fmt.Sprintf("auth_algs=1\n") //Use WPA authentication
		config += fmt.Sprintf("wpa=2\n")       //Use WPA2
		//ToDo: check if PSK could be provided encrypted
		config += fmt.Sprintf("wpa_key_mgmt=WPA-PSK\n") //Use a pre-shared key

		config += fmt.Sprintf("wpa_passphrase=%s\n", ws.Ap_BSS.PSK) //Set PSK
		config += fmt.Sprintf("rsn_pairwise=CCMP\n")                //Use Use AES, instead of TKIP
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

func hostapdCreateConfigFile2(s *pb.WiFiSettings, filename string) (err error) {
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
}//ToDo: Create netlink based implementation (not relying on 'iw'): low priority

func (wsvc WiFiService) UpdateStateFromIw() (err error) {
	proc := exec.Command("/sbin/iw", "dev", wsvc.IfaceName, "info")
	res, err := proc.CombinedOutput()
	if err != nil {
		return errors.New(fmt.Sprintf("Error fetching wifi info: '%s'\niw output: %s", err, res))
	}


	/*
	AP
	--
	Interface wlan0
	ifindex 2
	wdev 0x1
	addr b8:27:eb:71:bb:bc
	ssid \xf0\x9f\x92\xa5\xf0\x9f\x96\xa5\xf0\x9f\x92\xa5 \xe2\x93\x85\xe2\x9e\x83\xe2\x93\x8c\xe2\x93\x83\xf0\x9f\x85\x9f\xe2\x9d\xb6
	type AP
	wiphy 0
	channel 2 (2417 MHz), width: 20 MHz, center1: 2417 MHz
	txpower 31.00 dBm

	NOT CONNECTED
	-------------
	Interface wlan0
	ifindex 2
	wdev 0x1
	addr b8:27:eb:71:bb:bc
	type managed
	wiphy 0
	channel 2 (2417 MHz), width: 20 MHz, center1: 2417 MHz
	txpower 31.00 dBm

	CONNECTED
	-----------
	Interface wlan0
	ifindex 2
	wdev 0x1
	addr b8:27:eb:71:bb:bc
	ssid WLAN-579086
	type managed
	wiphy 0
	channel 6 (2437 MHz), width: 20 MHz, center1: 2437 MHz
	txpower 31.00 dBm


	 */

	output := string(res)

	//split into BSS sections
	reSsid := regexp.MustCompile("(?m)ssid (.*)\n")
	reMode := regexp.MustCompile("(?m)type (.*)\n")
	reChannel := regexp.MustCompile("(?m)channel ([0-9]+) .*\n")

	strSsid_sub := reSsid.FindStringSubmatch(output)
	strSsid := ""
	if len(strSsid_sub) > 1 {
		unSsid,uerr := strconv.Unquote(fmt.Sprintf("\"%s\"", strSsid_sub[1]))
		if uerr == nil {
			strSsid = unSsid
		} else {
			fmt.Println("Unquote error", uerr)
		}

	}
//	fmt.Printf("SSID: %s\n", strSsid)

	strChannel_sub := reChannel.FindStringSubmatch(output)
	strChannel := "0"
	if len(strChannel_sub) > 1 {
		strChannel = strChannel_sub[1]
	}
//	fmt.Printf("Channel: %s\n", strChannel)

	strMode_sub := reMode.FindStringSubmatch(output)
	strMode := "0"
	if len(strMode_sub) > 1 {
		strMode = strMode_sub[1]
	}

	switch strings.ToLower(strMode) {
	case "ap":
		wsvc.State.Mode = pb.WiFiStateMode_AP_UP
	default:
		wsvc.State.Mode = pb.WiFiStateMode_STA_NOT_CONNECTED
	}

	if len(strSsid) > 0 {
		wsvc.State.Ssid = strSsid
		if wsvc.State.Mode == pb.WiFiStateMode_STA_NOT_CONNECTED {
			wsvc.State.Mode = pb.WiFiStateMode_STA_CONNECTED // when a SSID is present, the wifi interface is connected to an AP
		}
	}

	intCh := 0
	intCh,_ = strconv.Atoi(strChannel)
	wsvc.State.Channel = uint32(intCh)

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

func ProcSoftKill(cmd *exec.Cmd, timeToKill time.Duration) (err error) {
	if cmd.Process == nil {
		// process already dead
		return nil
	}

	//send SIGTERM for softkill
	cmd.Process.Signal(syscall.SIGTERM)

	//we wait for process to exit or issue SIGKILL after timeout
	hasExitted := make(chan interface{},0)
	go func() {
		cmd.Process.Wait() // even if waite ends with error, the process should have died
		//fmt.Println("WAIT RES ENDED")
		close(hasExitted)
	}()

	select {
	case <- hasExitted:
		//fmt.Println("HAS EXITED")
		return nil
	case <-time.After(timeToKill):
		//timeout exceeded, send SIGKILL
		//fmt.Println("TIMEOUT, SENDING SIGKILL")
		cmd.Process.Kill()
		return nil
	}
}