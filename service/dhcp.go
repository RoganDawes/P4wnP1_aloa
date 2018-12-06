// +build linux

package service

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/mame82/P4wnP1_aloa/common_web"
	pb "github.com/mame82/P4wnP1_aloa/proto"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

/*
The DHCP server part relies on "dnsmasq" and thus depends on the binary
Note: dnsmasq default option have to be disabled explicitly if not needed, by setting an empty value (1, 3, 6, 12, 28)

The DHCP client relies on dhcpcd binary
*/

func pidFileDHCPSrv(nameIface string) string {
	return fmt.Sprintf("/var/run/dnsmasq_%s.pid", nameIface)
}

func pidFileDHCPClient(nameIface string) string {
	return fmt.Sprintf("/var/run/dhcpcd-%s.pid", nameIface)
}

func leaseFileDHCPSrv(s *pb.DHCPServerSettings) (lf string) {
	return common_web.NameLeaseFileDHCPSrv(s.ListenInterface) //default lease file
}

func NameConfigFileDHCPSrv(nameIface string) string {
	return fmt.Sprintf("/tmp/dnsmasq_%s.conf", nameIface)
}


func (nim *NetworkInterfaceManager) StartDHCPClient() (err error) {
	nameIface := nim.InterfaceName
	log.Printf("Starting DHCP client for interface '%s'...\n", nameIface)

	//check if interface is valid
	if_exists := CheckInterfaceExistence(nameIface)
	if !if_exists {
		return errors.New(fmt.Sprintf("The given interface '%s' doesn't exist", nameIface))
	}


	//We use the run command and allow dhcpcd to daemonize
	proc := exec.Command("/sbin/dhcpcd", "-b", "-C", "wpa_supplicant", nameIface) //we avoid starting wpa_supplicant along with the dhcp client
	dhcpcd_out, err := proc.CombinedOutput()
	//err = proc.Run()
	if err != nil { return err}
	fmt.Printf("Dhcpcd output for %s:\n%s", nameIface, dhcpcd_out)


	log.Printf("... DHCP client for interface '%s' started\n", nameIface)
	return nil
}

func (nim *NetworkInterfaceManager) IsDHCPClientRunning() (running bool, pid int, err error) {
	nameIface := nim.InterfaceName
	if_exists := CheckInterfaceExistence(nameIface)
	if !if_exists {
		return false, 0, errors.New(fmt.Sprintf("The given interface '%s' doesn't exist", nameIface))
	}

	pid_file := pidFileDHCPClient(nameIface)

	//Check if the pidFile exists
	if _, err := os.Stat(pid_file); os.IsNotExist(err) {
		return false, 0,nil //file doesn't exist, so we assume dhcpcd isn't running
	}

	//File exists, read the PID
	content, err := ioutil.ReadFile(pid_file)
	if err != nil { return false, 0, err}
	pid, err = strconv.Atoi(strings.TrimSuffix(string(content), "\n"))
	if err != nil { return false, 0, errors.New(fmt.Sprintf("Error parsing PID file %s: %v", pid_file, err))}

	//With PID given, check if the process is indeed running (pid_file could stay, even if the process has died already)
	err_kill := syscall.Kill(pid, 0) //sig 0: doesn't send a signal, but error checking is still performed
	switch err_kill{
	case nil:
		//ToDo: Check if the running process image is indeed dhcpcd
		return true, pid, nil //Process is running
	case syscall.ESRCH:
		//Process doesn't exist
		return false, pid, nil
	case syscall.EPERM:
		//process exists, but we have no access permission
		return true, pid, err_kill
	default:
		return false, pid, err_kill
	}
}


func (nim *NetworkInterfaceManager) StopDHCPClient() (err error) {
	nameIface := nim.InterfaceName
	log.Printf("Stopping DHCP client for interface '%s'...\n", nameIface)

	//check if interface is valid
	if_exists := CheckInterfaceExistence(nameIface)
	if !if_exists {
		return errors.New(fmt.Sprintf("The given interface '%s' doesn't exist", nameIface))
	}


	//We use the run command and allow dhcpcd to daemonize
	proc := exec.Command("/sbin/dhcpcd", "-x", nameIface)
	dhcpcd_out, err := proc.CombinedOutput()
	//err = proc.Run()
	if err != nil { return err}
	fmt.Printf("Dhcpcd out for %s:\n%s", nameIface, dhcpcd_out)



	log.Printf("... DHCP client for interface '%s' stopped\n", nameIface)
	return nil
}

