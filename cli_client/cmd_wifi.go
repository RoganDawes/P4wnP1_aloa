package cli_client

import (
	"github.com/spf13/cobra"
	pb "github.com/mame82/P4wnP1_aloa/proto"
	"fmt"
	"google.golang.org/grpc/status"
	"os"
	"errors"
	"strings"
)

//Empty settings used to store cobra flags
var (
	tmpWifiStrReg        string = ""
	tmpWifiStrChannel    uint8  = 0
	tmpWifiHideSSID      bool   = false
	tmpWifiDisabled      bool   = false
	tmpWifiDisableNexmon bool   = false
	tmpWifiSSID          string = ""
	tmpWifiPSK           string = ""
)

/*
func init(){
	//Configure spew for struct deep printing (disable using printer interface for gRPC structs)
	spew.Config.Indent="\t"
	spew.Config.DisableMethods = true
	spew.Config.DisablePointerAddresses = true
}
*/

var wifiCmd = &cobra.Command{
	Use:   "wifi",
	Short: "Configure WiFi (spawn Access Point or join WiFi networks)",
}

var wifiSetCmd = &cobra.Command{
	Use:   "set",
	Short: "set WiFi settings",
	Long:  ``,
}

var wifiSetAPCmd = &cobra.Command{
	Use:   "ap",
	Short: "Configure WiFi interface as access point",
	Long:  ``,
	Run:   cobraWifiSetAP,
}

var wifiSetStaCmd = &cobra.Command{
	Use:   "sta",
	Short: "Configure WiFi interface to join a network as station",
	Long:  ``,
	Run:   cobraWifiSetSta,
}

var wifiGetCmd = &cobra.Command{
	Use:   "get",
	Short: "get WiFi settings",
	Long:  ``,
	Run:   cobraWifiGet,
}

func cobraWifiGet(cmd *cobra.Command, args []string) {
	return
}

func cobraWifiSetAP(cmd *cobra.Command, args []string) {
	settings, err := createWifiAPSettings(tmpWifiStrChannel, tmpWifiStrReg, tmpWifiSSID, tmpWifiPSK, tmpWifiHideSSID, tmpWifiDisableNexmon, tmpWifiDisabled)
	if err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(-1) //exit with error
		return
	}

	fmt.Printf("Deploying WiFi inteface settings:\n\t%v\n", settings)

	state, err := ClientDeployWifiSettings(StrRemoteHost, StrRemotePort, settings)
	if err != nil {
		fmt.Println(status.Convert(err).Message())
		os.Exit(-1) //exit with error
	} else {
		fmt.Printf("%+v\n", state)
	}
	return
}

func cobraWifiSetSta(cmd *cobra.Command, args []string) {
	settings, err := createWifiStaSettings(tmpWifiStrReg, tmpWifiSSID, tmpWifiPSK, tmpWifiDisableNexmon, tmpWifiDisabled)

	if err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(-1) //exit with error
		return
	}

	fmt.Printf("Deploying WiFi inteface settings:\n\t%v\n", settings)

	state,err := ClientDeployWifiSettings(StrRemoteHost, StrRemotePort, settings)
	if err != nil {
		fmt.Println(status.Convert(err).Message())
		os.Exit(-1) //exit with error
	} else {
		fmt.Printf("%+v\n", state)
	}
	return
}

