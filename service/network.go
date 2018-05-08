package service

import (
	"github.com/docker/libcontainer/netlink"
	"net"
	"log"
)

func CreateBridge(name string) (err error) {
	return netlink.CreateBridge(name, false)
}

func setInterfaceMac(name string, mac string) error {
	return netlink.SetMacAddress(name, mac)
}

func DeleteBridge(name string) error {
	return netlink.DeleteBridge(name)
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