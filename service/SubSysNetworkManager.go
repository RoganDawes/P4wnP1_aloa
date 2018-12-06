package service

import (
	"errors"
	"fmt"
	//"github.com/docker/libcontainer/netlink"
	"github.com/mame82/P4wnP1_aloa/netlink"
	pb "github.com/mame82/P4wnP1_aloa/proto"
	"github.com/mame82/P4wnP1_aloa/service/util"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"sync"
)

var (
	ErrUnmanagedInterface = errors.New("Not a managed network interface")
)

func NewNetworkManager(rootService *Service) (nm *NetworkManager, err error){
	nm = &NetworkManager{
		rootSvc: rootService,
		ManagedInterfaces: make(map[string]*NetworkInterfaceManager),
	}

	// Add managed interfaces

	// USB
	err = nm.AddManagedInterface(GetDefaultNetworkSettingsUSB())
	if err != nil { return }
	// WiFi
	err = nm.AddManagedInterface(GetDefaultNetworkSettingsWiFi())
	if err != nil { return }
	// Bluetooth
	err = nm.AddManagedInterface(GetDefaultNetworkSettingsBluetooth())
	if err != nil { return }

	//ToDo: Deploy all settings once, to assure consistency of server state and real adapter configuration

	return nm, nil
}


type NetworkManager struct {
	ManagedInterfaces map[string]*NetworkInterfaceManager
	rootSvc *Service
}

func (nm *NetworkManager) AddManagedInterface(startupConfig *pb.EthernetInterfaceSettings) (err error) {
	nim,err := NewNetworkInterfaceManager(nm, startupConfig.Name, startupConfig)
	if err != nil { return err }
	nm.ManagedInterfaces[startupConfig.Name] = nim
	return
}

func (nm *NetworkManager) GetManagedInterfaceNames() (ifnames []string) {
	ifnames = make([]string, len(nm.ManagedInterfaces))
	i:=0
	for name,_ := range nm.ManagedInterfaces {
		ifnames[i] = name
		i += 1
	}
	return
}

func (nm *NetworkManager) GetManagedInterface(name string) (nim *NetworkInterfaceManager, err error) {
	if nim, exists := nm.ManagedInterfaces[name]; exists {
		return nim, nil
	} else {
		return nil, ErrUnmanagedInterface
	}
}



type NetworkInterfaceState struct {
	InterfacePresent bool
	CurrentSettings *pb.EthernetInterfaceSettings
}

// ToDo: interface watcher (up/down --> auto redeploy)
type NetworkInterfaceManager struct {
	nm *NetworkManager
	InterfaceName string
	state *NetworkInterfaceState

	CmdDnsmasq        *exec.Cmd
	mutexDnsmasq      *sync.Mutex
	LoggerDnsmasq     *util.TeeLogger
	leaseMonitor *dnsmasqLeaseMonitor
}

func (nim *NetworkInterfaceManager) GetState() (res *NetworkInterfaceState) {
	return nim.state
}

func (nim *NetworkInterfaceManager) OnHandedOutDhcpLease(lease *DhcpLease) {
	fmt.Printf("Lease monitor %s LEASE: %v\n", nim.InterfaceName, lease)
	// should never happen (dnsmasq output parsing error otherwise)
	if nim.InterfaceName != lease.Iface {
		fmt.Println("Interface of handed out DHCP lease doesn't match managed interface, ignoring ...")
		return
	}

	//generate trigger event
	nim.nm.rootSvc.SubSysEvent.Emit(ConstructEventTriggerDHCPLease(lease.Iface, lease.Mac.String(), lease.Ip.String(), lease.Host))
}

func (nim *NetworkInterfaceManager) OnReceivedDhcpRelease(release *DhcpLease) {
	fmt.Printf("Lease monitor %s RELEASE: %v\n", nim.InterfaceName, release)
	// should never happen (dnsmasq output parsing error otherwise)
	if nim.InterfaceName != release.Iface {
		fmt.Println("Interface for received DHCP release doesn't match managed interface, ignoring ...")
		return
	}
}

func (nim *NetworkInterfaceManager) ReDeploy() (err error) {
	/*
	if settings, existing := ServiceState.StoredNetworkSettings[ifName]; existing {
		log.Printf("Redeploying stored Network settings for interface '%s' ...\n", ifName)
		return ConfigureInterface(settings)
	} else {
		return errors.New(fmt.Sprintf("No stored interface settings found for '%s'\n", ifName))
	}
	*/
	return nim.DeploySettings(nim.state.CurrentSettings)
}

