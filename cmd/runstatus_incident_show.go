package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

func init() {
	runstatusIncidentCmd.AddCommand(&cobra.Command{
		Use:     "show [page name] <incident name | id>",
		Short:   "Show an incident details",
		Aliases: gShowAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return cmd.Usage()
			}

			page := gCurrentAccount.DefaultRunstatusPage
			incident := args[0]

			if gCurrentAccount.DefaultRunstatusPage == "" && len(args) == 1 {
				fmt.Fprintf(os.Stderr, `Error: No default runstat.us page is set:
  Please specify a page in parameter or add it to %q

  `, gConfigFilePath)
				return cmd.Usage()
			}

			if len(args) > 1 {
				page = args[0]
				incident = args[1]
			}

			return showRunstatusIncident(page, incident)
		},
	})
}

func showRunstatusIncident(p, i string) error {
	page, err := csRunstatus.GetRunstatusPage(gContext, egoscale.RunstatusPage{Subdomain: p})
	if err != nil {
		return err
	}

	incident, err := getIncidentByNameOrID(*page, i)
	if err != nil {
		return err
	}

	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{page.Subdomain})

	t.Append([]string{"ID", fmt.Sprint(incident.ID)})
	t.Append([]string{"Title", incident.Title})
	t.Append([]string{"State", incident.State})
	t.Append([]string{"Status", incident.Status})
	t.Append([]string{"Start Date", fmt.Sprint(incident.StartDate)})

	if incident.EndDate != nil {
		t.Append([]string{"End Date", fmt.Sprint(incident.EndDate)})
	}

	t.Append([]string{"Affected Services", strings.Join(incident.Services, "\n")})

	if len(incident.Events) > 0 {
		buf := bytes.NewBuffer(nil)
		et := table.NewEmbeddedTable(buf)
		et.SetHeader([]string{" "})
		for _, e := range incident.Events {
			et.Append([]string{fmt.Sprint(e.Created), e.Status, e.Text})
		}
		et.Render()
		t.Append([]string{"Event Stream", buf.String()})
	}

	t.Render()

	return nil
}
