package cmd

import (
	"bytes"
	"errors"
	"os"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

func init() {
	runstatusCmd.AddCommand(&cobra.Command{
		Use:     "show <page name>",
		Short:   "Show a runstat.us page details",
		Aliases: gShowAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("show expects one page name or id")
			}
			return showRunstatusPage(args[0])
		},
	})
}

func showRunstatusPage(name string) error {
	page, err := csRunstatus.GetRunstatusPage(gContext, egoscale.RunstatusPage{Subdomain: name})
	if err != nil {
		return err
	}

	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{page.Subdomain})

	t.Append([]string{"Name", page.Subdomain})
	t.Append([]string{"URL", page.PublicURL})
	t.Append([]string{"Email", page.SupportEmail})
	t.Append([]string{"State", page.State})

	if len(page.Services) > 0 {
		buf := bytes.NewBuffer(nil)
		st := table.NewEmbeddedTable(buf)
		st.SetHeader([]string{" "})
		for _, svc := range page.Services {
			st.Append([]string{svc.Name, svc.State})
		}
		st.Render()
		t.Append([]string{"Services", buf.String()})
	}

	if len(page.Incidents) > 0 {
		buf := bytes.NewBuffer(nil)
		it := table.NewEmbeddedTable(buf)
		for _, i := range page.Incidents {
			it.Append([]string{i.Title, i.State, i.Status, formatSchedule(i.StartDate, i.EndDate)})
		}
		it.Render()
		t.Append([]string{"Incidents", buf.String()})
	}

	if len(page.Maintenances) > 0 {
		buf := bytes.NewBuffer(nil)
		mt := table.NewEmbeddedTable(buf)
		for _, m := range page.Maintenances {
			mt.Append([]string{m.Title, m.Status, formatSchedule(m.StartDate, m.EndDate)})
		}
		mt.Render()
		t.Append([]string{"Maintenances", buf.String()})
	}

	t.Render()

	return nil
}
