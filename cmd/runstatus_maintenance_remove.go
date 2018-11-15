package cmd

import (
	"fmt"
	"os"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var runstatusMaintenanceRemoveCmd = &cobra.Command{
	Use:     "remove [page name] <maintenance name>",
	Short:   "Remove maintenance from a runstat.us page",
	Aliases: gRemoveAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		pageName := gCurrentAccount.DefaultRunstatusPage
		maintenanceName := args[0]

		if gCurrentAccount.DefaultRunstatusPage == "" && len(args) == 1 {
			fmt.Fprintf(os.Stderr, `Error: No default runstat.us page is set:
  Please specify a page in parameter or add it to %q

  `, gConfigFilePath)
			return cmd.Usage()
		}

		if len(args) > 1 {
			pageName = args[0]
			maintenanceName = args[1]
		}

		runstatusPage, err := csRunstatus.GetRunstatusPage(gContext, egoscale.RunstatusPage{Subdomain: pageName})
		if err != nil {
			return err
		}

		maintenance, err := getMaintenanceByNameOrID(*runstatusPage, maintenanceName)
		if err != nil {
			return err
		}

		if !askQuestion(fmt.Sprintf("sure you want to remove %q maintenance", maintenanceName)) {
			return nil
		}

		if err := csRunstatus.DeleteRunstatusMaintenance(gContext, *maintenance); err != nil {
			return fmt.Errorf("error removing %q:\n%v", maintenanceName, err)
		}
		fmt.Printf("Maintenance %q successfully removed\n", maintenanceName)
		return nil
	},
}

func init() {
	runstatusMaintenanceCmd.AddCommand(runstatusMaintenanceRemoveCmd)
}
