package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// runstatusIncidentShowCmd represents the show command
var runstatusIncidentShowCmd = &cobra.Command{
	Use:     "show [page name] <incident name | id>",
	Short:   "Show an incident detail",
	Aliases: gShowAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		pageName := gCurrentAccount.DefaultRunstatusPage
		incidentName := args[0]

		if gCurrentAccount.DefaultRunstatusPage == "" && len(args) == 1 {
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

		incident, err := getIncidentByNameOrID(*runstatusPage, incidentName)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)
		fmt.Fprintf(w, "Title:\t%s\n", incident.Title)            // nolint: errcheck
		fmt.Fprintf(w, "Description:\t%s\n", incident.StatusText) // nolint: errcheck
		fmt.Fprintf(w, "State:\t%s\n", incident.State)            // nolint: errcheck
		fmt.Fprintf(w, "Status:\t%s\n", incident.Status)          // nolint: errcheck

		return w.Flush()
	},
}

func init() {
	runstatusIncidentCmd.AddCommand(runstatusIncidentShowCmd)
}
