package service

import (
	pb "../proto"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"log"
	"errors"
	"strconv"
	"syscall"
	"strings"
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
	return NameLeaseFileDHCPSrv(s.ListenInterface) //default lease file
}

func NameLeaseFileDHCPSrv(nameIface string) (lf string) {
	return fmt.Sprintf("/tmp/dnsmasq_%s.leases", nameIface)
}


func NameConfigFileDHCPSrv(nameIface string) string {
	return fmt.Sprintf("/tmp/dnsmasq_%s.conf", nameIface)
}


func StartDHCPClient(nameIface string) (err error) {
	log.Printf("Starting DHCP client for interface '%s'...\n", nameIface)

	//check if interface is valid
	if_exists,_ := CheckInterfaceExistence(nameIface)
	if !if_exists {
		return errors.New(fmt.Sprintf("The given interface '%s' doesn't exist", nameIface))
	}


	//We use the run command and allow dhcpcd to daemonize

	proc := exec.Command("/sbin/dhcpcd", "-C", "wpa_supplicant", nameIface) //we avoid starting wpa_supplicant along with the dhcp client
	dhcpcd_out, err := proc.CombinedOutput()
	//err = proc.Run()
	if err != nil { return err}
	fmt.Printf("Dhcpcd output for %s:\n%s", nameIface, dhcpcd_out)


	log.Printf("... DHCP client for interface '%s' started\n", nameIface)
	return nil
}

func IsDHCPClientRunning(nameIface string) (running bool, pid int, err error) {
	if_exists,_ := CheckInterfaceExistence(nameIface)
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


func StopDHCPClient(nameIface string) (err error) {
	log.Printf("Stoping DHCP client for interface '%s'...\n", nameIface)

	//check if interface is valid
	if_exists,_ := CheckInterfaceExistence(nameIface)
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



func StartDHCPServer(nameIface string, configPath string) (err error)  {
	log.Printf("Starting DHCP server for interface '%s' with config '%s'...\n", nameIface, configPath)

	//check if interface is valid
	if_exists,_ := CheckInterfaceExistence(nameIface)
	if !if_exists {
		return errors.New(fmt.Sprintf("The given interface '%s' doesn't exist", nameIface))
	}

	//Check if there's already a DHCP server running for the given interface
	running, _, err := IsDHCPServerRunning(nameIface)
	if err != nil { return errors.New(fmt.Sprintf("Error fetching state of DHCP server: %v\n", err)) }
	if running {return errors.New(fmt.Sprintf("Error starting DHCP server for interface '%s', there is already a DHCP server running\n", nameIface))}

	//We use the run command and allow dnsmasq to daemonize
	proc := exec.Command("/usr/sbin/dnsmasq", "-x", pidFileDHCPSrv(nameIface), "-C", configPath)
	//dnsmasq_out, err := proc.CombinedOutput()
	err = proc.Run()
	if err != nil { return err}
	//fmt.Printf("Dnsmasq out %s\n", dnsmasq_out)


	log.Printf("... DHCP server for interface '%s' started\n", nameIface)
	return nil
}


func IsDHCPServerRunning(nameIface string) (running bool, pid int, err error) {
	if_exists,_ := CheckInterfaceExistence(nameIface)
	if !if_exists {
		return false, 0, errors.New(fmt.Sprintf("The given interface '%s' doesn't exist", nameIface))
	}

	pid_file := pidFileDHCPSrv(nameIface)

	//Check if the pidFile exists
	if _, err := os.Stat(pid_file); os.IsNotExist(err) {
		return false, 0,nil //file doesn't exist, so we assume dnsmasq isn't running
	}

	//File exists, read the PID
	content, err := ioutil.ReadFile(pid_file)
	if err != nil { return false, 0, err}
	pid, err = strconv.Atoi(strings.TrimSuffix(string(content), "\n"))
	if err != nil { return false, 0, errors.New(fmt.Sprintf("Error parsing PID file %s: %v", pid_file, err))}

	//With PID given, check if the process is indeed running (pid_file could stay, even if the dnsmasq process has died already)
	err_kill := syscall.Kill(pid, 0) //sig 0: doesn't send a signal, but error checking is still performed
	switch err_kill{
	case nil:
		//ToDo: Check if the running process image is indeed dnsmasq
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

func StopDHCPServer(nameIface string) (err error)  {
	//don't check if interface is valid, to allow closing of orphaned DHCP procs (interface went down while running)
	log.Printf("Stopping DHCP server for interface '%s' ...\n", nameIface)
	running,pid,err := IsDHCPServerRunning(nameIface)
	if err != nil { return }

	if running {
		//send SIGTERM
		err = syscall.Kill(pid, syscall.SIGTERM)
		if err != nil { return }
	} else {
		log.Printf("... DHCP server for interface '%s' wasn't started\n", nameIface)
	}

	running,pid,err = IsDHCPServerRunning(nameIface)
	if err != nil { return }
	if (running) {
		log.Printf("... couldn't terminate DHCP server for interface '%s'\n", nameIface)
	} else {
		log.Printf("... DHCP server for interface '%s' stopped\n", nameIface)
	}

	//Delete PID file
	os.Remove(pidFileDHCPSrv(nameIface))

	//Deleting leaseFile
	os.Remove(NameLeaseFileDHCPSrv(nameIface))
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
		//ToDo: check rangeLower + rangeUpper to be valid IP addresses
		if len(pRange.LeaseTime) > 0 {
			config += fmt.Sprintf("dhcp-range=%s,%s,%s\n", pRange.RangeLower, pRange.RangeUpper, pRange.LeaseTime)
		} else {
			//default to 5 minute lease
			config += fmt.Sprintf("dhcp-range=%s,%s,5m\n", pRange.RangeLower, pRange.RangeUpper)
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


	return
}

