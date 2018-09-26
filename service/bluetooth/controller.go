package bluetooth

import (
	"github.com/godbus/dbus"
	"net"

	//"github.com/mame82/P4wnP1_go/service"
	"github.com/mame82/mblue-toolz/btmgmt"
	"github.com/mame82/mblue-toolz/dbusHelper"
	"github.com/mame82/mblue-toolz/toolz"
	"github.com/pkg/errors"
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
}

func (c *Controller) SetSSP(val bool) (err error) {
	mgmt,err := btmgmt.NewBtMgmt()
	if err != nil { return ErrChgSetting }

	s,err := mgmt.SetSecureSimplePairing(c.Index, val)
	if err != nil || s.SecureSimplePairing != val {
		return ErrChgSetting
	}
	return
}

func (c *Controller) SetHighSpeed(val bool) (err error) {
	mgmt,err := btmgmt.NewBtMgmt()
	if err != nil { return ErrChgSetting }

	s,err := mgmt.SetHighSpeed(c.Index, val)
	if err != nil || s.HighSpeed != val {
		return ErrChgSetting
	}
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

	return
}
