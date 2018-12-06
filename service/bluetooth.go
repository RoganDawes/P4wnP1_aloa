// +build linux,arm

package service

import (
	"errors"
	"fmt"
	pb "github.com/mame82/P4wnP1_aloa/proto"
	"github.com/mame82/P4wnP1_aloa/service/bluetooth"
	"github.com/mame82/mblue-toolz/toolz"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	BT_MINIMUM_BLUEZ_VERSION_MAJOR = 5
	BT_MINIMUM_BLUEZ_VERSION_MINOR = 43
)

type BtService struct {
	RootSvc *Service

	serviceAvailable bool
	Controller       *bluetooth.Controller
	BrName           string
	//bridgeIfDeployed bool

	defaultSettings *pb.BluetoothSettings //This settings are changed if the BluetoothService isn't up, but settings are deployed (Master Template on startup)

	Agent *bluetooth.DefaultAgent

	serviceAvailableLock *sync.Mutex
}

//Notes: If a bluetooth controller could be found with `bluetooth.FindFirstAvailableController()`
// this also means that the bluetoothd is running, as the Bluez mgmt-api is used to check for controllers

// P4wnP1 doesn't depend on late starting systemd services like DBus, bluetoothd etc.
// In order to assure that Bluetooth functionality is present, the P4wnP1 systemd service would have to depend
// on such "late starting" services. This again means the P4wnP1 daemon would load very late and functionality like
// USB gadgets wouldn't work till service the P4wnP1 systemd service gets started. Tests with current Kali releases have
// shown that P4wnP1 is up and reachable network after about 20 to 30 seconds.
// If the P4wnP1 systemd service is changed to depend on bluetooth.service this duration increases up to 2 minutes,
// which is NOT ACCEPTABLE.
// On the other hand, access to the bluez-stack and the hci device (if present) is already possible some seconds
// after the P4wnP1 service has started (doesn't take 2 minutes).
// To deal with that, the bluetooth subsystem keeps retrying to find a working bluetooth adapter till "retryTimeout"
// is reached (the argument is handed in to NewBtService). If the hci adapter gets deployed after some seconds (as it
// happens in my tests) this means service startup of P4wnP1 increases by this duration (about 6 seconds in tests)
// On a system with a mis-configured bluetooth stack or with missing bluetooth hardware (Pi0 without WiFi/Bluetooth)
// this would mean P4wnP1 service startup consumes the full retryTime (shouldn't be the case, as we target a RPi0W
// with a custom build Kali image, which assures correct bluetooth stack setup; Pi0 without WiFi isn't supported).
//
// The current behavior could be changed, if NewBtService gets wrapped into a go-routine. The shortcoming would be,
// that every call to to BtService functions (like StartNap) would fail, even if the missing adapter could show up some
// seconds later.
//
// A polished future solution would be, to combine creation of Bluetooth SubSystem and initial configuration deployment
// in a go routine and let RPC calls relying on bluetooth sub system fail, till bluetooth is usable. This would mean that
// an event has to be PUSHED to the webclient, once bluetooth is usable.



