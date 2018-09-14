package service

import (
	"errors"
	"fmt"
	"github.com/mame82/P4wnP1_go/service/util"
	"log"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const (
	bt_dev_name                    string = "hci0"
	BT_MINIMUM_BLUEZ_VERSION_MAJOR        = 5
	BT_MINIMUM_BLUEZ_VERSION_MINOR        = 43
)

const (
	BT_AGENT_MODE_DISPLAY_YES_NO     = BtAgentMode("DisplayYesNo")
	BT_AGENT_MODE_DISPLAY_ONLY       = BtAgentMode("DisplayOnly")
	BT_AGENT_MODE_KEYBOARD_ONLY      = BtAgentMode("KeyboardOnly")
	BT_AGENT_MODE_NO_INPUT_NO_OUTPUT = BtAgentMode("NoInputNoOutput")
)

type BtAgentMode string

type BtService struct {
	DevName          string
	BrName           string
	PathBtConf       string
	bridgeIfDeployed bool

	Agent   *BtAgent
	Adapter *BtAdapter
}

type BtAdapter struct {
	/*
	  --set <property> <value>
  Where `property` is one of:
     Name
     Discoverable
     DiscoverableTimeout
     Pairable
     PairableTimeout
     Powered


	root@raspberrypi:~# bt-adapter -i
[hci0]
  Name: raspberrypi
  Address: B8:27:EB:8E:44:43
  Alias: raspberrypi [rw]
  Class: 0x0
  Discoverable: 0 [rw]
  DiscoverableTimeout: 180 [rw]
  Discovering: 0
  Pairable: 1 [rw]
  PairableTimeout: 0 [rw]
  Powered: 1 [rw]
  UUIDs: [00001801-0000-1000-8000-00805f9b34fb, AVRemoteControl, PnPInformation, 00001800-0000-1000-8000-00805f9b34fb, AVRemoteControlTarget]

	 */

	*sync.Mutex
	Address             net.HardwareAddr
	DeviceName          string
	Name                string // Not changeable
	Alias               string
	Discoverable        bool
	DiscoverableTimeout uint64
	Pairable            bool
	PairableTimeout     uint64
	Powered             bool
}

var (
	reAdapterAddress             = regexp.MustCompile("(?m)Address: ([0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2})")
	reAdapterName                = regexp.MustCompile("(?m)Name: (.*)\n")
	reAdapterAlias               = regexp.MustCompile("(?m)Alias: (.*) \\[")
	reAdapterDiscoverable        = regexp.MustCompile("(?m)Discoverable: ([01])")
	reAdapterDiscoverableTimeout = regexp.MustCompile("(?m)DiscoverableTimeout: ([0-9]+)")
	reAdapterPairable            = regexp.MustCompile("(?m)Pairable: ([01])")
	reAdapterPairableTimeout     = regexp.MustCompile("(?m)PairableTimeout: ([0-9]+)")
	reAdapterPowered             = regexp.MustCompile("(?m)Powered: ([01])")

	eAdapterParseOutput = errors.New("Error parsing output of `bt-adapter -i`")
	eAdapterSetAdapter  = errors.New("Error setting adapter options with `bt-adapter -s`")
)

func (bAd *BtAdapter) DeploySate() (err error) {
	proc := exec.Command("/usr/bin/bt-adapter", "-s", "Alias", bAd.Alias, "-a", bAd.DeviceName)
	if proc.Run() != nil {
		return eAdapterSetAdapter
	}
	proc = exec.Command("/usr/bin/bt-adapter", "-s", "Discoverable", BoolToIntStr(bAd.Discoverable), "-a", bAd.DeviceName)
	if proc.Run() != nil {
		return eAdapterSetAdapter
	}
	proc = exec.Command("/usr/bin/bt-adapter", "-s", "DiscoverableTimeout", strconv.Itoa(int(bAd.DiscoverableTimeout)), "-a", bAd.DeviceName)
	if proc.Run() != nil {
		return eAdapterSetAdapter
	}
	proc = exec.Command("/usr/bin/bt-adapter", "-s", "Pairable", BoolToIntStr(bAd.Pairable), "-a", bAd.DeviceName)
	if proc.Run() != nil {
		return eAdapterSetAdapter
	}
	proc = exec.Command("/usr/bin/bt-adapter", "-s", "PairableTimeout", strconv.Itoa(int(bAd.PairableTimeout)), "-a", bAd.DeviceName)
	if proc.Run() != nil {
		return eAdapterSetAdapter
	}
	proc = exec.Command("/usr/bin/bt-adapter", "-s", "Powered", BoolToIntStr(bAd.Powered), "-a", bAd.DeviceName)
	if proc.Run() != nil {
		return eAdapterSetAdapter
	}
	bAd.updateMembers()
	return
}

// Updates members via `bt-adapter -i`
func (bAd *BtAdapter) updateMembers() (err error) {
	proc := exec.Command("/usr/bin/bt-adapter", "-i", "-a", bAd.DeviceName)
	res, err := proc.CombinedOutput()
	if err != nil {
		return errors.New(fmt.Sprintf("Error running `bt-adapter -i -a %s`: %s\niw output: %s", bAd.DeviceName, err, res))
	}
	output := string(res)

	strAdapterAddress := ""
	strAdapterName := ""
	strAdapterAlias := ""
	strAdapterDiscoverable := ""
	strAdapterDiscoverableTimeout := ""
	strAdapterPairable := ""
	strAdapterPairableTimeout := ""
	strAdapterPowered := ""

	if matches := reAdapterAddress.FindStringSubmatch(output); len(matches) > 1 {
		strAdapterAddress = matches[1]
	} else {
		return eAdapterParseOutput
	}

	if matches := reAdapterName.FindStringSubmatch(output); len(matches) > 1 {
		strAdapterName = matches[1]
	} else {
		return eAdapterParseOutput
	}

	if matches := reAdapterAlias.FindStringSubmatch(output); len(matches) > 1 {
		strAdapterAlias = matches[1]
	} else {
		return eAdapterParseOutput
	}

	if matches := reAdapterDiscoverable.FindStringSubmatch(output); len(matches) > 1 {
		strAdapterDiscoverable = matches[1]
	} else {
		return eAdapterParseOutput
	}

	if matches := reAdapterDiscoverableTimeout.FindStringSubmatch(output); len(matches) > 1 {
		strAdapterDiscoverableTimeout = matches[1]
	} else {
		return eAdapterParseOutput
	}

	if matches := reAdapterPairable.FindStringSubmatch(output); len(matches) > 1 {
		strAdapterPairable = matches[1]
	} else {
		return eAdapterParseOutput
	}

	if matches := reAdapterPairableTimeout.FindStringSubmatch(output); len(matches) > 1 {
		strAdapterPairableTimeout = matches[1]
	} else {
		return eAdapterParseOutput
	}

	if matches := reAdapterPowered.FindStringSubmatch(output); len(matches) > 1 {
		strAdapterPowered = matches[1]
	} else {
		return eAdapterParseOutput
	}

	/*
	fmt.Println("strAdapterAddress", strAdapterAddress)
	fmt.Println("strAdapterName", strAdapterName)
	fmt.Println("strAdapterAlias", strAdapterAlias)
	fmt.Println("strAdapterDiscoverable", strAdapterDiscoverable)
	fmt.Println("strAdapterDiscoverableTimeout", strAdapterDiscoverableTimeout)
	fmt.Println("strAdapterPairable", strAdapterPairable)
	fmt.Println("strAdapterPairableTimeout", strAdapterPairableTimeout)
	fmt.Println("strAdapterPowered", strAdapterPowered)
	*/

	if bAd.Address, err = net.ParseMAC(strAdapterAddress); err != nil {
		return err
	}
	if bAd.Discoverable, err = strconv.ParseBool(strAdapterDiscoverable); err != nil {
		return err
	}
	if bAd.DiscoverableTimeout, err = strconv.ParseUint(strAdapterDiscoverableTimeout, 10, 64); err != nil {
		return err
	}
	if bAd.Pairable, err = strconv.ParseBool(strAdapterPairable); err != nil {
		return err
	}
	if bAd.PairableTimeout, err = strconv.ParseUint(strAdapterPairableTimeout, 10, 64); err != nil {
		return err
	}
	if bAd.Powered, err = strconv.ParseBool(strAdapterPowered); err != nil {
		return err
	}
	bAd.Name = strAdapterName
	bAd.Alias = strAdapterAlias

	/*
	log.Printf("adapter: %+v\n", bAd)
	*/

	return
}

func NewBtAdapter() (bAd *BtAdapter, err error) {
	bAd = &BtAdapter{
		Mutex: &sync.Mutex{},
	}
	err = bAd.updateMembers()
	if err != nil {
		return nil, err
	}
	return
}

type BtAgent struct {
	*exec.Cmd
	*sync.Mutex //mutex for wpa-supplicant proc
	mode        BtAgentMode
	pinFilePath string
	logger      *util.TeeLogger
}

func (ba *BtAgent) Start(mode BtAgentMode) (err error) {
	log.Printf("Starting bt-agent with mode '%s'...\n", ba.mode)

	ba.Lock()
	defer ba.Unlock()

	ba.mode = mode

	//stop if already running
	if ba.Cmd != nil {
		// avoid deadlock
		ba.Unlock()
		ba.Stop()
		ba.Lock()
	}

	ba.Cmd = exec.Command("/usr/bin/bt-agent", "-c", string(ba.mode))
	ba.Cmd.Stdout = ba.logger.LogWriter
	ba.Cmd.Stderr = ba.logger.LogWriter
	err = ba.Cmd.Start()
	if err != nil {
		ba.Cmd.Wait()
		return errors.New(fmt.Sprintf("Error starting bt-agent '%v'", err))
	}
	log.Println("... bt-agent started")
	return nil

}

func (ba *BtAgent) Stop() (err error) {
	log.Println("Stopping bt-agent...")
	ba.Lock()
	defer ba.Unlock()
	if ba.Cmd == nil {
		log.Println("bt-agent already stopped")
	}
	if ba.Process == nil {
		return errors.New("Couldn't access bt-agent process")
	}
	ba.Process.Kill()
	ba.Process.Wait()
	if ba.ProcessState == nil {
		return errors.New("Couldn't access bt-agent process state")
	}
	if !ba.ProcessState.Exited() {
		return errors.New("bt-agent didn't terminate after SIGKILL")
	}
	ba.Cmd = nil
	return nil
}

func NewBtAgent() (res *BtAgent) {
	res = &BtAgent{
		Mutex:       &sync.Mutex{},
		mode:        BT_AGENT_MODE_NO_INPUT_NO_OUTPUT,
		pinFilePath: "",
		logger:      util.NewTeeLogger(true),
	}
	res.logger.SetPrefix("bt-agent: ")

	return res
}

func NewBtService() (res *BtService) {
	res = &BtService{
		Agent:      NewBtAgent(),
		DevName:    bt_dev_name,
		BrName:     BT_ETHERNET_BRIDGE_NAME,
		PathBtConf: "",
	}
	if err := res.CheckExternalBinaries(); err != nil {
		panic(err)
	}
	if err := CheckBluezVersion(); err != nil {
		panic(err)
	}
	// ToDo Check if bluetooth service is loaded
	if btAdp, errAd := NewBtAdapter(); errAd != nil {
		panic(errAd)
	} else {
		res.Adapter = btAdp
	}

	return
}

func (bt *BtService) StartNAP() (err error) {
	log.Println("Bluetooth: starting NAP...")
	// assure bnep module is loaded
	if err = CheckBnep(); err != nil {
		return err
	}

	// Create a bridge interface
	if errBr := bt.EnableBridge(); errBr != nil {
		log.Println("Bridge exists already")
	}

	// start bt-agent with "No Input, No Output" capabilities
	if err = bt.Agent.Start(BT_AGENT_MODE_NO_INPUT_NO_OUTPUT); err != nil {
		return err
	}

	fmt.Println("Reconfigure adapter to be discoverable and pairable")
	bt.Adapter.Alias = "P4wnP1"
	bt.Adapter.Discoverable = true
	bt.Adapter.DiscoverableTimeout = 0
	bt.Adapter.Pairable = true
	bt.Adapter.PairableTimeout = 0
	if err = bt.Adapter.DeploySate(); err != nil {
		return err
	} else {
		log.Printf("... reconfiguration succeeded: %+v\n", bt.Adapter)
	}

	return
}

func (bt *BtService) StopNAP() (err error) {
	log.Println("Bluetooth: stopping NAP...")

	//Stop bt-agent
	bt.Agent.Stop()

	// Delete bridge interface
	bt.DisableBridge()

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
func (bt BtService) CheckExternalBinaries() error {
	bins := []string{"modprobe", "lsmod", "bt-adapter", "bt-agent", "bt-device", "bt-network", "bluetoothd"}
	for _, bin := range bins {
		if !binaryAvailable(bin) {
			return errors.New(bin + " seems to be missing, please install it")
		}

	}
	return nil
}

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
