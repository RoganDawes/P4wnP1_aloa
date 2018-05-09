package cli_client

import (
	"github.com/spf13/cobra"
	pb "../proto"
	"net"
	"errors"
	"fmt"
	"google.golang.org/grpc/status"
)

//Empty settings used to store cobra flags
var (

	tmpStrInterface string = ""
	tmpStrAddress4 string = ""
	tmpStrNetmask4 string = ""
	tmpDisabled bool = false
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

	fmt.Printf("Deployin ethernet inteface settings:\n\t%v\n", settings)

	err = ClientDeployEthernetInterfaceSettings(StrRemoteHost, StrRemotePort, settings)
	if err != nil {
		fmt.Println(status.Convert(err).Message())
	}
	return
}

func cobraNetSetDHCPClient(cmd *cobra.Command, args []string) {
	return
}

func cobraNetSetDHCPServer(cmd *cobra.Command, args []string) {
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
}