func NewBtService(rootService *Service, retryTimeout time.Duration) (res *BtService) {
	res = &BtService{
		RootSvc: rootService,
		Agent:   bluetooth.NewDefaultAgent("1337"),
		BrName:  BT_ETHERNET_BRIDGE_NAME,
		serviceAvailableLock: &sync.Mutex{},
		defaultSettings:GetDefaultBluetoothSettings(),
	}

	log.Println("Starting Bluetooth sub system...")
	if err := CheckBluezVersion(); err != nil {
		log.Println(err)

		res.setServiceAvailable(false)
		return
	}

	go func() {
		timeStart := time.Now()
		for timeSinceStart := time.Since(timeStart); !res.serviceAvailable && (timeSinceStart < retryTimeout); timeSinceStart = time.Since(timeStart) {
			if c, err := bluetooth.FindFirstAvailableController(); err == nil {
				res.setServiceAvailable(true)
				res.Controller = c
				log.Printf("... bluetooth controller found '%s' after %v\n", res.Controller.DBusPath, timeSinceStart)
			} else {
				log.Printf("Re-check bluetooth adapter existence %v\n", timeSinceStart)
				res.setServiceAvailable(false)
			}
			time.Sleep(time.Second * 1)
		}

		if !res.serviceAvailable {
			log.Printf("No bluetooth adapter found after %v\n", retryTimeout)
		} else {
			//register the agent
			res.Agent.Start(toolz.AGENT_CAP_NO_INPUT_NO_OUTPUT)

			// Deploy default settings

			_,err := res.DeployBluetoothControllerInformation(res.defaultSettings.Ci)
			if err != nil {
				log.Println("Not able to deploy default bluetooth settings: ", err.Error())
			} else {
				_,err = res.DeployBluetoothAgentSettings(res.defaultSettings.As)
				if err != nil {
					log.Println("Not able to deploy default bluetooth agent settings: ", err.Error())
				}
			}
			log.Println("Finished setting up bluetooth")
		}
	}()


	return
}

func (bt *BtService) ReplaceDefaultSettings(s *pb.BluetoothSettings) {
	bt.defaultSettings = s
}

func (bt *BtService) Stop() {
	bt.Agent.Stop() // unregister the agent again
	if ci,err := bt.Controller.ReadControllerInformation(); err == nil {
		if ci.ServiceNetworkServerNap {
			bt.UnregisterNetworkServer(toolz.UUID_NETWORK_SERVER_NAP)
		}
		if ci.ServiceNetworkServerGn {
			bt.UnregisterNetworkServer(toolz.UUID_NETWORK_SERVER_GN)
		}
		if ci.ServiceNetworkServerPanu {
			bt.UnregisterNetworkServer(toolz.UUID_NETWORK_SERVER_PANU)
		}
	}
	bt.DisableBridge()
}

func (bt *BtService) setServiceAvailable(val bool)  {
	bt.serviceAvailableLock.Lock()
	defer bt.serviceAvailableLock.Unlock()
	bt.serviceAvailable = val
}

func (bt *BtService) IsServiceAvailable() bool  {
	bt.serviceAvailableLock.Lock()
	defer bt.serviceAvailableLock.Unlock()
	return bt.serviceAvailable
}

func (bt *BtService) DeployBluetoothNetworkService(btNwSvc *pb.BluetoothNetworkService) (err error) {
	uuid := toolz.UUID_NETWORK_SERVER_NAP
	switch btNwSvc.Type {
	case pb.BluetoothNetworkServiceType_NAP:
		uuid = toolz.UUID_NETWORK_SERVER_NAP
	case pb.BluetoothNetworkServiceType_PANU:
		uuid = toolz.UUID_NETWORK_SERVER_PANU
	case pb.BluetoothNetworkServiceType_GN:
		uuid = toolz.UUID_NETWORK_SERVER_GN
	}
	if btNwSvc.ServerOrConnect {
		// start server for given network service
		if btNwSvc.RegisterOrUnregister {
			return bt.RegisterNetworkServer(uuid)
		} else {
			return bt.UnregisterNetworkServer(uuid)
		}
	} else {
		//(dis)connect from/to given network network service of given remote device

		if btNwSvc.RegisterOrUnregister {
			// register == connect
			return bt.ConnectNetwork(btNwSvc.MacOrName, uuid)
		} else {
			// unregister == disconnect
			return bt.DisconnectNetwork(btNwSvc.MacOrName)
		}
	}
}


func (bt *BtService) GetBluetoothAgentSettings() (as *pb.BluetoothAgentSettings, err error) {
	if !bt.IsServiceAvailable() {
		return &pb.BluetoothAgentSettings{},bluetooth.ErrBtSvcNotAvailable
	}
	as = &pb.BluetoothAgentSettings{}

	pin,err := bt.GetPIN()
	if err != nil { return as,err }
	as.Pin = pin
	return
}


