package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/egoscale"
)

var runstatusMaintenanceRemoveCmd = &cobra.Command{
	Use:     "remove [PAGE] MAINTENANCE-NAME",
	Short:   "Remove maintenance from a runstat.us page",
	Aliases: gRemoveAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		pageName := account.CurrentAccount.DefaultRunstatusPage
		maintenanceName := args[0]

		if account.CurrentAccount.DefaultRunstatusPage == "" && len(args) == 1 {
			fmt.Fprintf(os.Stderr, `Missing page argument.

  Please specify a page in parameter or
  Set the key "defaultRunstatusPage" into %q.
`, gConfigFilePath)
			return fmt.Errorf("missing default runstat.us page")
		}

		if len(args) > 1 {
			pageName = args[0]
			maintenanceName = args[1]
		}

		runstatusPage, err := csRunstatus.GetRunstatusPage(gContext, egoscale.RunstatusPage{Subdomain: pageName})
		if err != nil {
			return err
		}

		maintenance, err := getRunstatusMaintenanceByNameOrID(*runstatusPage, maintenanceName)
		if err != nil {
			return err
		}

		// TODO: add "--force" flag
		if !askQuestion(fmt.Sprintf("Remove maintenance %q (%d) from %q?", maintenance.Title, maintenance.ID, pageName)) {
			return nil
		}

		if err := csRunstatus.DeleteRunstatusMaintenance(gContext, *maintenance); err != nil {
			return fmt.Errorf("error removing %q:\n%v", maintenanceName, err)
		}
		fmt.Println(maintenance.ID)
		return nil
	},
}

func init() {
	runstatusMaintenanceCmd.AddCommand(runstatusMaintenanceRemoveCmd)
}
