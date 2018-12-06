package bluetooth

import (
	"fmt"
	"github.com/godbus/dbus"
	"github.com/mame82/mblue-toolz/bt_uuid"
	"net"
	"strings"
	"sync"
	"time"

	//"github.com/mame82/P4wnP1_aloa/service"
	"github.com/mame82/mblue-toolz/btmgmt"
	"github.com/mame82/mblue-toolz/dbusHelper"
	"github.com/mame82/mblue-toolz/toolz"
	"errors"
	pb "github.com/mame82/P4wnP1_aloa/proto"

)

/*
This code assumes that the first bluetooth controller never gets detached,
as it should be the case for a Pi Zero W. It doesn't account for plug&play Bluetooth controllers.
Attaching an additional controller (f.e. USB) could lead to errors, which aren't handled by this code.
 */

type Controller struct {
	DBusPath dbus.ObjectPath // The path of the controller, used by DBus (f.e. 'hci0')
	Index    uint16          // The index of the controller, when "Bluetooth Management Socket" is used (mgmt-api)
	adapter  *toolz.Adapter1
	mgmt *btmgmt.BtMgmt


	deploySettingsFromControllerInformationLock *sync.Mutex
}

func (c *Controller) SetSSP(val bool) (err error) {
	s,err := c.mgmt.SetSecureSimplePairing(c.Index, val)
	if err != nil || s.SecureSimplePairing != val {
		return ErrChgSetting
	}
	return
}

func (c *Controller) SetHighSpeed(val bool) (err error) {
	s,err := c.mgmt.SetHighSpeed(c.Index, val)
	if err != nil || s.HighSpeed != val {
		return ErrChgSetting
	}
	return
}

func (c *Controller) SetBondable(val bool) (err error) {
	s,err := c.mgmt.SetBondable(c.Index, val)
	if err != nil || s.Bondable != val {
		return ErrChgSetting
	}
	return
}

func (c *Controller) SetLowEnergy(val bool) (err error) {
	s,err := c.mgmt.SetLowEnergy(c.Index, val)
	if err != nil || s.LowEnergy != val {
		return ErrChgSetting
	}
	return
}

func (c *Controller) SetLinkLevelSecurity(val bool) (err error) {
	s,err := c.mgmt.SetLinkSecurity(c.Index, val)
	if err != nil || s.LinkLevelSecurity != val {
		return ErrChgSetting
	}
	return
}

func (c *Controller) SetConnectable(val bool) (err error) {
	s,err := c.mgmt.SetConnectable(c.Index, val)
	if err != nil || s.Connectable != val {
		return ErrChgSetting
	}
	return
}

func (c *Controller) SetFastConnectable(val bool) (err error) {
	s,err := c.mgmt.SetFastConnectable(c.Index, val)
	if err != nil || s.FastConnectable != val {
		return ErrChgSetting
	}
	return
}

func (c *Controller) SetDiscoverableExt(discoverable bool, timeout time.Duration) (err error) {
	timeoutSeconds := uint16(timeout.Seconds())
	discoverableMode := btmgmt.NOT_DISCOVERABLE
	if discoverable {
		discoverableMode = btmgmt.GENERAL_DISCOVERABLE
	}
	if timeoutSeconds > 0 && discoverable {
		discoverableMode = btmgmt.LIMITED_DISCOVERABLE
	}
	s,err := c.mgmt.SetDiscoverable(c.Index, discoverableMode, timeoutSeconds)
	if err != nil || s.Discoverable != discoverable {
		return ErrChgSetting
	}
	return
}

func (c *Controller) ReadControllerInformation() (ctlInfo *btmgmt.ControllerInformation ,err error) {
	mgmt,err := btmgmt.NewBtMgmt()
	if err != nil { return nil,ErrReadSetting }

	ctlInfo,err = mgmt.ReadControllerInformation(c.Index)
	if err != nil { return ctlInfo,err }

	//fmt.Printf("ReadControllerInformation:\n%+v\n", ctlInfo)



	uuidsEnabled,err := c.CheckUUIDList([]string{bt_uuid.NAP_UUID, bt_uuid.PANU_UUID, bt_uuid.GN_UUID})
	if err != nil { return ctlInfo,err }
	ctlInfo.ServiceNetworkServerNap = uuidsEnabled[0]
	ctlInfo.ServiceNetworkServerPanu = uuidsEnabled[1]
	ctlInfo.ServiceNetworkServerGn = uuidsEnabled[2]


	return
}


