package service

import (
	pb "github.com/mame82/P4wnP1_go/proto"
	"log"
	"github.com/docker/libcontainer/netlink"
	"net"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"os"
	"io/ioutil"
	"strconv"
	"syscall"
	"time"
	"bufio"
	"io"
	"regexp"
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
	Settings *pb.WiFiSettings
}

type BSS struct {
	SSID string
	BSSID net.HardwareAddr
	Frequency int
	BeaconInterval time.Duration //carefull, on IE level beacon interval isn't meassured in milliseconds
	AuthMode WiFiAuthMode
	Signal float32 //Signal strength in dBm
}

func (state WifiState) GetDeployWifiSettings() (ws *pb.WiFiSettings,err error) {
	return state.Settings, nil
}


func (state *WifiState) DeployWifiSettings(ws *pb.WiFiSettings) (err error) {
	// ToDo: Lock state while setting up

	log.Printf("Trying to deploy WiFi settings:\n%v\n", ws)
	ifName := wifi_if_name

	//Get Interface
	iface, err := net.InterfaceByName(ifName)
	if err != nil {
		return errors.New(fmt.Sprintf("No WiFi interface present: %v\n", err))
	}


	if ws.DisableNexmon {
		//load legacy driver + firmware
		if wifiIsNexmonLoaded() {
			err = wifiLoadLegacy()
			if err != nil {return}
		}
	} else {
		//load nexmon driver + firmware
		if !wifiIsNexmonLoaded() {
			err = wifiLoadNexmon()
			if err != nil {return}
		}
	}

	if ws.Disabled {
		log.Printf("Setting WiFi interface %s to DOWN\n", iface.Name)
		err = netlink.NetworkLinkDown(iface)
	} else {
		log.Printf("Setting WiFi interface %s to UP\n", iface.Name)
		err = netlink.NetworkLinkUp(iface)
	}

	//set proper regulatory dom
	err = wifiSetReg(ws.Reg)
	if err != nil {
		log.Printf("Error setting WiFi regulatory domain '%s': %v\n", ws.Reg, err) //we don't abort on error here
	}

	switch ws.Mode {
	case pb.WiFiSettings_AP:
		//generate hostapd.conf (overwrite old one)
		hostapdCreateConfigFile(ws, confFileHostapd(ifName))

		//start hostapd
		err = wifiStartHostapd(ifName)
		if err != nil { return err }
	case pb.WiFiSettings_STA:
		//kill hostapd in case it is still running
		err = wifiStopHostapd(ifName)
		if err != nil { return err }

		if ws.BssCfgClient == nil  { return errors.New("Error: WiFi mode set to station (STA) but no BSS configuration for target WiFi provided")}
		if len(ws.BssCfgClient.SSID) == 0 { return errors.New("Error: WiFi mode set to station (STA) but no SSID provided to identify BSS to join")}

		//stop wpa_supplicant if needed (avoid conflicts with scanning)
		wifiStopWpaSupplicant(wifi_if_name)

		//scan for provided wifi
		scanres, err := WifiScan(ifName)
		if err != nil {
			return errors.New(fmt.Sprintf("Scanning for existing WiFi networks failed: %v", err))
		}
		var matchingBss *BSS = nil
		for _,bss := range scanres {
			if bss.SSID == ws.BssCfgClient.SSID {
				matchingBss = &bss
				break
			}
		}
		if matchingBss == nil {
			return errors.New(fmt.Sprintf("SSID not found during scan: '%s'", ws.BssCfgClient.SSID))
		}


		if len(ws.BssCfgClient.PSK) == 0 {
			//seems we should connect an OPEN AUTHENTICATION network
			if matchingBss.AuthMode != WiFiAuthMode_OPEN {
				return errors.New(fmt.Sprintf("WiFi SSID '%s' found during scan, but authentication mode isn't OPEN and no PSK was provided", ws.BssCfgClient.SSID))
			}

			//ToDo: try to connect open network
		} else {
			err = WifiCreateWpaSupplicantConfigFile(ws.BssCfgClient.SSID, ws.BssCfgClient.PSK, confFileWpaSupplicant(wifi_if_name))
			if err != nil { return err }
			//ToDo: proper error handling, in case connection not possible
			err = wifiStartWpaSupplicant(wifi_if_name, WPA_SUPPLICANT_CONNECT_TIMEOUT)
			if err != nil { return err }
		}



	}

	log.Printf("... WiFi settings deployed successfully, checking for stored interface configuration...\n")

	// store new state
	state.Settings = ws

	ReInitNetworkInterface(ifName)
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
		reg = "US"  //default
		log.Printf("No ISO/IEC 3166-1 alpha2 regulatory domain provided, defaulting to '%s'\n", reg)
	}

	reg = strings.ToUpper(reg)

	proc := exec.Command("/sbin/iw", "reg", "set", reg)
	err = proc.Run()
	if err != nil { return err}

	log.Printf("Notified kernel to use ISO/IEC 3166-1 alpha2 regulatory domain '%s'\n", reg)
	return nil
}

