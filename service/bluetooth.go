package service

import (
	"errors"
	"fmt"
	"github.com/mame82/P4wnP1_go/service/bluetooth"
	"github.com/mame82/mblue-toolz/toolz"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	BT_MINIMUM_BLUEZ_VERSION_MAJOR        = 5
	BT_MINIMUM_BLUEZ_VERSION_MINOR        = 43
)

type BtService struct {
	ServiceAvailable bool
	Controller       *bluetooth.Controller
	BrName           string
	bridgeIfDeployed bool

	Agent   *bluetooth.DefaultAgent
}


func NewBtService() (res *BtService) {
	res = &BtService{
		Agent:      bluetooth.NewDefaultAgent("4321"),
		BrName:     BT_ETHERNET_BRIDGE_NAME,
	}

	// ToDo Check if bluetooth service is loaded
	if c,err := bluetooth.FindFirstAvailableController(); err == nil {
		res.ServiceAvailable = true
		res.Controller = c
	} else {
		res.ServiceAvailable = false
		return
	}
	if err := CheckBluezVersion(); err != nil {
		fmt.Println(err)
		res.ServiceAvailable = false
		return
	}

	return
}

// ToDo: Move all controller specific tasks to controller
func (bt *BtService) StartNAP() (err error) {
	if !bt.ServiceAvailable { return bluetooth.ErrBtSvcNotAvailable
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
	if err = bt.Agent.Start(toolz.AGENT_CAP_NO_INPUT_NO_OUTPUT); err != nil {
		return err
	}

	// Disable simple secure pairing to make PIN requests work
	bt.Controller.SetPowered(false)
	bt.Controller.SetSSP(false)
	bt.Controller.SetPowered(true)

	// Configure adapter
	fmt.Println("Reconfigure adapter to be discoverable and pairable")
	err = bt.Controller.SetAlias("P4wnP1")
	if err != nil { return }
	err = bt.Controller.SetDiscoverableTimeout(0)
	if err != nil { return }
	err = bt.Controller.SetPairableTimeout(0)
	if err != nil { return }
	err = bt.Controller.SetDiscoverable(true)
	if err != nil { return }
	err = bt.Controller.SetPairable(true)

	// Enable PAN networking for bridge
	nw,err := toolz.NetworkServer(bt.Controller.DBusPath)
	if err != nil { return }
	//defer nw.Close()
	err = nw.Register(toolz.UUID_NETWORK_SERVER_NAP, BT_ETHERNET_BRIDGE_NAME)
	if err != nil { return }

	return
}

func (bt *BtService) StopNAP() (err error) {
	if !bt.ServiceAvailable { return bluetooth.ErrBtSvcNotAvailable	}
	log.Println("Bluetooth: stopping NAP...")

	//Stop bt-agent
	bt.Agent.Stop()

	// Delete bridge interface
	bt.DisableBridge()

	// Unregister pan service
	nw,err := toolz.NetworkServer(bt.Controller.DBusPath)
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

func (bt *BtService) EnableBridge() (err error) {
	log.Println("Creating bluetooth bridge interface", BT_ETHERNET_BRIDGE_NAME)
	//Create the bridge
	err = CreateBridge(bt.BrName)
	if err != nil {
		return err
	}

	err = setInterfaceMac(BT_ETHERNET_BRIDGE_NAME, BT_ETHERNET_BRIDGE_MAC)
	if err != nil {
		return err
	}
	//	SetBridgeSTP(BT_ETHERNET_BRIDGE_NAME, true) //enable spanning tree
	SetBridgeForwardDelay(BT_ETHERNET_BRIDGE_NAME, 0)
	if err != nil {
		return err
	}

	//enable the bridge
	err = NetworkLinkUp(BT_ETHERNET_BRIDGE_NAME)
	if err != nil {
		return err
	}

	bt.bridgeIfDeployed = true
	return
}

func (bt *BtService) DisableBridge() {
	log.Println("Deleting bluetooth bridge interface", BT_ETHERNET_BRIDGE_NAME)
	//we ignore error results and assume bridge is disable after this call (error could be created if bridge if wasn't existent, too)
	DeleteBridge(BT_ETHERNET_BRIDGE_NAME)
	bt.bridgeIfDeployed = false
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
	fmt.Printf("Bluez %d.%d found (minimum needed %d.%d)\n", major, minor, BT_MINIMUM_BLUEZ_VERSION_MAJOR, BT_MINIMUM_BLUEZ_VERSION_MINOR)
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
