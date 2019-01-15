package cmd

import (
	"os"

	"github.com/exoscale/cli/table"
	"github.com/spf13/cobra"
)

// runstatusListCmd represents the list command
var runstatusListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List runstat.us pages",
	Aliases: gListAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		pages, err := csRunstatus.ListRunstatusPages(gContext)
		if err != nil {
			return err
		}

		table := table.NewTable(os.Stdout)
		table.SetHeader([]string{"Page Name", "Public URL"})
		for _, page := range pages {
			table.Append([]string{page.Subdomain, page.PublicURL})
		}
		table.Render()

		return nil
	},
}

func init() {
	runstatusCmd.AddCommand(runstatusListCmd)
}