func (bt *BtService) DeployBluetoothAgentSettings(src *pb.BluetoothAgentSettings) (res *pb.BluetoothAgentSettings, err error) {
	if !bt.IsServiceAvailable() {
		return &pb.BluetoothAgentSettings{},bluetooth.ErrBtSvcNotAvailable
	}
	res = &pb.BluetoothAgentSettings{}
	err = bt.SetPIN(src.Pin)
	if err != nil { return }
	return bt.GetBluetoothAgentSettings()
}


func (bt *BtService) DeployBluetoothControllerInformation(newBtCiRpc *pb.BluetoothControllerInformation) (updateBtCiRpc *pb.BluetoothControllerInformation, err error) {
	if !bt.IsServiceAvailable() {
		return &pb.BluetoothControllerInformation{},bluetooth.ErrBtSvcNotAvailable
	}

	btCi := bluetooth.BluetoothControllerInformationFromRpc(newBtCiRpc)
	bridgeNameNap := BT_ETHERNET_BRIDGE_NAME
	bridgeNamePanu := BT_ETHERNET_BRIDGE_NAME
	bridgeNameGn := BT_ETHERNET_BRIDGE_NAME

	// Update provided network services if needed
	if btCi.ServiceNetworkServerNap || btCi.ServiceNetworkServerGn || btCi.ServiceNetworkServerPanu {
		err = bt.EnableBridge()
		if err != nil { return &pb.BluetoothControllerInformation{},err }
	} else {
		bt.DisableBridge()
	}

	log.Println("Updating settings from controller information...")
	updatedCi,err := bt.Controller.UpdateSettingsFromChangedControllerInformation(btCi, bridgeNameNap, bridgeNamePanu, bridgeNameGn)
	log.Printf("Deployed bluetooth settings\n%+v\n%v\n", updatedCi, err)
	if err != nil { return &pb.BluetoothControllerInformation{},err }
	updateBtCiRpc = bluetooth.BluetoothControllerInformationToRpc(updatedCi)
	return updateBtCiRpc, nil
}


func (bt *BtService) GetControllerInformation() (ctlInfo *pb.BluetoothControllerInformation ,err error) {
	if !bt.IsServiceAvailable() {
		return &pb.BluetoothControllerInformation{},bluetooth.ErrBtSvcNotAvailable
	}
	btCi,err := bt.Controller.ReadControllerInformation()
	if err != nil { return &pb.BluetoothControllerInformation{},err}
	btCiRpc := bluetooth.BluetoothControllerInformationToRpc(btCi)
	btCiRpc.IsAvailable = bt.IsServiceAvailable()
	return btCiRpc,nil
}