func (nim *NetworkInterfaceManager) DeploySettings(settings *pb.EthernetInterfaceSettings) (err error) {
	//Get Interface
	iface, err := net.InterfaceByName(settings.Name)
	if err != nil {
		nim.state.InterfacePresent = false
		//return err
		return nil //Not having the interface present isn't an error
	} else {
		nim.state.InterfacePresent = true
	}

	//stop DHCP server / client if still running
	nim.StopDHCPServer()
	nim.StopDHCPClient()

	switch settings.Mode {
	case pb.EthernetInterfaceSettings_MANUAL:
		//Generate net
		ipNet, err := IpNetFromIPv4AndNetmask(settings.IpAddress4, settings.Netmask4)
		if err != nil { return err }

		//Flush old IPs
		netlink.NetworkLinkFlush(iface)
		//set IP
		log.Printf("Setting Interface %s to IP %s\n", iface.Name, settings.IpAddress4)
		netlink.NetworkLinkAddIp(iface, net.ParseIP(settings.IpAddress4), ipNet)

		if settings.Enabled {
			log.Printf("Setting Interface %s to UP\n", iface.Name)
			err = netlink.NetworkLinkUp(iface)
			if err != nil { return err }
			log.Printf("Setting Interface %s to MULTICAST to ON\n", iface.Name)
			err = netlink.NetworkSetMulticast(iface, true)
			if err != nil { return err }

		} else {
			log.Printf("Setting Interface %s to DOWN\n", iface.Name)
			err = netlink.NetworkLinkDown(iface)
			if err != nil { return err }
		}

	case pb.EthernetInterfaceSettings_DHCP_SERVER:
		//Generate net
		ipNet, err := IpNetFromIPv4AndNetmask(settings.IpAddress4, settings.Netmask4)
		if err != nil { return err }

		//Flush old IPs
		netlink.NetworkLinkFlush(iface)
		//set IP
		log.Printf("Setting Interface %s to IP %s\n", iface.Name, settings.IpAddress4)
		netlink.NetworkLinkAddIp(iface, net.ParseIP(settings.IpAddress4), ipNet)

		if settings.Enabled {
			log.Printf("Setting Interface %s to UP\n", iface.Name)
			err = netlink.NetworkLinkUp(iface)
			if err != nil { return err }
			log.Printf("Setting Interface %s to MULTICAST to ON\n", iface.Name)
			err = netlink.NetworkSetMulticast(iface, true)
			if err != nil { return err }


			//check DhcpServerSettings
			if settings.DhcpServerSettings == nil {
				err = errors.New(fmt.Sprintf("Ethernet configuration for interface %s is set to DHCP Server mode, but doesn't provide DhcpServerSettings", settings.Name))
				log.Println(err)
				return err
			}
			ifName := settings.Name
			confName := NameConfigFileDHCPSrv(ifName)
			err = DHCPCreateConfigFile(settings.DhcpServerSettings, confName)
			if err != nil {return err}
			//stop already running DHCPServers for the interface
			nim.StopDHCPServer()

			//special case: if the interface name is USB_ETHERNET_BRIDGE_NAME, we delete the old lease file
			// the flushing of still running leases is needed, as after USB reinit, RNDIS hosts aren't guaranteed to
			// receive the sam MAC, which would effectivly block reusing of a lease for the same IP (a problem, as in
			// typical DHCP server configurations for USB Ethernet, the same remote IP should be offered every time)
			if settings.Name == USB_ETHERNET_BRIDGE_NAME {
				log.Printf("Reconfiguration of USB Ethernert interface as DHCP server, trying to delete old lease file ...\n")
				errD := os.Remove(settings.DhcpServerSettings.LeaseFile)
				if errD == nil {
					log.Println(" ... old lease file deleted successfull")
				} else {
					log.Printf(" ... old lease couldn't be deleted (Fetching a new DHCP lease could take a while on USB ethernet)!\n\tReason: %v\n", errD)
				}
			}

			//start the DHCP server
			err = nim.StartDHCPServer(confName)
			if err != nil {return err}
		} else {
			log.Printf("Setting Interface %s to DOWN\n", iface.Name)
			err = netlink.NetworkLinkDown(iface)
		}
		if err != nil { return err }
	case pb.EthernetInterfaceSettings_DHCP_CLIENT:
		netlink.NetworkLinkFlush(iface)
		if settings.Enabled {
			log.Printf("Setting Interface %s to UP\n", iface.Name)
			err = netlink.NetworkLinkUp(iface)
			if err != nil { return err }
			log.Printf("Setting Interface %s to MULTICAST to ON\n", iface.Name)
			err = netlink.NetworkSetMulticast(iface, true)
			if err != nil { return err }

			nim.StartDHCPClient()
		} else {
			log.Printf("Setting Interface %s to DOWN\n", iface.Name)
			err = netlink.NetworkLinkDown(iface)
			if err != nil { return err }
		}

	}

	//Store latest settings
	settings.SettingsInUse = true

	//ServiceState.StoredNetworkSettings[settings.Name] = settings
	nim.state.CurrentSettings = settings

	return nil
}

