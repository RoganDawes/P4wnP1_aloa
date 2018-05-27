package service

import (
	pb "../proto"
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

type BSS struct {
	SSID string
	BSSID net.HardwareAddr
	Frequency int
	BeaconInterval time.Duration //carefull, on IE level beacon interval isn't meassured in milliseconds
	AuthMode WiFiAuthMode
	Signal float32 //Signal strength in dBm
}

func DeployWifiSettings(ws *pb.WiFiSettings) (err error) {
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
			err = wifiStartWpaSupplicant(wifi_if_name)
			if err != nil { return err }
		}



	}

	log.Printf("... WiFi settings deployed successfully\n")
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

func WifiTest() {
	fmt.Println("Hostapd settings:")
	conf, err := wifiCreateHostapdConfString(GetDefaultWiFiSettings())
	if err == nil {
		fmt.Println(conf)
	} else {
		fmt.Printf("Error creating hostapd config: %v\n", err)
	}

	fmt.Println("End of hostapd settings:")
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
	if err != nil { return err}


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

func wifiStartWpaSupplicant(nameIface string) (err error) {
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

	//stop hostapd if already running
	wifiStopWpaSupplicant(nameIface)


	//We use the run command and allow hostapd to daemonize
	//wpa_supplicant -P /tmp/wpa_supplicant.pid -i wlan0 -c /tmp/wpa_supplicant.conf -B
	proc := exec.Command("/sbin/wpa_supplicant", "-B", "-P", pidFileWpaSupplicant(nameIface), "-f", logFileWpaSupplicant(nameIface), "-c", confpath, "-i", nameIface)
	err = proc.Run()
	if err != nil { return err}


	log.Printf("... wpa_supplicant for interface '%s' started\n", nameIface)
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