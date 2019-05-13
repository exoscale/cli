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
		Short:   "Show runstat.us page details",
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

	t.Render()

	return nil
}
