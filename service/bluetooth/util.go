package bluetooth

import "net"

func compareHwAddr(a net.HardwareAddr, b net.HardwareAddr) bool {
	for idx,_ := range a {
		if a[idx] != b[idx] { return false }
	}
	return true
}
