package cmd

import (
	"os"

	"github.com/exoscale/cli/table"
	"github.com/spf13/cobra"
)

// runstatusServiceListCmd represents the list command
var runstatusServiceListCmd = &cobra.Command{
	Use:     "list [page name]+",
	Short:   "List services",
	Aliases: gListAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		pages, err := runstatusAllPages(args)
		if err != nil {
			return err
		}

		table := table.NewTable(os.Stdout)
		table.SetHeader([]string{"Page Name", "Name", "State"})

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
	},
}

func init() {
	runstatusServiceCmd.AddCommand(runstatusServiceListCmd)
}