func (c *Controller) StartDiscovery() error {
	return c.adapter.StartDiscovery()
}

func (c *Controller) StopDiscovery() error {
	return c.adapter.StopDiscovery()
}
/* Properties */
func (c *Controller) GetAddress() (res net.HardwareAddr, err error) {
	return c.adapter.GetAddress()
}

func (c *Controller) GetAddressType() (res string, err error) {
	return c.adapter.GetAddressType()
}

func (c *Controller) GetName() (res string, err error) {
	return c.adapter.GetName()
}

func (c *Controller) SetAlias(val string) (err error) {
	return c.adapter.SetAlias(val)
}

func (c *Controller) GetAlias() (res string, err error) {
	return c.adapter.GetAlias()
}

func (c *Controller) GetClass() (res uint32, err error) {
	return c.adapter.GetClass()
}

func (c *Controller) GetPowered() (res bool, err error) {
	return c.adapter.GetPowered()
}

func (c *Controller) SetPowered(val bool) (err error) {
	return c.adapter.SetPowered(val)
}

func (c *Controller) GetDiscoverable() (res bool, err error) {
	return c.adapter.GetDiscoverable()
}

func (c *Controller) SetDiscoverable(val bool) (err error) {
	return c.adapter.SetDiscoverable(val)
}

func (c *Controller) GetPairable() (res bool, err error) {
	return c.adapter.GetPairable()
}

func (c *Controller) SetPairable(val bool) (err error) {
	return c.adapter.SetPairable(val)
}

func (c *Controller) SetDiscoverableTimeout(val uint32) (err error) {
	return c.adapter.SetDiscoverableTimeout(val)
}

func (c *Controller) GetDiscoverableTimeout() (res uint32, err error) {
	return c.adapter.GetDiscoverableTimeout()
}

func (c *Controller) SetPairableTimeout(val uint32) (err error) {
	return c.adapter.SetPairableTimeout(val)
}

func (c *Controller) GetPairableTimeout() (res uint32, err error) {
	return c.adapter.GetPairableTimeout()
}

func (c *Controller) GetDiscovering() (res bool, err error) {
	return c.adapter.GetDiscovering()
}

func (c *Controller) GetUUIDs() (res []string, err error) {
	return c.adapter.GetUUIDs()
}

func (c *Controller) GetModalias() (res string, err error) {
	return c.adapter.GetModalias()
}

