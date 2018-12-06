package cli_client

import (
	"fmt"
	pb "github.com/mame82/P4wnP1_aloa/proto"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"os"
)



const (
	templateFlagNameBluetooth = "bluetooth"
	templateFlagNameNetwork = "network"
	templateFlagNameTriggerActions = "trigger-actions"
	templateFlagNameWifi = "wifi"
	templateFlagNameUsb = "usb"
	templateFlagNameFullSettings = "full"
)
var templateFlagNames = map[string]bool{
	templateFlagNameBluetooth: true,
	templateFlagNameNetwork: true,
	templateFlagNameTriggerActions: true,
	templateFlagNameWifi: true,
	templateFlagNameUsb: true,
	templateFlagNameFullSettings: true,
}



var (
	// deploy
	tmpTemplateTypeFullSettings   = ""
	tmpTemplateTypeNetwork        = ""
	tmpTemplateTypeWifi           = ""
	tmpTemplateTypeUsb            = ""
	tmpTemplateTypeBluetooth      = ""
	tmpTemplateTypeTriggerActions = ""
//	tmpTemplateName               = ""
	//list
	tmpTemplateTypeFullSettingsToggle   = false
	tmpTemplateTypeNetworkToggle        = false
	tmpTemplateTypeWifiToggle           = false
	tmpTemplateTypeUsbToggle            = false
	tmpTemplateTypeBluetoothToggle      = false
	tmpTemplateTypeTriggerActionsToggle = false

)


func listTemplateType(ttype pb.ActionDeploySettingsTemplate_TemplateType) (err error) {
	fmt.Println("Templates of type", ttype, ":")
	fmt.Println("------------------------------------")
	list,err := ClientListTemplateType(TIMEOUT_MEDIUM, StrRemoteHost, StrRemotePort, ttype)
	if err != nil {
		fmt.Println("Error retrieving templates: ", err.Error())
		return err
	}
	for _,s := range list {
		fmt.Println(s)
	}
	fmt.Println()
	return nil
}

func deployTemplateType(ttype pb.ActionDeploySettingsTemplate_TemplateType, name string) (err error){
	fmt.Print("Deploying template of type ", ttype, ", name '", name, "': ...")
	err = ClientDeployTemplateType(TIMEOUT_MEDIUM, StrRemoteHost, StrRemotePort, ttype, name)
	if err != nil {
		fmt.Println("failed\n ", err.Error())
		return err
	}
	fmt.Println("success")
	return nil

}

func parseFlagsDeploy(cmd *cobra.Command) (res map[pb.ActionDeploySettingsTemplate_TemplateType]string) {
	res = make(map[pb.ActionDeploySettingsTemplate_TemplateType]string)
	cmd.Flags().Visit(func(flag *pflag.Flag) {
		if _,exists := templateFlagNames[flag.Name]; exists {
			switch flag.Name {
			case templateFlagNameBluetooth:
				res[pb.ActionDeploySettingsTemplate_BLUETOOTH] = flag.Value.String()
			case templateFlagNameFullSettings:
				res[pb.ActionDeploySettingsTemplate_FULL_SETTINGS] = flag.Value.String()
			case templateFlagNameUsb:
				res[pb.ActionDeploySettingsTemplate_USB] = flag.Value.String()
			case templateFlagNameWifi:
				res[pb.ActionDeploySettingsTemplate_WIFI] = flag.Value.String()
			case templateFlagNameTriggerActions:
				res[pb.ActionDeploySettingsTemplate_TRIGGER_ACTIONS] = flag.Value.String()
			case templateFlagNameNetwork:
				res[pb.ActionDeploySettingsTemplate_NETWORK] = flag.Value.String()
			}

		}
	})
	//fmt.Printf("%+v\n", res)
	return res
}

