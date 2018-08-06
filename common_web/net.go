package common_web

func NameLeaseFileDHCPSrv(nameIface string) (lf string) {
	return "/tmp/dnsmasq_" + nameIface + ".leases"
}
