package bluetooth

import "errors"

var (
	ErrBtSvcNotAvailable = errors.New("No bluetooth available")
	ErrBtNoMgmt          = errors.New("Couldn't access bluetooth mgmt-api")
	ErrChgSetting        = errors.New("Couldn't change controller setting to intended value")
	ErrReadSetting        = errors.New("Couldn't read controller setting")
	ErrDeviceNotFOund        = errors.New("Couldn't find given device")
)

