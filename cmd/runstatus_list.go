package cmd

import (
	"os"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// runstatusListCmd represents the list command
var runstatusListCmd = &cobra.Command{
	Use:     "list [page name]+",
	Short:   "List runstat.us pages",
	Aliases: gListAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		pages, err := runstatusAllPages(args)
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

func runstatusAllPages(args []string) ([]egoscale.RunstatusPage, error) {
	pages := []egoscale.RunstatusPage{}

	if len(args) == 0 {
		ps, err := csRunstatus.ListRunstatusPages(gContext)
		if err != nil {
			return nil, err
		}

		for i := range ps {
			pages = append(pages, ps[i])
		}
	}

	for _, arg := range args {
		page, err := csRunstatus.GetRunstatusPage(gContext, egoscale.RunstatusPage{Subdomain: arg})
		if err != nil {
			return nil, err
		}

		pages = append(pages, *page)
	}

	return pages, nil
}

func init() {
	runstatusCmd.AddCommand(runstatusListCmd)
}
