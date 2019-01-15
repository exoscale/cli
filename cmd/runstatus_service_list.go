package cmd

import (
	"os"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// runstatusServiceListCmd represents the list command
var runstatusServiceListCmd = &cobra.Command{
	Use:     "list [page name]+",
	Short:   "List services",
	Aliases: gListAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		table := table.NewTable(os.Stdout)
		table.SetHeader([]string{"Page Name", "Name", "State"})

		if len(args) < 1 {
			pages, err := csRunstatus.ListRunstatusPages(gContext)
			if err != nil {
				return err
			}

			for _, page := range pages {
				services, err := csRunstatus.ListRunstatusServices(gContext, page)
				if err != nil {
					return err
				}

				for _, service := range services {
					table.Append([]string{page.Subdomain, service.Name, service.State})
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

			for _, service := range runstatusPage.Services {
				table.Append([]string{arg, service.Name, service.State})
				arg = ""
			}
		}

		table.Render()

		return nil
	},
}

func init() {
	runstatusServiceCmd.AddCommand(runstatusServiceListCmd)
}
