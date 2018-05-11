package cli_client

import (
	"github.com/spf13/cobra"
	pb "../proto"
	"net"
	"errors"
	"fmt"
	"google.golang.org/grpc/status"
	"../service"
	"strings"
	"strconv"
)

//Empty settings used to store cobra flags
var (

	tmpStrInterface string = ""
	tmpStrAddress4 string = ""
	tmpStrNetmask4 string = ""
	tmpDisabled bool = false
	tmpDHCPSrvOptions []string = []string{}
	tmpDHCPSrvRanges []string = []string{}

)

/*
func init(){
	//Configure spew for struct deep printing (disable using printer interface for gRPC structs)
	spew.Config.Indent="\t"
	spew.Config.DisableMethods = true
	spew.Config.DisablePointerAddresses = true
}
*/

// usbCmd represents the usb command
var netCmd = &cobra.Command{
	Use:   "NET",
	Short: "Configure Network settings of ethernet interfaces (including USB ethernet if enabled)",
}

var netSetCmd = &cobra.Command{
	Use:   "set",
	Short: "set ethernet settings",
	Long: ``,
}

var netSetManualCmd = &cobra.Command{
	Use:   "manual",
	Short: "Configure given interface manually",
	Long: ``,
	Run: cobraNetSetManual,
}


var netSetDHCPClientCmd = &cobra.Command{
	Use:   "client",
	Short: "Configure given interface to run a DHCP client",
	Long: ``,
	Run: cobraNetSetDHCPClient,
}

var netSetDHCPServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Configure given interface to run a DHCP server",
	Long: ``,
	Run: cobraNetSetDHCPServer,
}
var netGetCmd = &cobra.Command{
	Use:   "get",
	Short: "get ethernet settings",
	Long: ``,
	Run: cobraNetGet,
}

func cobraNetGet(cmd *cobra.Command, args []string) {
	return
}

func cobraNetSetManual(cmd *cobra.Command, args []string) {
	settings, err := createManualSettings(tmpStrInterface, tmpStrAddress4, tmpStrNetmask4, tmpDisabled)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Deploying ethernet inteface settings:\n\t%v\n", settings)

	err = ClientDeployEthernetInterfaceSettings(StrRemoteHost, StrRemotePort, settings)
	if err != nil {
		fmt.Println(status.Convert(err).Message())
	}
	return
}

func cobraNetSetDHCPClient(cmd *cobra.Command, args []string) {
	settings, err := createDHCPClientSettings(tmpStrInterface, tmpDisabled)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Deploying ethernet inteface settings:\n\t%v\n", settings)

	err = ClientDeployEthernetInterfaceSettings(StrRemoteHost, StrRemotePort, settings)
	if err != nil {
		fmt.Println(status.Convert(err).Message())
	}
	return
}

func cobraNetSetDHCPServer(cmd *cobra.Command, args []string) {
	settings, err := createDHCPServerSettings(tmpStrInterface, tmpStrAddress4, tmpStrNetmask4, tmpDisabled, tmpDHCPSrvRanges, tmpDHCPSrvOptions)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Deploying ethernet inteface settings:\n\t%v\n", settings)

	err = ClientDeployEthernetInterfaceSettings(StrRemoteHost, StrRemotePort, settings)
	if err != nil {
		fmt.Println(status.Convert(err).Message())
	}
	return


}

func createManualSettings(iface string, ip4 string, mask4 string, disabled bool) (settings *pb.EthernetInterfaceSettings, err error) {
	if len(iface) == 0 { return nil,errors.New("Please provide a network interface name") }
	err = checkIPv4(ip4)
	if err != nil { return nil,errors.New("Please provide a valid IPv4 address for the interface") }
	err = checkIPv4(mask4)
	if err != nil { return nil,errors.New("Please provide a valid IPv4 netmask for the interface") }

	settings = &pb.EthernetInterfaceSettings {
		Mode: pb.EthernetInterfaceSettings_MANUAL,
		Name: iface,
		IpAddress4: ip4,
		Netmask4: mask4,
		Enabled: !disabled,
		DhcpServerSettings: nil,
	}
	return
}

func parseDhcpServerOptions(strOptions []string) (options map[uint32]string, err error) {
	options = map[uint32]string{}
	for _, strOpt := range strOptions {
		splOpt := strings.SplitN(strOpt,":", 2)
		if len(splOpt) == 0 {
			return nil,errors.New(fmt.Sprintf("Invalid DHCP option: %s\nOption format is \"<DHCPOptionNumber>:[DHCPOptionValue]\"", strOpt))
		}
		optNum, err := strconv.Atoi(splOpt[0])
		if err != nil || optNum < 0 {
			return nil,errors.New(fmt.Sprintf("Invalid DHCP option Number: %s\nOption format is \"<DHCPOptionNumber>:[DHCPOptionValue]\"", splOpt[0]))
		}
		uOptNum := uint32(optNum)
		if len(splOpt) == 1 {
			options[uOptNum] = ""
		//	fmt.Printf("Setting DHCP server option %d to empty value (disabeling option %d)\n", uOptNum, uOptNum)
		} else {
			//Replace '|' with ',' (a comma couldn't be used as it'd be interpreted as slice delimiter)
			strOptNew := strings.Replace(splOpt[1], "|", ",",-1)

			options[uOptNum] = strOptNew
		//	fmt.Printf("Setting DHCP server option %d to '%s'\n", uOptNum, splOpt[1])
		}
	}

	return options,nil
}

