package service

import (
	//"github.com/docker/libcontainer/netlink"
	"github.com/mame82/P4wnP1_go/netlink"
	"net"
	"log"
	"io/ioutil"
	"os"
	"fmt"

	pb "github.com/mame82/P4wnP1_go/proto"
	"errors"
)



func ReInitNetworkInterface(ifName string) (err error) {
	if settings, existing := ServiceState.StoredNetworkSetting[ifName]; existing {
		log.Printf("Redeploying stored Network settings for interface '%s' ...\n", ifName)
		return ConfigureInterface(settings)
	} else {
		return errors.New(fmt.Sprintf("No stored interface settings found for '%s'\n", ifName))
	}
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



func CheckInterfaceExistence(name string) (res bool, err error) {
	_, err = net.InterfaceByName(name)
	if err != nil {
		return false, err
	}
	return true, err
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

func ConfigureInterface(settings *pb.EthernetInterfaceSettings) (err error) {
	//Get Interface
	iface, err := net.InterfaceByName(settings.Name)
	if err != nil {	return err }

	//stop DHCP server / client if still running
	running, _, err := IsDHCPServerRunning(settings.Name)
	if (err == nil) && running {StopDHCPServer(settings.Name)}
	running, _, err = IsDHCPClientRunning(settings.Name)
	if (err == nil) && running {StopDHCPClient(settings.Name)}

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
			StopDHCPServer(ifName)

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
			err = StartDHCPServer(ifName, confName)
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

			StartDHCPClient(settings.Name)
		} else {
			log.Printf("Setting Interface %s to DOWN\n", iface.Name)
			err = netlink.NetworkLinkDown(iface)
			if err != nil { return err }
		}

	}

	//Store latest settings
	settings.SettingsInUse = true
	ServiceState.StoredNetworkSetting[settings.Name] = settings

	return nil
}