/*
// Notes: On Bluetooth settings
// P4wnP1 is meant to run headless, which has influence on Pairing mode. There's legacy pairing (outdated and insecure)
// which allows requesting a PIN from a remote device which wants to connect. The new Pairing mode is Secure Simple Pairing
// (SSP) which add in dynamic key creation on pairing, without static PINs. There are different ways two devices could be paired,
// the way is chosen depending on the capabilities of both devices. IT ISN'T POSSIBLE TO REQUEST A PREDEFINED PIN WITH
// SECURE SIMPLE PAIRING. Bonding (=Pairing) is handled with a random passkey or in just works mode.
// As P4wnP1 could not display a passkey or request user input for a confirmation (assuming interactive access solutions
// like webclient, cli_client or ssh aren't always used), we have to fall back to "just works" mode if we want to use SSP.
// Even if a static PIN is a security issue, using just works is even more insecure.
// On the other hand, the idea to disable SSP didn't work out either, because this won't allow to set the broadcom bluetooth
// adapter to high speed. Not having high speed enabled, ultimately results in a very slow connection for BNEP usage
// (in fact, if a NAP is turned on without high speed, a remote device is able to pair and connect, even to receive a DHCP
// lease from the server, but follow up traffic way to slow)
//
// Additionally it seems if an Android should be able to use a NAP provided via BNEP, a DHCP server has to be running
// and has to HAND OUT A ROUTER AND A DNS OPTION, POINTING TO THE IP OF P4wnP1.
// Even if no upstream connection is provided, connecting to P4wnP1 is possible, if, and only if, these two DHCP options
// are set. Otherwise the Adnroid device would stop communicating after the DHCP lease has been issued.
// The behavior of not fully working connections has been observed using a Samsung Android phone as remote device. It has
// not been confirmed that connection issues exist on other devices.
//
// Summary of NAP conditions:
// - DHCP server is running and provides DHCP option 3 and 6, both pointing to the IP of the bluetooth ethernet bridge
// - to be able to use high speed, SSP has to be enabled, which again means no PIN requests are possible
// - if SSP is enabled, only "just works" mode could be used and thus PAIRABLE and DISCOVERABLE should only be enabled
//   for a short duration

// ToDo: Move all controller specific tasks to controller
func (bt *BtService) StartNAP() (err error) {

	if !bt.IsServiceAvailable() {
		return bluetooth.ErrBtSvcNotAvailable
	}
	log.Println("Bluetooth: starting NAP...")
	// assure bnep module is loaded
	if err = CheckBnep(); err != nil {
		return err
	}

	// Create a bridge interface
	if errBr := bt.EnableBridge(); errBr != nil {
		log.Println("Bridge exists already")
	}

	// Register custom agent bt-agent with "No Input, No Output" capabilities
	// Note: This results in "just works" mode with no MitM protection (see notes above)
	if err = bt.Agent.Start(toolz.AGENT_CAP_NO_INPUT_NO_OUTPUT); err != nil {
		return err
	}

	// SSP and HS enabled, this disables PIN requests but is needed for NAP to work (see comments above)
	bt.Controller.SetPowered(false)
//	bt.Controller.SetSSP(true) //Couldn't use legacy mode (no Secure Simple Pairing, but PIN based pairing), otherwise HighSpeed couldn't be enabled
//	bt.Controller.SetHighSpeed(true) // Enable high speed mode, yeah (without high speed, NAP connections don't work as intended)
	bt.Controller.SetSSP(false) // Fall back to PIN authentication (legacy mode)
	bt.Controller.SetHighSpeed(false) // No high speed without SSP
	bt.Controller.SetPowered(true)

	// Configure adapter
	fmt.Println("Reconfigure adapter to be discoverable and pairable")
	err = bt.Controller.SetAlias("P4wnP1")
	if err != nil {
		return
	}
	err = bt.Controller.SetDiscoverableTimeout(0)
	if err != nil {
		return
	}
	err = bt.Controller.SetPairableTimeout(0)
	if err != nil {
		return
	}
	err = bt.Controller.SetDiscoverable(true)
	if err != nil {
		return
	}
	err = bt.Controller.SetPairable(true)

	time.Sleep(time.Second) //give some time before registering NAP to SDP

	// Enable PAN networking for bridge
	bt.RegisterNetworkServer(toolz.UUID_NETWORK_SERVER_NAP)

	if mi, err := bt.RootSvc.SubSysNetwork.GetManagedInterface(BT_ETHERNET_BRIDGE_NAME); err == nil {
		mi.ReDeploy()
	}

	return
}
*/

func (bt *BtService) SetPIN(pin string) (err error) {
	if !bt.IsServiceAvailable() {
		return bluetooth.ErrBtSvcNotAvailable
	}
	bt.Agent.SetPIN(pin)
	return
}

func (bt *BtService) GetPIN() (pin string, err error) {
	if !bt.IsServiceAvailable() {
		return pin,bluetooth.ErrBtSvcNotAvailable
	}
	return bt.Agent.GetPIN(), nil
}


func (bt *BtService) RegisterNetworkServer(uuid toolz.NetworkServerUUID) (err error) {
	if !bt.IsServiceAvailable() {
		return bluetooth.ErrBtSvcNotAvailable
	}

	return bt.Controller.RegisterNetworkServer(uuid, BT_ETHERNET_BRIDGE_NAME)
}

func (bt *BtService) UnregisterNetworkServer(uuid toolz.NetworkServerUUID) (err error) {
	if !bt.IsServiceAvailable() {
		return bluetooth.ErrBtSvcNotAvailable
	}
	return bt.Controller.UnregisterNetworkServer(uuid)
}