func NewNetworkInterfaceManager(nm *NetworkManager, ifaceName string, startupSettings *pb.EthernetInterfaceSettings) (nim *NetworkInterfaceManager, err error) {
	nim = &NetworkInterfaceManager{
		nm: nm,
		InterfaceName: ifaceName,
		state: &NetworkInterfaceState{},
		mutexDnsmasq: &sync.Mutex{},
		LoggerDnsmasq: util.NewTeeLogger(false),
	}
	nim.leaseMonitor = NewDnsmasqLeaseMonitor(nim)

	//nim.LoggerDnsmasq.SetPrefix("dnsmasq-" + ifaceName + ": ")
	nim.LoggerDnsmasq.AddOutput(nim.leaseMonitor)


	nim.state.CurrentSettings = startupSettings
	nim.ReDeploy()

	return
}


/* HELPER */
func nameLeaseFileDHCPSrv(nameIface string) (lf string) {
	return "/tmp/dnsmasq_" + nameIface + ".leases"
}

func ParseIPv4Mask(maskstr string) (net.IPMask, error) {
	mask := net.ParseIP(maskstr)
	if mask == nil { return nil, errors.New("Couldn't parse netmask") }

	net.ParseCIDR(maskstr)
	return net.IPv4Mask(mask[12], mask[13], mask[14], mask[15]), nil
}

func IpNetFromIPv4AndNetmask(ipv4 string, netmask string) (*net.IPNet, error) {
	mask, err := ParseIPv4Mask(netmask)
	if err != nil { return nil, err }

	ip := net.ParseIP(ipv4)
	if mask == nil { return nil, errors.New("Couldn't parse IP") }

	netw := ip.Mask(mask)

	return &net.IPNet{IP: netw, Mask: mask}, nil
}



func CreateBridge(name string) (err error) {
	return netlink.CreateBridge(name, false)
}

func setInterfaceMac(name string, mac string) error {
	return netlink.SetMacAddress(name, mac)
}

func DeleteBridge(name string) error {
	return netlink.DeleteBridge(name)
}

//Uses sysfs (not IOCTL)
func SetBridgeSTP(name string, stp_on bool) (err error) {
	value := "0"
	if (stp_on) { value = "1" }
	return ioutil.WriteFile(fmt.Sprintf("/sys/class/net/%s/bridge/stp_state", name), []byte(value), os.ModePerm)
}

func SetBridgeForwardDelay(name string, fd uint) (err error) {
	return ioutil.WriteFile(fmt.Sprintf("/sys/class/net/%s/bridge/forward_delay", name), []byte(fmt.Sprintf("%d", fd)), os.ModePerm)
}


// ToDo: remove error part
func CheckInterfaceExistence(name string) (res bool) {
	_, err := net.InterfaceByName(name)
	if err != nil {
		return false
	}
	return true
}

func NetworkLinkUp(name string) (err error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return err
	}

	err = netlink.NetworkLinkUp(iface)
	return
}

func AddInterfaceToBridgeIfExistent(bridgeName string, ifName string) (err error) {
	br, err := net.InterfaceByName(bridgeName)
	if err != nil {
		return err
	}
	iface, err := net.InterfaceByName(ifName)
	if err != nil {
		return err
	}

	err = netlink.AddToBridge(iface, br)
	if err != nil {
		return err
	}
	log.Printf("Interface %s added to bridge %s", ifName, bridgeName)

	//enable interface
	NetworkLinkUp(ifName)
	return nil
}