func WifiScan(ifName string) (result []BSS, err error) {
	if !wifiIwAvailable() { return nil,errors.New("The tool 'iw' is missing, please install it to make this work")}

	proc := exec.Command("/sbin/iw", ifName, "scan")
	res, err := proc.CombinedOutput()
	if err != nil {
		return nil,errors.New(fmt.Sprintf("Error running scan: '%s'\niw outpur: %s", err, res))
	}

	result, err = ParseIwScan(string(res))

	return
}

func WifiCreateWpaSupplicantConfigFile(ssid string, psk string, filename string) (err error) {
	log.Printf("Creating wpa_suuplicant configuration file at '%s'\n", filename)
	fileContent, err := wifiCreateWpaSupplicantConfString(ssid, psk)
	if err != nil {return}
	err = ioutil.WriteFile(filename, []byte(fileContent), os.ModePerm)
	return
}

func wifiCreateWpaSupplicantConfString(ssid string, psk string) (config string, err error) {
	if !wifiWpaPassphraseAvailable() { return "",errors.New("The tool 'wpa_passphrase' is missing, please install it to make this work")}


	proc := exec.Command("/usr/bin/wpa_passphrase", ssid, psk)
	cres, err := proc.CombinedOutput()

	if err != nil {
		return "",errors.New(fmt.Sprintf("Error craeting wpa_supplicant.conf for SSID '%s' with PSK '%s': %s", ssid, psk, string(cres)))
	}
	config = string(cres)
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

	config += fmt.Sprintf("driver=nl80211\n") //netlink capable driver
	config += fmt.Sprintf("hw_mode=g\n") //Use 2.4GHz band
	config += fmt.Sprintf("ieee80211n=1\n") //Enable 802.111n
	config += fmt.Sprintf("wmm_enabled=1\n") //Enable WMM
	config += fmt.Sprintf("ht_capab=[HT40][SHORT-GI-20][DSSS_CCK-40]\n") // 40MHz channels with 20ns guard interval
	config += fmt.Sprintf("macaddr_acl=0\n") //Accept all MAC addresses

	config += fmt.Sprintf("ssid=%s\n", ws.BssCfgAP.SSID)
	config += fmt.Sprintf("channel=%d\n", ws.ApChannel)

	if ws.AuthMode == pb.WiFiSettings_WPA2_PSK {
		config += fmt.Sprintf("auth_algs=1\n") //Use WPA authentication
		config += fmt.Sprintf("wpa=2\n") //Use WPA2
		//ToDo: check if PSK could be provided encrypted
		config += fmt.Sprintf("wpa_key_mgmt=WPA-PSK\n") //Use a pre-shared key

		config += fmt.Sprintf("wpa_passphrase=%s\n", ws.BssCfgAP.PSK) //Set PSK
		config += fmt.Sprintf("rsn_pairwise=CCMP\n") //Use Use AES, instead of TKIP
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
	if err != nil {return}
	err = ioutil.WriteFile(filename, []byte(fileContent), os.ModePerm)
	return
}

func wifiWpaSupplicantAvailable() bool {
	return binaryAvailable("wpa_supplicant")
}

func wifiWpaPassphraseAvailable() bool {
	return binaryAvailable("wpa_passphrase")
}

func wifiHostapdAvailable() bool {
	return binaryAvailable("hostapd")
}

func wifiIwAvailable() bool {
	return binaryAvailable("iw")
}

func wifiStartHostapd(nameIface string) (err error) {
	log.Printf("Starting hostapd for interface '%s'...\n", nameIface)

	//check if interface is valid
	if_exists,_ := CheckInterfaceExistence(nameIface)
	if !if_exists {
		return errors.New(fmt.Sprintf("The given interface '%s' doesn't exist", nameIface))
	}

	if !wifiHostapdAvailable() {
		return errors.New("hostapd seems to be missing, please install it")
	}

	confpath := confFileHostapd(nameIface)

	//stop hostapd if already running
	wifiStopHostapd(nameIface)


	//We use the run command and allow hostapd to daemonize
	proc := exec.Command("/usr/sbin/hostapd", "-B", "-P", pidFileHostapd(nameIface), "-f", logFileHostapd(nameIface), confpath)
	err = proc.Run()
	if err != nil {
		return errors.New(fmt.Sprintf("Error starting hostapd '%v'", err))
	}


	log.Printf("... hostapd for interface '%s' started\n", nameIface)
	return nil
}

func wifiStopHostapd(nameIface string) (err error) {
	log.Printf("... stop running hostapd processes for interface '%s'\n", nameIface)
	running,pid,err := wifiIsHostapdRunning(wifi_if_name)
	if err != nil { return err }
	if !running {
		log.Printf("... hostapd for interface '%s' isn't running, no need to stop it\n", nameIface)
		return nil
	}
	//kill the pid
	err = syscall.Kill(pid, syscall.SIGTERM)
	if err != nil { return }

	time.Sleep(500*time.Millisecond)

	//check if stopped
	running,pid,err = wifiIsHostapdRunning(nameIface)
	if err != nil { return }
	if (running) {
		log.Printf("... couldn't terminate hostapd for interface '%s'\n", nameIface)
	} else {
		log.Printf("... hostapd for interface '%s' stopped\n", nameIface)
	}

	//Delete PID file
	os.Remove(pidFileHostapd(nameIface))

	return nil
}


func wifiWpaSupplicantOutParser(chanResult chan string, reader *bufio.Reader) {
	log.Println("... Start monitoring wpa_supplicant output")

	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			//in case wpa_supplicant is killed, we should land here, which ends the goroutine
			if err != io.EOF {
				log.Printf("Can't read wpa_supplicant output: %s\n", err)
			}

			break
		}
		strLine := string(line)

		//fmt.Printf("Read:\n%s\n", strLine)

		switch {
		case strings.Contains(strLine, "WRONG_KEY"):
			log.Printf("Seems the provided PSK doesn't match\n")
			chanResult <- "WRONG_KEY"
			break
		case strings.Contains(strLine, "CTRL-EVENT-CONNECTED"):
			log.Printf("Connected to target network\n")
			chanResult <- "CONNECTED"
			break // stop loop
		}
	}
	log.Println("... stopped monitoring wpa_supplicant output")
}

