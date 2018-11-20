package cli_client

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)



var (
	// deploy
	tmpDBBackupName = ""
)


func init() {
	cmdDB := &cobra.Command{
		Use:   "db",
		Short: "Database backup and restore",
	}

	cmdDBBackup := &cobra.Command{
		Use:   "backup",
		Short: "Backup DB",
		Run: func(cmd *cobra.Command, args []string) {
			if len(tmpDBBackupName) == 0 {
				fmt.Println("A name for the backup has to be provided with the '--name' flag")
				os.Exit(-1)
			}
			fmt.Print("Creating backup ...")
			err := ClientDBBackup(TIMEOUT_LONG, StrRemoteHost, StrRemotePort, tmpDBBackupName)
			if err != nil {
				fmt.Println(" failed")
				fmt.Println(err.Error())
				os.Exit(-1)
			}
			fmt.Println(" success")
		},

	}

	cmdDBRestore := &cobra.Command{
		Use:   "restore",
		Short: "Restore DB",
		Run: func(cmd *cobra.Command, args []string) {
			if len(tmpDBBackupName) == 0 {
				fmt.Println("A name for the backup has to be provided with the '--name' flag")
				os.Exit(-1)
			}
			fmt.Print("Restoring ...")
			err := ClientDBRestore(TIMEOUT_LONG, StrRemoteHost, StrRemotePort, tmpDBBackupName)
			if err != nil {
				fmt.Println(" failed")
				fmt.Println(err.Error())
				os.Exit(-1)
			}
			fmt.Println(" success")
		},

	}

	cmdDBList := &cobra.Command{
		Use:   "list",
		Short: "List backups",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			list,err := ClientDBList(TIMEOUT_LONG, StrRemoteHost, StrRemotePort)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(-1)
			}

			if BoolJson {
				b, err := json.Marshal(list)
				if err == nil {
					fmt.Println(string(b))
				}
			} else {
				fmt.Println("Database backups:")
				for _, item := range list {
					fmt.Println(item)
				}
			}
		},
	}



	rootCmd.AddCommand(cmdDB)
	cmdDB.AddCommand(cmdDBBackup, cmdDBList, cmdDBRestore)

	cmdDBBackup.Flags().StringVarP(&tmpDBBackupName, "name", "n", "","Name of backup")
	cmdDBRestore.Flags().StringVarP(&tmpDBBackupName, "name", "n", "","Name of backup")

	cmdDBList.Flags().BoolVar(&BoolJson, "json", false, "Output results as JSON if applicable")

}
