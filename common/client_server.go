package common

import "fmt"

// File holds functions used by CLI client and P4wnP1 systemd service. Not to cross import the whole CLI/service
// package, is preferred over placing these functions in a contextual logic place.

func NameLeaseFileDHCPSrv(nameIface string) (lf string) {
	return fmt.Sprintf("/tmp/dnsmasq_%s.leases", nameIface)
}

