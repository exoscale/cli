package cmd

import (
	"os"
	"strconv"

	"github.com/exoscale/cli/table"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var runstatusIncidentListCmd = &cobra.Command{
	Use:     "list [page name]+",
	Short:   "List incidents from page(s)",
	Aliases: gListAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		pages, err := runstatusAllPages(args)
		if err != nil {
			return err
		}

		table := table.NewTable(os.Stdout)
		table.SetHeader([]string{"Page Name", "Title", "State", "Status", "When", "ID"})

		for _, page := range pages {
			incidents, err := csRunstatus.ListRunstatusIncidents(gContext, page)
			if err != nil {
				return err
			}

			for _, incident := range incidents {
				table.Append([]string{
					page.Subdomain,
					incident.Title,
					incident.State,
					incident.Status,
					formatSchedule(incident.StartDate, incident.EndDate),
					strconv.Itoa(incident.ID),
				})
				page.Subdomain = ""
			}
		}

		table.Render()

		return nil
	},
}

func init() {
	runstatusIncidentCmd.AddCommand(runstatusIncidentListCmd)
}
