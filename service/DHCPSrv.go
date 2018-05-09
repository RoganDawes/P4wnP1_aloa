package service

import (
	pb "../proto"
	"fmt"
	"io/ioutil"
	"os"
)

/*
The DHCP server part relies on "dnsmasq" and thus depends on the binary
Note: dnsmasq default option have to be disabled explicitly if not needed, by setting an empty value (1, 3, 6, 12, 28)
 */

func defaultLeaseFile(s *pb.DHCPServerSettings) (lf string) {
	return fmt.Sprintf("/tmp/dnsmasq_%s.leases", s.ListenInterface) //default lease file
}

func DHCPCreateConfigFile(s *pb.DHCPServerSettings, filename string) (err error) {
	file_content, err := DHCPCreateConfigFileString(s)
	if err != nil {return}
	err = ioutil.WriteFile(filename, []byte(file_content), os.ModePerm)
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
		config += fmt.Sprintf("dhcp-leasefile=%s\n", defaultLeaseFile(s)) //default lease file
	}


	//Iterate over Ranges
	for _, pRange := range s.Ranges {
		config += fmt.Sprintf("dhcp-range=%s,%s,%s\n", pRange.RangeLower, pRange.RangeUpper, pRange.LeaseTime)
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

