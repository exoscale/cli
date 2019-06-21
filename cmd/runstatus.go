package cmd

import (
	"fmt"
	"time"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// runstatusCmd represents the runstatus command
var runstatusCmd = &cobra.Command{
	Use:   "runstatus",
	Short: "Manage your Runstat.us pages",
	Long: `Focus on building your service,
knowing that when something does go wrong you can keep everyone informed using Runstatus.`,
}

func init() {
	RootCmd.AddCommand(runstatusCmd)
}

func getRunstatusPages(names []string) ([]egoscale.RunstatusPage, error) {
	if len(names) == 0 {
		return csRunstatus.ListRunstatusPages(gContext)
	}

	pages := []egoscale.RunstatusPage{}
	for _, name := range names {
		page, err := csRunstatus.GetRunstatusPage(gContext, egoscale.RunstatusPage{Subdomain: name})
		if err != nil {
			return nil, err
		}

		pages = append(pages, *page)
	}

	return pages, nil
}

func formatSchedule(start, end *time.Time) string {
	if start == nil || end == nil {
		return ""
	}
	return fmt.Sprintf("%s - %s", start, end.Sub(*start))
}