func (c *Controller) UpdateSettingsFromChangedControllerInformation(newCi *btmgmt.ControllerInformation, bridgeNameNAP string, bridgeNamePANU string, bridgeNameGN string) (currentCi *btmgmt.ControllerInformation,err error) {
	c.deploySettingsFromControllerInformationLock.Lock()
	defer c.deploySettingsFromControllerInformationLock.Unlock()


	if err != nil { return nil,err }

	currentCi,err = c.ReadControllerInformation()
	if err != nil { return }

	//Update alias if needed
	if currentCi.Name != newCi.Name {
		c.SetAlias(newCi.Name)
	}

	// Update changeable toggles
	if currentCi.CurrentSettings.SecureSimplePairing != newCi.CurrentSettings.SecureSimplePairing {
		err := c.SetSSP(newCi.CurrentSettings.SecureSimplePairing)
		if err != nil {
			fmt.Println("Error setting bluetooth SSP")
			currentCi,_ = c.ReadControllerInformation()
			return currentCi,err
		}
	}

	if currentCi.CurrentSettings.Connectable != newCi.CurrentSettings.Connectable {
		err := c.SetConnectable(newCi.CurrentSettings.Connectable)
		if err != nil {
			fmt.Println("Error setting bluetooth Connectable")
			currentCi,_ = c.ReadControllerInformation()
			return currentCi,err
		}
	}
	if currentCi.CurrentSettings.FastConnectable != newCi.CurrentSettings.FastConnectable {
		err := c.SetFastConnectable(newCi.CurrentSettings.Connectable)
		if err != nil {
			fmt.Println("Error setting bluetooth FastConnectable")
			currentCi,_ = c.ReadControllerInformation()
			return currentCi,err
		}
	}
	if currentCi.CurrentSettings.HighSpeed != newCi.CurrentSettings.HighSpeed {
		err := c.SetHighSpeed(newCi.CurrentSettings.HighSpeed)
		if err != nil {
			fmt.Println("Error setting bluetooth HighSpeed")
			currentCi,_ = c.ReadControllerInformation()
			return currentCi,err
		}
	}
	if currentCi.CurrentSettings.LowEnergy != newCi.CurrentSettings.LowEnergy {
		err := c.SetLowEnergy(newCi.CurrentSettings.LowEnergy)
		if err != nil {
			fmt.Println("Error setting bluetooth LowEnergy")
			currentCi,_ = c.ReadControllerInformation()
			return currentCi,err
		}
	}
	if currentCi.CurrentSettings.LinkLevelSecurity != newCi.CurrentSettings.LinkLevelSecurity {
		err := c.SetLinkLevelSecurity(newCi.CurrentSettings.LinkLevelSecurity)
		if err != nil {
			fmt.Println("Error setting bluetooth LinkLevelSecurity")
			currentCi,_ = c.ReadControllerInformation()
			return currentCi,err
		}
	}
	if currentCi.CurrentSettings.Powered != newCi.CurrentSettings.Powered {
		err := c.SetPowered(newCi.CurrentSettings.Powered)
		if err != nil {
			fmt.Println("Error setting bluetooth Powered")
			currentCi,_ = c.ReadControllerInformation()
			return currentCi,err
		}
	}
	if currentCi.CurrentSettings.Discoverable != newCi.CurrentSettings.Discoverable {
		err := c.SetDiscoverable(newCi.CurrentSettings.Discoverable)
		if err != nil {
			fmt.Println("Error setting bluetooth Discoverable")
			currentCi,_ = c.ReadControllerInformation()
			return currentCi,err
		}
	}
	if currentCi.CurrentSettings.Bondable != newCi.CurrentSettings.Bondable {
		err := c.SetBondable(newCi.CurrentSettings.Bondable)
		if err != nil {
			fmt.Println("Error setting bluetooth Bondable")
			currentCi,_ = c.ReadControllerInformation()
			return currentCi,err
		}
	}

	currentServices,err := c.CheckUUIDList([]string{bt_uuid.NAP_UUID, bt_uuid.PANU_UUID, bt_uuid.GN_UUID})
	if err != nil { return currentCi, err }
	if newCi.ServiceNetworkServerNap != currentServices[0] {
		if newCi.ServiceNetworkServerNap {
			// register NAP
			err = c.RegisterNetworkServer(toolz.UUID_NETWORK_SERVER_NAP, bridgeNameNAP)
		} else {
			err = c.UnregisterNetworkServer(toolz.UUID_NETWORK_SERVER_NAP)
		}
		if err != nil { return nil, err }
	}
	if newCi.ServiceNetworkServerPanu != currentServices[1] {
		if newCi.ServiceNetworkServerPanu {
			// register NAP
			err = c.RegisterNetworkServer(toolz.UUID_NETWORK_SERVER_PANU, bridgeNamePANU)
		} else {
			err = c.UnregisterNetworkServer(toolz.UUID_NETWORK_SERVER_PANU)
		}
		if err != nil { return nil, err }
	}
	if newCi.ServiceNetworkServerGn != currentServices[2] {
		if newCi.ServiceNetworkServerGn {
			// register NAP
			err = c.RegisterNetworkServer(toolz.UUID_NETWORK_SERVER_GN, bridgeNameGN)
		} else {
			err = c.UnregisterNetworkServer(toolz.UUID_NETWORK_SERVER_GN)
		}
		if err != nil { return nil, err }
	}

	return c.ReadControllerInformation()
}

func (c *Controller) RegisterNetworkServer(uuid toolz.NetworkServerUUID, bridgeName string) (err error) {
	nw, err := toolz.NetworkServer(c.DBusPath)
	if err != nil {
		return
	}
	return nw.Register(uuid, bridgeName)
}

func (c *Controller) UnregisterNetworkServer(uuid toolz.NetworkServerUUID) (err error) {
	nw, err := toolz.NetworkServer(c.DBusPath)
	if err != nil {
		return
	}
	return nw.Unregister(uuid)
}


func (c *Controller) ConnectNetwork(deviceMac string, uuid toolz.NetworkServerUUID) (err error) {
	//convet given address to net.HardwareAddress
	searchAddr,err := net.ParseMAC(deviceMac)
	if err != nil { return err }

	dev,err := c.GetDeviceByAddr(searchAddr)
	if err != nil { return err }

	//get device path
	path := dev.GetPath()

	nw, err := toolz.Network(path)
	if err != nil {
		return
	}
	return nw.Connect(uuid)
}

