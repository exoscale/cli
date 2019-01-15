package cmd

import (
	"os"
	"strconv"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var runstatusMaintenanceListCmd = &cobra.Command{
	Use:     "list [page name]+",
	Short:   "List maintenance from page(s)",
	Aliases: gListAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		table := table.NewTable(os.Stdout)
		table.SetHeader([]string{"Page Name", "Title", "Status", "When", "ID"})

		if len(args) < 1 {

			pages, err := csRunstatus.ListRunstatusPages(gContext)
			if err != nil {
				return err
			}

			for _, page := range pages {
				maintenances, err := csRunstatus.ListRunstatusMaintenances(gContext, page)
				if err != nil {
					return err
				}

				for _, maintenance := range maintenances {
					table.Append([]string{
						page.Subdomain,
						maintenance.Title,
						maintenance.Status,
						formatSchedule(maintenance.StartDate, maintenance.EndDate),
						strconv.Itoa(maintenance.ID),
					})
					page.Subdomain = ""
				}

			}
			table.Render()

			return nil
		}

		for _, arg := range args {
			page, err := csRunstatus.GetRunstatusPage(gContext, egoscale.RunstatusPage{Subdomain: arg})
			if err != nil {
				return err
			}

			maintenances, err := csRunstatus.ListRunstatusMaintenances(gContext, *page)
			if err != nil {
				return err
			}

			for _, maintenance := range maintenances {
				table.Append([]string{
					page.Subdomain,
					maintenance.Title,
					maintenance.Status,
					formatSchedule(maintenance.StartDate, maintenance.EndDate),
					strconv.Itoa(maintenance.ID),
				})
				page.Subdomain = ""
			}
		}
		table.Render()

		return nil
	},
}

func init() {
	runstatusMaintenanceCmd.AddCommand(runstatusMaintenanceListCmd)
}
