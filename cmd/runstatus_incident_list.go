package cmd

import (
	"os"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var runstatusIncidentListCmd = &cobra.Command{
	Use:     "list [page name]+",
	Short:   "List incidents from page(s)",
	Aliases: gListAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		table := table.NewTable(os.Stdout)
		table.SetHeader([]string{"Page Name", "Title", "State", "Status"})

		if len(args) < 1 {

			pages, err := csRunstatus.ListRunstatusPages(gContext)
			if err != nil {
				return err
			}

			for _, page := range pages {
				incidents, err := csRunstatus.ListRunstatusIncidents(gContext, page)
				if err != nil {
					return err
				}

				for _, incident := range incidents {
					table.Append([]string{page.Subdomain, incident.Title, incident.State, incident.Status})
					page.Subdomain = ""
				}

			}

			table.Render()
			return nil
		}

		for _, arg := range args {
			runstatusPage, err := csRunstatus.GetRunstatusPage(gContext, egoscale.RunstatusPage{Subdomain: arg})
			if err != nil {
				return err
			}

			incidents, err := csRunstatus.ListRunstatusIncidents(gContext, *runstatusPage)
			if err != nil {
				return err
			}
			for _, incident := range incidents {
				table.Append([]string{arg, incident.Title, incident.State, incident.Status})
				arg = ""
			}
		}

		table.Render()

		return nil
	},
}

func init() {
	runstatusIncidentCmd.AddCommand(runstatusIncidentListCmd)
}