func wifiStartWpaSupplicant(nameIface string, timeout time.Duration) (err error) {
	log.Printf("Starting wpa_supplicant for interface '%s'...\n", nameIface)

	//check if interface is valid
	if_exists,_ := CheckInterfaceExistence(nameIface)
	if !if_exists {
		return errors.New(fmt.Sprintf("The given interface '%s' doesn't exist", nameIface))
	}

	if !wifiWpaSupplicantAvailable() {
		return errors.New("wpa_supplicant seems to be missing, please install it")
	}

	confpath := confFileWpaSupplicant(nameIface)

	//stop wpa_supplicant if already running
	wifiStopWpaSupplicant(nameIface)


	//we monitor output of wpa_supplicant till we are connected, fail due to wrong PSK or timeout is reached
	//Note: PID file creation doesn't work when not started as daemon, so we do it manually, later on
	proc := exec.Command("/sbin/wpa_supplicant",  "-P", pidFileWpaSupplicant(nameIface), "-c", confpath, "-i", nameIface)

	wpa_stdout, err := proc.StdoutPipe()
	if err != nil { return err}

	err = proc.Start()
	if err != nil { return err}

	//Create PID file by hand, as we're not running wpa_supplicant in daemon mode
	err = ioutil.WriteFile(pidFileWpaSupplicant(nameIface), []byte(fmt.Sprintf("%d", proc.Process.Pid)), os.ModePerm)

	//result channel
	wpa_res := make(chan string, 1)
	//start output parser
	wpa_stdout_reader := bufio.NewReader(wpa_stdout)

	go wifiWpaSupplicantOutParser(wpa_res, wpa_stdout_reader)

	//analyse output
	select {
		case res := <-wpa_res:
			if strings.Contains(res, "CONNECTED") {
				//We could return success and keep wpa_supplicant running
				log.Println("... connected to given WiFi network, wpa_supplicant running")
				return nil
			}
			if strings.Contains(res, "WRONG_KEY") {
				//we stop wpa_supplicant and return err
				log.Println("... seems the wrong PSK wwas provided for the given WiFi network, stopping wpa_supplicant ...")
				//wifiStopWpaSupplicant(nameIface)
				log.Println("... killing wpa_supplicant")
				proc.Process.Kill()
				return errors.New("Wrong PSK")
			}
		case <- time.After(timeout):
			//we stop wpa_supplicant and return err
			log.Printf("... wpa_supplicant reached timeout of '%v' without beeing able to connect to given network\n", timeout)
			log.Println("... killing wpa_supplicant")
			//wifiStopWpaSupplicant(nameIface)
			proc.Process.Kill()
			return errors.New("TIMEOUT REACHED")
	}


	return nil
}


