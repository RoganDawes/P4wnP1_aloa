package bluetooth

import (
	"fmt"
	"github.com/godbus/dbus"
	"github.com/mame82/mblue-toolz/bt_uuid"
	"github.com/mame82/mblue-toolz/dbusHelper"
	"github.com/mame82/mblue-toolz/toolz"
)

type DefaultAgent struct {
	pincode    string // only used if Secure Simple Pairing is disabled and currentCap = DISPLAY_ONLY
	currentCap toolz.AgentCapability
}

// ------------ START OF AGENT INTERFACE IMPLEMENTATION ------------
func (a *DefaultAgent) RegistrationPath() string {
	return toolz.AgentDefaultRegisterPath
}

func (a *DefaultAgent) Release() *dbus.Error {
	fmt.Println("DefaultAgent release called")
	return nil
}

func (a *DefaultAgent) RequestPinCode(device dbus.ObjectPath) (pincode string, err *dbus.Error) {
	fmt.Printf("DefaultAgent request pincode called, returning string '%s'\n", a.pincode)
	// Called when SSP is off and currentCap != CAP_NO_INPUT_NO_OUTPUT, we could use a pre-generated PIN
	// and the remote device has to enter the same one
	return a.pincode, nil
}

func (a *DefaultAgent) DisplayPinCode(device dbus.ObjectPath, pincode string) *dbus.Error {
	fmt.Printf("DefaultAgent display pincode called, code: '%s'\n", pincode)
	return nil
}

func (a *DefaultAgent) RequestPasskey(device dbus.ObjectPath) (passkey uint32, err *dbus.Error) {
	fmt.Println("DefaultAgent request passkey called, returning integer 1337")
	// Called with SSP on and Cap == AGENT_CAP_KEYBOARD_ONLY
	// The needed passkey is random, we can't know the correct return value upfront (in contrast to a PIN
	// which has to be entered on both devices and match)

	return 1337, nil
}

func (a *DefaultAgent) DisplayPasskey(device dbus.ObjectPath, passkey uint32, entered uint16) *dbus.Error {
	fmt.Printf("DefaultAgent display passkey called, passkey: %d\n", passkey)
	return nil
}

func (a *DefaultAgent) RequestConfirmation(device dbus.ObjectPath, passkey uint32) *dbus.Error {
	fmt.Printf("DefaultAgent request confirmation called for passkey: %d\n", passkey)
	// Called when SSP on and
	// currentCap == AGENT_CAP_DISPLAY_ONLY ||
	// currentCap == AGENT_CAP_KEYBOARD_DISPLAY ||
	// currentCap == AGENT_CAP_DISPLAY_YES_NO
	fmt.Println("... rejecting passkey")
	return toolz.ErrRejected
}

func (a *DefaultAgent) RequestAuthorization(device dbus.ObjectPath) *dbus.Error {
	fmt.Println("DefaultAgent request authorization called")
	fmt.Println("... accepting")
	return nil
}

func (a *DefaultAgent) AuthorizeService(device dbus.ObjectPath, uuid string) *dbus.Error {
	devStr,_ := dbusHelper.DBusDevPathToHwAddr(device) // ignore error
	/*
	// alternate way to retrieve the device address (and call functions of the device)
	if d,e := toolz.Device(device); e == nil {
		addr,_ := d.GetAddress()
		fmt.Println(addr)
	}
	*/

	fmt.Printf("DefaultAgent authorize service called for UUID: %s from %s\n", uuid, devStr)
	switch uuid {
	case bt_uuid.BNEP_SVC_UUID:
		fmt.Println("... granting BNEP access")
		return nil
	default:
		fmt.Println("... rejecting")
		return toolz.ErrRejected
	}
	return toolz.ErrRejected
}

func (a *DefaultAgent) Cancel() *dbus.Error {
	fmt.Println("DefaultAgent cancel called")
	return nil
}
// ------------ END OF AGENT INTERFACE IMPLEMENTATION ------------

func (a *DefaultAgent) Start(cap toolz.AgentCapability) (err error) {
	a.currentCap = cap
	return toolz.RegisterDefaultAgent(a, cap)
}

func (a *DefaultAgent) Stop() (err error) {
	return toolz.UnregisterAgent(a.RegistrationPath())
}


func (a *DefaultAgent) SetPIN(pin string)  {
	a.pincode = pin
}

func (a *DefaultAgent) GetPIN() (pin string)  {
	return a.pincode
}

func NewDefaultAgent(pincode string) (res *DefaultAgent) {
	return &DefaultAgent{
		pincode:    pincode,
		currentCap: toolz.AGENT_CAP_NO_INPUT_NO_OUTPUT,
	}
}