func (bt *BtService) ConnectNetwork(deviceMac string, uuid toolz.NetworkServerUUID) (err error) {
	if !bt.IsServiceAvailable() {
		return bluetooth.ErrBtSvcNotAvailable
	}
	return bt.Controller.ConnectNetwork(deviceMac, uuid)
}

func (bt *BtService) DisconnectNetwork(deviceMac string) (err error) {
	if !bt.IsServiceAvailable() {
		return bluetooth.ErrBtSvcNotAvailable
	}
	return bt.Controller.DisconnectNetwork(deviceMac)
}

func (bt *BtService) IsServerNAPEnabled() (res bool, err error) {
	if !bt.IsServiceAvailable() {
		return false,bluetooth.ErrBtSvcNotAvailable
	}
	return bt.Controller.IsServerNAPEnabled()
}

func (bt *BtService) IsServerPANUEnabled() (res bool, err error) {
	if !bt.IsServiceAvailable() {
		return false,bluetooth.ErrBtSvcNotAvailable
	}
	return bt.Controller.IsServerPANUEnabled()
}

func (bt *BtService) IsServerGNEnabled() (res bool, err error) {
	if !bt.IsServiceAvailable() {
		return false,bluetooth.ErrBtSvcNotAvailable
	}
	return bt.Controller.IsServerGNEnabled()
}

func (bt *BtService) CheckUUIDEnabled(uuids []string) (enabled []bool, err error) {
	if !bt.IsServiceAvailable() {
		return []bool{},bluetooth.ErrBtSvcNotAvailable
	}

	return bt.Controller.CheckUUIDList(uuids)
}

/*
func (bt *BtService) StopNAP() (err error) {
	if !bt.IsServiceAvailable() {
		return bluetooth.ErrBtSvcNotAvailable
	}
	log.Println("Bluetooth: stopping NAP...")

	//Stop bt-agent
	bt.Agent.Stop()

	// Delete bridge interface
	bt.DisableBridge()

	// Unregister pan service
	nw, err := toolz.NetworkServer(bt.Controller.DBusPath)
	//if err != nil { return }
	defer nw.Close()
	err = nw.Unregister("pan")
	//if err != nil { return }

	err = bt.Controller.SetDiscoverable(false)
	//if err != nil { return }
	err = bt.Controller.SetPairable(false)
	//if err != nil { return }

	return
}
*/

// ToDo: Lock bridge creation
func (bt *BtService) EnableBridge() (err error) {
	log.Println("Creating bluetooth bridge interface", bt.BrName)
	exists := CheckInterfaceExistence(bt.BrName)
	if exists {
		log.Printf("... interface %s exists alread\n", bt.BrName)
	} else {
		log.Printf("... interface %s doesn't exist call bridge create\n", bt.BrName)
		//Create the bridge
		err = CreateBridge(bt.BrName)
		if err != nil {
			log.Printf("...error in CreateBridge %v\n", err)
			return err
		}
	}

	log.Printf("... set interface MAC %v\n", BT_ETHERNET_BRIDGE_MAC)
	err = setInterfaceMac(bt.BrName, BT_ETHERNET_BRIDGE_MAC)
	if err != nil {
		log.Printf("...error in setInterfaceMac %v\n", err)
		return err
	}

	log.Println("... set forward delay to 0")
	err = SetBridgeForwardDelay(bt.BrName, 0)
	if err != nil {
		log.Printf("...error in SetBridgeForwardDelay %v\n", err)
		return err
	}

	log.Println("... set spanning tree to off")
	err = SetBridgeSTP(bt.BrName, false)
	if err != nil {
		log.Printf("...error in BridgeSetSTP %v\n", err)
		return err
	}

	//enable the bridge
	log.Println("... bring bridge up")
	err = NetworkLinkUp(bt.BrName)
	if err != nil {
		log.Printf("...error in NetworkLinkUp %v\n", err)
		return err
	}

	// Reconfigure network
	log.Println("... reconfigure ethernet settings for interface", bt.BrName)
	if mi, err := bt.RootSvc.SubSysNetwork.GetManagedInterface(bt.BrName); err == nil {
		mi.ReDeploy()
		// disable IPv6 for bridge interface
		//ioutil.WriteFile("/proc/sys/net/ipv6/conf/" + bt.BrName + "/disable_ipv6", []byte("1"), os.ModePerm)
		// disable IPv6 for all interfaces
		ioutil.WriteFile("/proc/sys/net/ipv6/conf/all/disable_ipv6", []byte("1"), os.ModePerm)
	}


	return
}