func (nim *NetworkInterfaceManager) StartDHCPServer(configPath string) (err error)  {
	nim.mutexDnsmasq.Lock()
	defer nim.mutexDnsmasq.Unlock()

	//stop dnsmasq if already running
	if nim.CmdDnsmasq != nil {
		// avoid deadlock
		nim.mutexDnsmasq.Unlock()
		nim.StopDHCPServer()
		nim.mutexDnsmasq.Lock()
	}

	nameIface := nim.InterfaceName
	log.Printf("Starting dnsmasq for interface '%s' with config '%s'...\n", nameIface, configPath)

	//check if interface is valid
	if_exists := CheckInterfaceExistence(nameIface)
	if !if_exists {
		return errors.New(fmt.Sprintf("The given interface '%s' doesn't exist", nameIface))
	}


	nim.CmdDnsmasq = exec.Command("/usr/sbin/dnsmasq", "--log-facility=-", "-k", "-x", pidFileDHCPSrv(nameIface), "-C", configPath)
	nim.CmdDnsmasq.Stdout = nim.LoggerDnsmasq.LogWriter
	nim.CmdDnsmasq.Stderr = nim.LoggerDnsmasq.LogWriter


	err = nim.CmdDnsmasq.Start()
	if err != nil {
		nim.CmdDnsmasq.Wait()
		return errors.New(fmt.Sprintf("Error starting dnsmasq '%v'", err))
	}



	log.Printf("... DHCP server for interface '%s' started\n", nameIface)
	return nil
}

func (nim *NetworkInterfaceManager) StopDHCPServer() (err error)  {
	eSuccess := fmt.Sprintf("... dnsmasq for interface '%s' stopped", nim.InterfaceName)
	eCantStop := fmt.Sprintf("... couldn't terminate dnsmasq for interface '%s'", nim.InterfaceName)

	log.Println("... killing dnsmasq")
	nim.mutexDnsmasq.Lock()
	defer nim.mutexDnsmasq.Unlock()

	if nim.CmdDnsmasq == nil {
		log.Printf("... dnsmasq for interface '%s' isn't running, no need to stop it\n", nim.InterfaceName)
		return nil
	}

	err = ProcSoftKill(nim.CmdDnsmasq, time.Second)
	if err != nil { return errors.New(eCantStop) }


	nim.CmdDnsmasq = nil
	log.Println(eSuccess)
	return nil
}

func DHCPCreateConfigFile(s *pb.DHCPServerSettings, filename string) (err error) {
	file_content, err := DHCPCreateConfigFileString(s)
	if err != nil {return}
	err = ioutil.WriteFile(filename, []byte(file_content), os.ModePerm)
	//ToDo: test config with `dnsmasq -C configfile --test`
	return
}

func DHCPCreateConfigFileString(s *pb.DHCPServerSettings) (config string, err error) {
	config = fmt.Sprintf("interface=%s\n", s.ListenInterface)
	//bind only o the given interface, except suppressed by `doNotBindInterface` option
	if !s.DoNotBindInterface { config += fmt.Sprintf("bind-interfaces\n") }
	config += fmt.Sprintf("port=%d\n", s.ListenPort)
	if len(s.CallbackScript) > 0 { config += fmt.Sprintf("dhcp-script=%s\n", s.CallbackScript) }
	if len(s.LeaseFile) > 0 {
		config += fmt.Sprintf("dhcp-leasefile=%s\n", s.LeaseFile)
	} else {
		config += fmt.Sprintf("dhcp-leasefile=%s\n", leaseFileDHCPSrv(s)) //default lease file
	}


	//Iterate over Ranges
	for _, pRange := range s.Ranges {
		//ToDo: regex check for leaseTime
		/*
		If the lease time is
              given, then leases will be given for that length of  time.  The
              lease  time is in seconds, or minutes (eg 45m) or hours (eg 1h)
              or "infinite". If not given, the  default  lease  time  is  one
              hour.  The  minimum lease time is two minutes
		 */
		//ToDo: check rangeLower + rangeUpper to be valid IP addresses
		if len(pRange.LeaseTime) > 0 {
			config += fmt.Sprintf("dhcp-range=%s,%s,%s\n", pRange.RangeLower, pRange.RangeUpper, pRange.LeaseTime)
		} else {
			//default to 5 minute lease
			config += fmt.Sprintf("dhcp-range=%s,%s\n", pRange.RangeLower, pRange.RangeUpper)
		}

	}

	//Iterate over options
	//
	// Note: for duplicates only the last one should be used, but not in any case, see
	// https://developers.google.com/protocol-buffers/docs/proto3#maps)
	//
	// "... When parsing from the wire or when merging, if there are duplicate map keys the last key seen is used.
	// When parsing a map from text format, parsing may fail if there are duplicate keys...."
	for o_num, o_val := range s.Options {
		o_str := fmt.Sprintf("dhcp-option=%d", o_num)
		if len(o_val) > 0 {
			o_str += fmt.Sprintf(",%s\n", o_val)
		} else {
			o_str += "\n"
		}
		config += o_str
	}
	config += fmt.Sprintf("log-dhcp\n") //extensive logging by default
	if (!s.NotAuthoritative) { config += fmt.Sprintf("dhcp-authoritative\n") }

	//Iterate over static hosts
	for _,host := range s.StaticHosts {
		config+=fmt.Sprintf("dhcp-host=%s,%s\n", host.Mac, host.Ip)
	}

	return
}