func parseDhcpServerRanges(strRanges []string) (ranges []*pb.DHCPServerRange, err error) {
	ranges = []*pb.DHCPServerRange{}
	for _,strRange := range strRanges {
		splRange := strings.Split(strRange, "|")
		if len(splRange) != 3 && len(splRange) != 2 {
			return nil,errors.New(fmt.Sprintf("Invalid DHCP range: %s\nOption format is \"<first IPv4>|<last IPv4>[|leaseTime]\"", strRange))
		}

		if net.ParseIP(splRange[0]) == nil {
			return nil, errors.New(fmt.Sprintf("%s in range '%s' is no valid IP address", splRange[0], strRange))
		}
		if net.ParseIP(splRange[1]) == nil {
			return nil, errors.New(fmt.Sprintf("%s in range '%s' is no valid IP address", splRange[1], strRange))
		}
		pRange := &pb.DHCPServerRange{
			RangeLower: splRange[0],
			RangeUpper: splRange[1],
		}
		if len(splRange) > 2 {
			//ToDo: Regex check lease time to be valid [0-9]+[mh]
			pRange.LeaseTime = splRange[2]
		}

		ranges = append(ranges, pRange)
	}
	return ranges,nil
}


func createDHCPServerSettings(iface string, ip4 string, mask4 string, disabled bool, strRanges []string, strOptions []string) (settings *pb.EthernetInterfaceSettings, err error) {
	if len(iface) == 0 { return nil,errors.New("Please provide a network interface name") }
	err = checkIPv4(ip4)
	if err != nil { return nil,errors.New("Please provide a valid IPv4 address for the interface") }
	err = checkIPv4(mask4)
	if err != nil { return nil,errors.New("Please provide a valid IPv4 netmask for the interface") }


	options, err := parseDhcpServerOptions(strOptions)
	if err != nil {
		return nil, err
	}
	ranges, err := parseDhcpServerRanges(strRanges)
	if err != nil {
		return nil, err
	}


	settings = &pb.EthernetInterfaceSettings {
		Mode: pb.EthernetInterfaceSettings_DHCP_SERVER,
		Name: iface,
		IpAddress4: ip4,
		Netmask4: mask4,
		Enabled: !disabled,
		DhcpServerSettings: &pb.DHCPServerSettings{
			ListenInterface:iface,
			LeaseFile: service.NameLeaseFileDHCPSrv(iface),
			CallbackScript: "",
			ListenPort: 0, //Disable DNS
			DoNotBindInterface: false, //only listen on given interface
			NotAuthoritative: false, // be authoritative
			Ranges: ranges,
			Options: options,
		},
	}
	return
}

func createDHCPClientSettings(iface string, disabled bool) (settings *pb.EthernetInterfaceSettings, err error) {
	if len(iface) == 0 { return nil,errors.New("Please provide a network interface name") }

	settings = &pb.EthernetInterfaceSettings {
		Mode: pb.EthernetInterfaceSettings_DHCP_CLIENT,
		Name: iface,
		Enabled: !disabled,
		DhcpServerSettings: nil,
	}
	return
}

func checkIPv4(ip4 string) error {
	ip := net.ParseIP(ip4)
	if ip == nil {
		return errors.New(fmt.Sprintf("Error parsing IP address '%s'\n",ip4))
	}
	if ip.To4() == nil {
		return errors.New(fmt.Sprintf("Not an IPv4 address '%s'\n",ip4))
	}
	return nil
}


func init() {
	rootCmd.AddCommand(netCmd)
	netCmd.AddCommand(netGetCmd)
	netCmd.AddCommand(netSetCmd)
	netSetCmd.AddCommand(netSetManualCmd)
	netSetCmd.AddCommand(netSetDHCPClientCmd)
	netSetCmd.AddCommand(netSetDHCPServerCmd)

	netSetCmd.PersistentFlags().StringVarP(&tmpStrInterface, "interface","i", "", "The name of the ethernet interface to work on")
	netSetCmd.PersistentFlags().BoolVarP(&tmpDisabled, "disable","d", false, "The flag disables the given interface (omitting the flag enables the interface")
	netSetManualCmd.Flags().StringVarP(&tmpStrAddress4, "address","a", "", "The IPv4 address to use for the interface")
	netSetManualCmd.Flags().StringVarP(&tmpStrNetmask4, "netmask","m", "", "The IPv4 netmask to use for the interface")
	netSetDHCPServerCmd.Flags().StringVarP(&tmpStrAddress4, "address","a", "", "The IPv4 address to use for the interface")
	netSetDHCPServerCmd.Flags().StringVarP(&tmpStrNetmask4, "netmask","m", "", "The IPv4 netmask to use for the interface")
	netSetDHCPServerCmd.Flags().StringSliceVarP(&tmpDHCPSrvRanges, "range", "r",[]string{""}, "A DHCP Server range in form \"<lowest IPv4>|<highest IPv4>[|lease time]\" (the flag could be used multiple times)")
	netSetDHCPServerCmd.Flags().StringSliceVarP(&tmpDHCPSrvOptions, "option", "o",[]string{""}, "A DHCP Server option in form \"<option number>:<value1|value2>\" (Option values have to be separated by '|' not by ','. The flag could be used multiple times)")
}