func (c *Controller) DisconnectNetwork(deviceMac string) (err error) {
	//convet given address to net.HardwareAddress
	searchAddr,err := net.ParseMAC(deviceMac)
	if err != nil { return err }

	dev,err := c.GetDeviceByAddr(searchAddr)
	if err != nil { return err }

	//get device path
	path := dev.GetPath()

	nw, err := toolz.Network(path)
	if err != nil {
		return
	}
	return nw.Disconnect()
}


func (c *Controller) IsServerNAPEnabled() (res bool, err error) {
	uuids,err := c.GetUUIDs()
	if err != nil { return res,err }

	for _,uuid := range uuids {
		if uuid == bt_uuid.NAP_UUID {
			return true,nil
		}
	}
	return false,nil
}

func (c *Controller) IsServerPANUEnabled() (res bool, err error) {
	uuids,err := c.GetUUIDs()
	if err != nil { return res,err }

	for _,uuid := range uuids {
		if uuid == bt_uuid.PANU_UUID {
			return true,nil
		}
	}
	return false,nil
}

func (c *Controller) IsServerGNEnabled() (res bool, err error) {
	uuids,err := c.GetUUIDs()
	if err != nil { return res,err }

	for _,uuid := range uuids {
		if uuid == bt_uuid.GN_UUID {
			return true,nil
		}
	}
	return false,nil
}

//Check if given UUIDs are present on adapter
func (c *Controller) CheckUUIDList(uuidsToCheck []string) (res []bool, err error) {
	uuids,err := c.GetUUIDs()
	if err != nil { return res,err }

	//Convert to map for easy lookup
	uuidMap := make(map[string]interface{})

	for _,uuid := range uuids {
		uuidMap[uuid] = nil
	}

	res = make([]bool,len(uuidsToCheck))
	for idx,uuidToCheck := range uuidsToCheck {
		_,exists := uuidMap[uuidToCheck]
		res[idx] = exists
	}
	return
}


func (c *Controller) GetPathDevices() (results []dbus.ObjectPath, err error) {
	// grab DBus object path of all available Adapters (=controller) from DBus API
	om,err := dbusHelper.NewObjectManager()
	defer om.Close()
	if err != nil { return nil,err }
	pathDevices,err := om.GetAllObjectsPathOfInterface(toolz.DBusNameDevice1Interface)
	if err != nil { return nil,err }

	// iterate over path elements and chack if they belong to the current adapter
	results = make([]dbus.ObjectPath,0)
	for _,devDBusPath := range pathDevices {
		if !devDBusPath.IsValid() {continue}
		if strings.HasPrefix(string(devDBusPath), string(c.DBusPath)) {
			results = append(results, devDBusPath)
		}
	}

	return results,nil
}

func (c *Controller) GetDevices() (results []*toolz.Device1, err error) {
	dbusPathDevices,err := c.GetPathDevices()
	if err != nil { return results,err }

	results = make([]*toolz.Device1, len(dbusPathDevices))
	for i,pathDevice := range dbusPathDevices {
		dev,err := toolz.Device(pathDevice)
		if err != nil { return nil,err }
		results[i] = dev

	}

	return results,nil
}

func (c *Controller) GetDeviceByAddr(addr net.HardwareAddr) (res *toolz.Device1, err error) {
	// fetch all devices
	devs,err := c.GetDevices()
	if err != nil { return nil,err }
	// check if one of the devices uses the given mac
	for _,dev := range devs {
		devAddr,addrErr := dev.GetAddress()
		if addrErr != nil { continue }

		if compareHwAddr(addr, devAddr) {
			//same addresses
			return dev,nil
		}
	}
	return nil,ErrDeviceNotFOund
}




