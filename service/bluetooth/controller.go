package bluetooth

import (
	"fmt"
	"github.com/godbus/dbus"
	"net"
	"sync"
	"time"

	//"github.com/mame82/P4wnP1_go/service"
	"github.com/mame82/mblue-toolz/btmgmt"
	"github.com/mame82/mblue-toolz/dbusHelper"
	"github.com/mame82/mblue-toolz/toolz"
	"github.com/pkg/errors"
	pb "github.com/mame82/P4wnP1_go/proto"

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
	//fmt.Printf("ReadControllerInformation:\n%+v\n", ctlInfo)
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

func (c *Controller) UpdateSettingsFromChangedControllerInformation(newCi *btmgmt.ControllerInformation) (currentCi *btmgmt.ControllerInformation,err error) {
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
	if currentCi.CurrentSettings.SecureSimplePairing != newCi.CurrentSettings.SecureSimplePairing {
		err := c.SetSSP(newCi.CurrentSettings.SecureSimplePairing)
		if err != nil {
			fmt.Println("Error setting bluetooth SSP")
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

	return c.ReadControllerInformation()
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
	}
	return
}

func BluetoothControllerInformationFromRpc(src *pb.BluetoothControllerInformation) (target *btmgmt.ControllerInformation) {
	// Only changable settings are regarded
	target = &btmgmt.ControllerInformation{
		Name: src.Name,
		CurrentSettings: *BluetoothControllerSettingsFromRpc(src.CurrentSettings),
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