func (bt *BtService) DisableBridge() {
	log.Println("Deleting bluetooth bridge interface", bt.BrName)
	//we ignore error results and assume bridge is disable after this call (error could be created if bridge if wasn't existent, too)
	exists := CheckInterfaceExistence(bt.BrName)
	if exists {
		DeleteBridge(bt.BrName)
	}
}

// assures bnep kernel module is loaded
func CheckBnep() error {
	log.Printf("Checking for 'bnep' module...")
	out, err := exec.Command("lsmod").Output()
	if err != nil {
		log.Fatal(err)
	}

	if strings.Contains(string(out), "bnep") {
		log.Printf("... bnep loaded")
		return nil
	}

	//if here, libcomposite isn't loaded ... try to load
	log.Printf("Kernel module 'bnep' not loaded, trying to load ...")
	err = exec.Command("modprobe", "bnep").Run()
	if err == nil {
		log.Printf("... bnep loaded")
	}

	return err
}

/*
ToDo: The binaries used (bluez-tools) should be replaced by custom functions interfacing with bluez D-Bus API, later on.
Example: https://github.com/muka/go-bluetooth
 */

/*
func (bt BtService) CheckExternalBinaries() error {
	bins := []string{"modprobe", "lsmod", "bt-adapter", "bt-agent", "bt-device", "bt-network", "bluetoothd"}
	for _, bin := range bins {
		if !binaryAvailable(bin) {
			return errors.New(bin + " seems to be missing, please install it")
		}

	}
	return nil
}
*/

// ToDo: Get rid of this as soon as an API function is found
// btmgt tool is able to determine Bluez version, mgmt-api is only able to determine Management version (which should be 1.14)
func GetBluezVersion() (major int, minor int, err error) {
	eGeneral := errors.New("Couldn't retrieve bluez version")
	proc := exec.Command("/usr/sbin/bluetoothd", "-v")
	res, err := proc.CombinedOutput()
	if err != nil {
		err = errors.New(fmt.Sprintf("Error fetching Bluez version: '%s'\nbluetoothd output: %s", err, res))
		return
	}

	matches := regexp.MustCompile("(?m)([0-9]+).([0-9]+)").FindStringSubmatch(string(res))
	if len(matches) != 3 {
		err = eGeneral
		return
	}

	major, err = strconv.Atoi(matches[1])
	if err != nil {
		err = eGeneral
		return
	}
	minor, err = strconv.Atoi(matches[2])
	if err != nil {
		err = eGeneral
		return
	}

	return
}

func CheckBluezVersion() (err error) {
	eGeneral := errors.New("Newer Bluez version needed")
	major, minor, err := GetBluezVersion()
	if err != nil {
		return err
	}
	log.Printf("Bluez %d.%d found (minimum needed %d.%d)\n", major, minor, BT_MINIMUM_BLUEZ_VERSION_MAJOR, BT_MINIMUM_BLUEZ_VERSION_MINOR)
	if major > BT_MINIMUM_BLUEZ_VERSION_MAJOR {
		return nil
	}
	if major == BT_MINIMUM_BLUEZ_VERSION_MAJOR {
		if minor >= BT_MINIMUM_BLUEZ_VERSION_MINOR {
			return nil
		} else {
			return eGeneral
		}
	}
	return eGeneral
}

func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func BoolToIntStr(b bool) string {
	return strconv.Itoa(BoolToInt(b))
}