func FindFirstAvailableController() (ctl *Controller, err error) {
	// use btmgmt to fetch first controller index
	mgmt,err := btmgmt.NewBtMgmt()
	if err != nil { return nil, err }
	cil,err := mgmt.ReadControllerIndexList()
	if err != nil { return nil, err }

	ctl = &Controller{}
	if len(cil.Indices) > 0 {
		ctl.Index = cil.Indices[0]
	} else {
		return nil, ErrBtSvcNotAvailable
	}
	ctl.mgmt = mgmt

	// retrieve additional info for the controller from mgmt-api
	ci,err := mgmt.ReadControllerInformation(ctl.Index)
	if err != nil { return nil,err }

	// grab DBus object path of all available Adapters (=controller) from DBus API
	om,err := dbusHelper.NewObjectManager()
	defer om.Close()
	if err != nil { return nil,err }
	pathAdapters,err := om.GetAllObjectsPathOfInterface(toolz.DBusNameAdapter1Interface)
	if err != nil { return nil,err }


	for _,pathAdapter := range pathAdapters {
		// create adapter object
		adp,err := toolz.Adapter(pathAdapter)
		if err != nil {	continue } // skip adapter
		hciAdapterAddr,err := adp.GetAddress()
		if err != nil {
			adp.Close()
			continue
		} // skip adapter
		// compare address of Controller from mmgmt-api with the adapter from adapter-api
		if compareHwAddr(hciAdapterAddr, ci.Address.Addr) {
			ctl.DBusPath = pathAdapter
			ctl.adapter = adp
			break // exit for loop
		} else {
			adp.Close()
		}
	}
	if ctl.adapter == nil {
		return nil,errors.New("Found controller via 'bluetooth management socket', but no match on DBus API")
	}

	ctl.deploySettingsFromControllerInformationLock = &sync.Mutex{}
	return
}



func BluetoothControllerInformationToRpc(src *btmgmt.ControllerInformation) (target *pb.BluetoothControllerInformation) {
	target = &pb.BluetoothControllerInformation{
		IsAvailable: false,
		ClassOfDevice: src.ClassOfDevice.Octets,
		BluetoothVersion: uint32(src.BluetoothVersion),
		Address: src.Address.Addr,
		Manufacturer: uint32(src.Manufacturer),
		Name: src.Name,
		ShortName: src.ShortName,
		SupportedSettings: BluetoothControllerSettingsToRpc(&src.SupportedSettings),
		CurrentSettings: BluetoothControllerSettingsToRpc(&src.CurrentSettings),
		ServiceNetworkServerGn: src.ServiceNetworkServerGn,
		ServiceNetworkServerPanu: src.ServiceNetworkServerPanu,
		ServiceNetworkServerNap: src.ServiceNetworkServerNap,
	}
	return
}

func BluetoothControllerInformationFromRpc(src *pb.BluetoothControllerInformation) (target *btmgmt.ControllerInformation) {
	// Only changable settings are regarded
	target = &btmgmt.ControllerInformation{
		Name: src.Name,
		CurrentSettings: *BluetoothControllerSettingsFromRpc(src.CurrentSettings),
		ServiceNetworkServerGn: src.ServiceNetworkServerGn,
		ServiceNetworkServerPanu: src.ServiceNetworkServerPanu,
		ServiceNetworkServerNap: src.ServiceNetworkServerNap,
	}
	return
}

func BluetoothControllerSettingsToRpc(src *btmgmt.ControllerSettings) (target *pb.BluetoothControllerSettings) {
	target = &pb.BluetoothControllerSettings{
		StaticAddress: src.StaticAddress,
		ControllerConfiguration: src.ControllerConfiguration,
		Privacy: src.Privacy,
		Powered: src.Powered,
		DebugKeys: src.DebugKeys,
		Discoverable: src.Discoverable,
		Bondable: src.Bondable,
		SecureConnections: src.SecureConnections,
		Advertising: src.Advertising,
		LowEnergy: src.LowEnergy,
		HighSpeed: src.HighSpeed,
		BrEdr: src.BrEdr,
		SecureSimplePairing: src.SecureSimplePairing,
		LinkLevelSecurity: src.LinkLevelSecurity,
		Connectable: src.Connectable,
		FastConnectable: src.FastConnectable,
	}
	return
}

func BluetoothControllerSettingsFromRpc(src *pb.BluetoothControllerSettings) (target *btmgmt.ControllerSettings) {
	target = &btmgmt.ControllerSettings{
		StaticAddress: src.StaticAddress,
		ControllerConfiguration: src.ControllerConfiguration,
		Privacy: src.Privacy,
		Powered: src.Powered,
		DebugKeys: src.DebugKeys,
		Discoverable: src.Discoverable,
		Bondable: src.Bondable,
		SecureConnections: src.SecureConnections,
		Advertising: src.Advertising,
		LowEnergy: src.LowEnergy,
		HighSpeed: src.HighSpeed,
		BrEdr: src.BrEdr,
		SecureSimplePairing: src.SecureSimplePairing,
		LinkLevelSecurity: src.LinkLevelSecurity,
		Connectable: src.Connectable,
		FastConnectable: src.FastConnectable,
	}
	return
}