func parseFlagsList(cmd *cobra.Command) (res []pb.ActionDeploySettingsTemplate_TemplateType) {
	cmd.Flags().Visit(func(flag *pflag.Flag) {
		if _,exists := templateFlagNames[flag.Name]; exists {
			switch flag.Name {
			case templateFlagNameBluetooth:
				res = append(res, pb.ActionDeploySettingsTemplate_BLUETOOTH)
			case templateFlagNameFullSettings:
				res = append(res, pb.ActionDeploySettingsTemplate_FULL_SETTINGS)
			case templateFlagNameUsb:
				res = append(res, pb.ActionDeploySettingsTemplate_USB)
			case templateFlagNameWifi:
				res = append(res, pb.ActionDeploySettingsTemplate_WIFI)
			case templateFlagNameTriggerActions:
				res = append(res, pb.ActionDeploySettingsTemplate_TRIGGER_ACTIONS)
			case templateFlagNameNetwork:
				res = append(res, pb.ActionDeploySettingsTemplate_NETWORK)
			}
		}
	})

	if len(res) == 0 {
		res = append(res,
			pb.ActionDeploySettingsTemplate_FULL_SETTINGS,
			pb.ActionDeploySettingsTemplate_BLUETOOTH,
			pb.ActionDeploySettingsTemplate_USB,
			pb.ActionDeploySettingsTemplate_WIFI,
			pb.ActionDeploySettingsTemplate_TRIGGER_ACTIONS,
			pb.ActionDeploySettingsTemplate_NETWORK,
			)
	}

	//fmt.Printf("%+v\n", res)
	return res
}

func init() {
	cmdTemplate := &cobra.Command{
		Use:   "template",
		Short: "Deploy and list templates",
	}

	cmdTemplateDeploy := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy given gadget settings",
		Run: func(cmd *cobra.Command, args []string) {
			deployList := parseFlagsDeploy(cmd)
			for ttype,name := range deployList {
				err := deployTemplateType(ttype,name)
				if err != nil {
					os.Exit(-1)
				}
			}
		},

	}

	cmdTemplateList := &cobra.Command{
		Use:   "list",
		Short: "List templates",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			showList := parseFlagsList(cmd)
			for _,ttype := range showList {
				err := listTemplateType(ttype)
				if err != nil {
					os.Exit(-1)
				}
			}

		},
	}



	rootCmd.AddCommand(cmdTemplate)
	cmdTemplate.AddCommand(cmdTemplateDeploy, cmdTemplateList)

	cmdTemplateList.Flags().BoolVarP(&tmpTemplateTypeBluetoothToggle, templateFlagNameBluetooth, "b", false,"List existing bluetooth settings templates")
	cmdTemplateList.Flags().BoolVarP(&tmpTemplateTypeNetworkToggle, templateFlagNameNetwork, "n", false,"List existing network settings templates")
	cmdTemplateList.Flags().BoolVarP(&tmpTemplateTypeTriggerActionsToggle, templateFlagNameTriggerActions, "t", false,"List existing trigger action templates")
	cmdTemplateList.Flags().BoolVarP(&tmpTemplateTypeWifiToggle, templateFlagNameWifi, "w", false,"List existing WiFi settings templates")
	cmdTemplateList.Flags().BoolVarP(&tmpTemplateTypeUsbToggle, templateFlagNameUsb, "u", false,"List existing USB settings templates")
	cmdTemplateList.Flags().BoolVarP(&tmpTemplateTypeFullSettingsToggle, templateFlagNameFullSettings, "f", false,"List existing full settings templates")

	cmdTemplateDeploy.Flags().StringVarP(&tmpTemplateTypeBluetooth, templateFlagNameBluetooth, "b", "","Deploy Bluetooth template")
	cmdTemplateDeploy.Flags().StringVarP(&tmpTemplateTypeNetwork, templateFlagNameNetwork, "n", "","Deploy network settings template")
	cmdTemplateDeploy.Flags().StringVarP(&tmpTemplateTypeTriggerActions, templateFlagNameTriggerActions, "t", "","Deploy trigger action template")
	cmdTemplateDeploy.Flags().StringVarP(&tmpTemplateTypeWifi, templateFlagNameWifi, "w", "","Deploy WiFi settings templates")
	cmdTemplateDeploy.Flags().StringVarP(&tmpTemplateTypeUsb, templateFlagNameUsb, "u", "","Deploy USB settings template")
	cmdTemplateDeploy.Flags().StringVarP(&tmpTemplateTypeFullSettings, templateFlagNameFullSettings, "f", "","Deploy full settings template")

}
