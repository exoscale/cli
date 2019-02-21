package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/exoscale/cli/table"
	"github.com/spf13/cobra"
)

// runstatusShowCmd represents the list command
var runstatusShowCmd = &cobra.Command{
	Use:     "show [page name]+",
	Short:   "Show runstat.us page details",
	Aliases: gShowAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		pages, err := runstatusAllPages(args)
		if err != nil {
			return err
		}

		for _, page := range pages {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)

			fmt.Fprintf(w, "Page:\t%s\n", page.Subdomain)     // nolint: errcheck
			fmt.Fprintf(w, "URL:\t%s\n", page.PublicURL)      // nolint: errcheck
			fmt.Fprintf(w, "Email:\t%s\n", page.SupportEmail) // nolint: errcheck
			fmt.Fprintf(w, "State:\t%s\n", page.State)        // nolint: errcheck

			if len(page.Services) > 0 {
				fmt.Fprintf(w, "Services:\n") // nolint: errcheck
				w.Flush()                     // nolint: errcheck

				table := table.NewTable(os.Stdout)
				table.SetHeader([]string{"Name", "State", "ID"})

				for _, svc := range page.Services {
					table.Append([]string{
						svc.Name,
						svc.State,
						strconv.Itoa(svc.ID),
					})
				}
				table.Render()

				fmt.Fprintln(w, "") // nolint: errcheck
			} else {
				fmt.Fprintf(w, "Services:\tn/a\n") // nolint: errcheck
			}

			if len(page.Incidents) > 0 {
				fmt.Fprintf(w, "Incidents:\n") // nolint: errcheck
				w.Flush()                      // nolint: errcheck

				table := table.NewTable(os.Stdout)
				table.SetHeader([]string{"Title", "State", "Status", "When", "ID"})

				for _, incident := range page.Incidents {
					table.Append([]string{
						incident.Title,
						incident.State,
						incident.Status,
						formatSchedule(incident.StartDate, incident.EndDate),
						strconv.Itoa(incident.ID),
					})
				}
				table.Render()

				fmt.Fprintln(w, "") // nolint: errcheck
			} else {
				fmt.Fprintf(w, "Incidents:\tn/a\n") // nolint: errcheck
			}

			if len(page.Maintenances) > 0 {
				fmt.Fprintf(w, "Maintenances:\n") // nolint: errcheck
				w.Flush()                         // nolint: errcheck

				table := table.NewTable(os.Stdout)
				table.SetHeader([]string{"Title", "Status", "When", "ID"})

				for _, maintenance := range page.Maintenances {
					table.Append([]string{
						maintenance.Title,
						maintenance.Status,
						formatSchedule(maintenance.StartDate, maintenance.EndDate),
						strconv.Itoa(maintenance.ID),
					})
				}
				table.Render()

				fmt.Fprintln(w, "") // nolint: errcheck
			} else {
				fmt.Fprintf(w, "Maintenances:\tn/a\n") // nolint: errcheck
			}
		}

		return nil
	},
}

func init() {
	runstatusCmd.AddCommand(runstatusShowCmd)
}
