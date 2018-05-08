package service

import (
	"github.com/docker/libcontainer/netlink"
	"net"
	"log"
	"io/ioutil"
	"os"
	"fmt"

	pb "../proto"
	"errors"
)

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
	return nil
}

func ConfigureInterface(settings *pb.EthernetInterfaceSettings) (err error) {
	//Get Interface
	iface, err := net.InterfaceByName(settings.Name)
	if err != nil {	return err }

	switch settings.Mode {
	case pb.EthernetInterfaceSettings_MANUAL:
		//Generate net
		ipNet, err := IpNetFromIPv4AndNetmask(settings.IpAddress4, settings.Netmask4)
		if err != nil { return err }

		//set IP
		netlink.NetworkLinkAddIp(iface, net.ParseIP(settings.IpAddress4), ipNet)
	}

	return nil
}