func wifiStopWpaSupplicant(nameIface string) (err error) {
	log.Printf("... stop running wpa_supplicant processes for interface '%s'\n", nameIface)
	running,pid,err := wifiIsWpaSupplicantRunning(wifi_if_name)
	if err != nil { return err }
	if !running {
		log.Printf("... wpa_supplicant for interface '%s' isn't running, no need to stop it\n", nameIface)
		return nil
	}
	//kill the pid
	err = syscall.Kill(pid, syscall.SIGTERM)
	if err != nil { return }

	time.Sleep(500*time.Millisecond)

	//check if stopped
	running,pid,err = wifiIsHostapdRunning(nameIface)
	if err != nil { return }
	if (running) {
		log.Printf("... couldn't terminate wpa_supplicant for interface '%s'\n", nameIface)
	} else {
		log.Printf("... wpa_supplicant for interface '%s' stopped\n", nameIface)
	}

	//Delete PID file
	os.Remove(pidFileHostapd(nameIface))

	return nil
}

func pidFileHostapd(nameIface string) string {
	return fmt.Sprintf("/var/run/hostapd_%s.pid", nameIface)
}

func logFileHostapd(nameIface string) string {
	return fmt.Sprintf("/tmp/hostapd_%s.log", nameIface)
}

func confFileHostapd(nameIface string) string {
	return fmt.Sprintf("/tmp/hostapd_%s.conf", nameIface)
}


func confFileWpaSupplicant(nameIface string) string {
	return fmt.Sprintf("/tmp/wpa_supplicant_%s.conf", nameIface)
}

func logFileWpaSupplicant(nameIface string) string {
	return fmt.Sprintf("/tmp/wpa_supplicant_%s.log", nameIface)
}

func pidFileWpaSupplicant(nameIface string) string {
	return fmt.Sprintf("/var/run/wpa_supplicant_%s.pid", nameIface)
}


func wifiIsHostapdRunning(nameIface string) (running bool, pid int, err error) {
	pid_file := pidFileHostapd(nameIface)

	//Check if the pidFile exists
	if _, err := os.Stat(pid_file); os.IsNotExist(err) {
		return false, 0,nil //file doesn't exist, so we assume hostapd isn't running
	}

	//File exists, read the PID
	content, err := ioutil.ReadFile(pid_file)
	if err != nil { return false, 0, err}
	pid, err = strconv.Atoi(strings.TrimSuffix(string(content), "\n"))
	if err != nil { return false, 0, errors.New(fmt.Sprintf("Error parsing PID file %s: %v", pid_file, err))}

	//With PID given, check if the process is indeed running (pid_file could stay, even if the hostapd process has died already)
	err_kill := syscall.Kill(pid, 0) //sig 0: doesn't send a signal, but error checking is still performed
	switch err_kill{
	case nil:
		//ToDo: Check if the running process image is indeed hostapd
		return true, pid, nil //Process is running
	case syscall.ESRCH:
		//Process doesn't exist
		return false, pid, nil
	case syscall.EPERM:
		//process exists, but we have no access permission
		return true, pid, err_kill
	default:
		return false, pid, err_kill
	}
}

func wifiIsWpaSupplicantRunning(nameIface string) (running bool, pid int, err error) {
	pid_file := pidFileWpaSupplicant(nameIface)

	//Check if the pidFile exists
	if _, err := os.Stat(pid_file); os.IsNotExist(err) {
		return false, 0,nil //file doesn't exist, so we assume wpa_supplicant isn't running
	}

	//File exists, read the PID
	content, err := ioutil.ReadFile(pid_file)
	if err != nil { return false, 0, err}
	pid, err = strconv.Atoi(strings.TrimSuffix(string(content), "\n"))
	if err != nil { return false, 0, errors.New(fmt.Sprintf("Error parsing PID file %s: %v", pid_file, err))}

	//With PID given, check if the process is indeed running (pid_file could stay, even if the wpa_supplicant process has died already)
	err_kill := syscall.Kill(pid, 0) //sig 0: doesn't send a signal, but error checking is still performed
	switch err_kill{
	case nil:
		//ToDo: Check if the running process image is indeed wpa_supplicant
		return true, pid, nil //Process is running
	case syscall.ESRCH:
		//Process doesn't exist
		return false, pid, nil
	case syscall.EPERM:
		//process exists, but we have no access permission
		return true, pid, err_kill
	default:
		return false, pid, err_kill
	}
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