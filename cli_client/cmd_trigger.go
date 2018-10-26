package cli_client

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"os"
)



const (
	templateFlagTriggerGroupName = "group-name"
	templateFlagTriggerGroupValue = "group-value"
)

var (
	// deploy
	tmpTriggerGroupName   = ""
	tmpTriggerGroupValue = int32(0)
)

func TriggerCheckFlags(cmd *cobra.Command) {
	valDefined,nameDefined := false,false
	cmd.Flags().Visit(func(flag *pflag.Flag) {
		if flag.Name == templateFlagTriggerGroupName {
			nameDefined = true
		}
		if flag.Name == templateFlagTriggerGroupValue {
			valDefined = true
		}
	})

	check := true
	if !nameDefined {
		fmt.Printf("The '%s' flag has to be set\n", templateFlagTriggerGroupName)
		check = false
	}
	if !valDefined {
		fmt.Printf("The '%s' flag has to be set\n", templateFlagTriggerGroupValue)
		check = false
	}

	if !check {
		os.Exit(-1)
	}
}

func init() {
	cmdTrigger := &cobra.Command{
		Use:   "trigger",
		Short: "Fire a group send action or wait for a group receive trigger",
	}

	cmdTriggerSend := &cobra.Command{
		Use:   "send",
		Short: "Fire a group send action",
		Run: func(cmd *cobra.Command, args []string) {
			TriggerCheckFlags(cmd)
			fmt.Printf("Sending value %d to group '%s'...", tmpTriggerGroupValue, tmpTriggerGroupName)
			err := ClientTriggerGroupSend(StrRemoteHost,StrRemotePort,tmpTriggerGroupName,tmpTriggerGroupValue)
			if err != nil {
				fmt.Println("error: ", err)
				os.Exit(-1)
			}
			fmt.Println("success")
		},

	}

	cmdTriggerWait := &cobra.Command{
		Use:   "wait",
		Short: "wait for a group receive trigger",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			TriggerCheckFlags(cmd)
			fmt.Printf("Waiting for value %d on group '%s'...", tmpTriggerGroupValue, tmpTriggerGroupName)
			err := ClientTriggerGroupWait(StrRemoteHost,StrRemotePort,tmpTriggerGroupName,tmpTriggerGroupValue)
			if err != nil {
				fmt.Println("error: ", err)
				os.Exit(-1)
			}
			fmt.Println("received")
		},
	}



	rootCmd.AddCommand(cmdTrigger)
	cmdTrigger.AddCommand(cmdTriggerSend, cmdTriggerWait)

	cmdTriggerSend.Flags().StringVarP(&tmpTriggerGroupName, templateFlagTriggerGroupName, "n", "","Name of the group to send to")
	cmdTriggerSend.Flags().Int32VarP(&tmpTriggerGroupValue, templateFlagTriggerGroupValue, "v", 0,"The value to send")

	cmdTriggerWait.Flags().StringVarP(&tmpTriggerGroupName, templateFlagTriggerGroupName, "n", "","Name of the group to listen")
	cmdTriggerWait.Flags().Int32VarP(&tmpTriggerGroupValue, templateFlagTriggerGroupValue, "v", 0,"The value to wait for")
}
