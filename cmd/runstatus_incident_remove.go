package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/egoscale"
)

var runstatusIncidentRemoveCmd = &cobra.Command{
	Use:     "remove [PAGE] INCIDENT-NAME|ID",
	Short:   "Remove incident from a runstat.us page",
	Aliases: gRemoveAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		pageName := account.CurrentAccount.DefaultRunstatusPage
		incidentName := args[0]

		if account.CurrentAccount.DefaultRunstatusPage == "" && len(args) == 1 {
			fmt.Fprintf(os.Stderr, `Error: No default runstat.us page is set:
  Please specify a page in parameter or add it to %q

  `, gConfigFilePath)
			return cmd.Usage()
		}

		if len(args) > 1 {
			pageName = args[0]
			incidentName = args[1]
		}

		runstatusPage, err := csRunstatus.GetRunstatusPage(gContext, egoscale.RunstatusPage{Subdomain: pageName})
		if err != nil {
			return err
		}

		incident, err := getRunstatusIncidentByNameOrID(*runstatusPage, incidentName)
		if err != nil {
			return err
		}

		// TODO: add "--force" flag
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete incident %q?", incidentName)) {
			return nil
		}
		if err := csRunstatus.DeleteRunstatusIncident(gContext, *incident); err != nil {
			return fmt.Errorf("error removing %q:\n%v", incidentName, err)
		}

		fmt.Printf("Incident %q successfully removed\n", incidentName)

		return nil
	},
}

func init() {
	runstatusIncidentCmd.AddCommand(runstatusIncidentRemoveCmd)
}