func createWifiAPSettings(channel uint8, reg string, strSSID string, strPSK string, hideSsid bool, nonexmon bool, disabled bool) (settings *pb.WiFiSettings, err error) {
	if channel < 1 || channel > 14 {
		return nil, errors.New(fmt.Sprintf("Only 2.4GHz channels between 1 and 14 are supported, but '%d' was given\n", channel))
	}

	if len(reg) != 2 {
		return nil, errors.New(fmt.Sprintf("Regulatory domain has to consist of two uppercase letters (ISO/IEC 3166-1 alpha2), but '%s' was given\n", reg))
	}
	reg = strings.ToUpper(reg)

	if len(strSSID) < 1 || len(strSSID) > 32 {
		return nil, errors.New(fmt.Sprintf("SSID has to consist of 1 to 32 ASCII letters (even if hidden), but '%s' was given\n", strSSID))
	}

	if len(strPSK) > 0 && len(strPSK) < 8 {
		return nil, errors.New(fmt.Sprintf("A non-empty PSK implies WPA2 and has to have a minimum of 8 characters, but given PSK has '%d' charactres\n", len(strPSK)))
	}

	settings = &pb.WiFiSettings{
		WorkingMode: pb.WiFiWorkingMode_AP,
		AuthMode:    pb.WiFiAuthMode_OPEN,
		Disabled:    disabled,
		Regulatory:  reg,
		Channel:     uint32(channel),
		HideSsid:    hideSsid,
		Ap_BSS: &pb.WiFiBSSCfg{
			SSID: strSSID,
			PSK:  strPSK,
		},
		Client_BSSList: []*pb.WiFiBSSCfg{},
		Nexmon:         !nonexmon,
		Name:           "default",
	}

	if len(strPSK) > 0 {
		settings.AuthMode = pb.WiFiAuthMode_WPA2_PSK //if PSK is given use WPA2
	}

	return settings, err
}

func createWifiStaSettings(reg string, strSSID string, strPSK string, nonexmon bool, disabled bool) (settings *pb.WiFiSettings, err error) {
	if len(reg) != 2 {
		return nil, errors.New(fmt.Sprintf("Regulatory domain has to consist of two uppercase letters (ISO/IEC 3166-1 alpha2), but '%s' was given\n", reg))
	}
	reg = strings.ToUpper(reg)

	if len(strSSID) < 1 || len(strSSID) > 32 {
		return nil, errors.New(fmt.Sprintf("SSID has to consist of 1 to 32 ASCII letters (even if hidden), but '%s' was given\n", strSSID))
	}

	if len(strPSK) > 0 && len(strPSK) < 8 {
		return nil, errors.New(fmt.Sprintf("A non-empty PSK implies WPA2 and has to have a minimum of 8 characters, but given PSK has '%d' charactres\n", len(strPSK)))
	}

	settings = &pb.WiFiSettings{
		WorkingMode: pb.WiFiWorkingMode_STA,
		AuthMode:    pb.WiFiAuthMode_OPEN,
		Disabled:    disabled,
		Regulatory:  reg,
		Client_BSSList: []*pb.WiFiBSSCfg{
			&pb.WiFiBSSCfg{
				SSID: strSSID,
				PSK:  strPSK,
			},
		},
		Nexmon: !nonexmon,
		Ap_BSS: &pb.WiFiBSSCfg{}, //not needed
		Name: "default",
		HideSsid: false,
	}

	if len(strPSK) > 0 {
		settings.AuthMode = pb.WiFiAuthMode_WPA2_PSK //if PSK is given use WPA2
	}

	return settings, err
}

func init() {
	rootCmd.AddCommand(wifiCmd)
	//wifiCmd.AddCommand(wifiGetCmd)
	wifiCmd.AddCommand(wifiSetCmd)
	wifiSetCmd.AddCommand(wifiSetAPCmd)
	wifiSetCmd.AddCommand(wifiSetStaCmd)

	wifiSetCmd.PersistentFlags().StringVarP(&tmpWifiStrReg, "reg", "r", "US", "Sets the regulatory domain according to ISO/IEC 3166-1 alpha2")
	wifiSetCmd.PersistentFlags().BoolVarP(&tmpWifiDisabled, "disable", "d", false, "The flag disables the WiFi interface (omitting the flag enables the interface")
	wifiSetCmd.PersistentFlags().BoolVarP(&tmpWifiDisableNexmon, "nonexmon", "n", false, "Don't use the modified nexmon firmware")
	wifiSetCmd.PersistentFlags().StringVarP(&tmpWifiSSID, "ssid", "s", "", "The SSID to use for an Access Point or to join as station")
	wifiSetCmd.PersistentFlags().StringVarP(&tmpWifiPSK, "psk", "k", "", "The Pre-Shared-Key to use for the Access Point (if empty, an OPEN AP is created) or for the network")

	wifiSetAPCmd.Flags().Uint8VarP(&tmpWifiStrChannel, "channel", "c", 1, "The WiFi channel to use for the Access Point")
	wifiSetAPCmd.Flags().BoolVarP(&tmpWifiHideSSID, "hide", "x", false, "Hide the SSID of the Access Point")
}