// Lease/Release tracker
var reLease = regexp.MustCompile(".*DHCPACK\\((.*)\\) ([0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}) ([0-9a-f]{2}:[0-9a-f]{2}:[0-9a-f]{2}:[0-9a-f]{2}:[0-9a-f]{2}:[0-9a-f]{2}) (.*)")
var reRelease = regexp.MustCompile(".*DHCPRELEASE\\((.*)\\) ([0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}) ([0-9a-f]{2}:[0-9a-f]{2}:[0-9a-f]{2}:[0-9a-f]{2}:[0-9a-f]{2}:[0-9a-f]{2})")


type DhcpLease struct {
	Release bool
	Iface string
	Ip net.IP
	Mac net.HardwareAddr
	Host string //only used for lease, not release
}

type dnsmasqLeaseMonitor struct {
	nim *NetworkInterfaceManager
}

func (m *dnsmasqLeaseMonitor) Write(p []byte) (n int, err error) {

	/*
	dnsmasq-wlan0: 16:53:49 dnsmasq-dhcp[1855]: 1450307105 DHCPACK(wlan0) 172.24.0.18 34:e6:xx:xx:xx:xx who-knows
	dnsmasq-wlan0: 16:53:58 dnsmasq-dhcp[1855]: 4200697351 DHCPRELEASE(wlan0) 172.24.0.18 34:e6:xx:xx:xx:xx
	 */
	lineScanner := bufio.NewScanner(bytes.NewReader(p))
	lineScanner.Split(bufio.ScanLines)
	for lineScanner.Scan() {
		line := string(lineScanner.Bytes())
		switch {
		case strings.Contains(line, "DHCPACK"):
			//fmt.Printf("Lease monitor %s LEASE: %s\n", m.nim.InterfaceName, line)

			leaseMatches := reLease.FindStringSubmatch(line)
			if len(leaseMatches) > 3 {
				lease := &DhcpLease{}
				lease.Iface = leaseMatches[1]
				lease.Ip = net.ParseIP(leaseMatches[2])
				mac,errP := net.ParseMAC(leaseMatches[3])
				if errP != nil { continue } //ignore if mac address couldn't be parsed
				lease.Mac = mac
				if len(leaseMatches) > 4 {
					//assume 4th match is hostname
					lease.Host = leaseMatches[4]
				}
				m.nim.OnHandedOutDhcpLease(lease)
			}

			/*
			for i,m := range leaseMatches {
				fmt.Printf("\tRegex lease %d: %s\n", i, m)
			}
			*/



		case strings.Contains(line, "DHCPRELEASE"):
			//fmt.Printf("Lease monitor %s RELEASE: %s\n", m.nim.InterfaceName, line)
			leaseMatches := reRelease.FindStringSubmatch(line)
			if len(leaseMatches) > 3 {
				release := &DhcpLease{}
				release.Iface = leaseMatches[1]
				release.Ip = net.ParseIP(leaseMatches[2])
				mac,errP := net.ParseMAC(leaseMatches[3])
				if errP != nil { continue } //ignore if mac address couldn't be parsed
				release.Mac = mac
				release.Release = true

				m.nim.OnReceivedDhcpRelease(release)
			}

		}

	}


	return len(p),nil
}


func NewDnsmasqLeaseMonitor(nim *NetworkInterfaceManager) *dnsmasqLeaseMonitor {
	return &dnsmasqLeaseMonitor{
		nim: nim,
	}
}
