package cmd

import (
	"fmt"
	"os"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

func init() {
	runstatusServiceCmd.AddCommand(&cobra.Command{
		Use:     "show [page name] <service name>",
		Short:   "Show a service",
		Aliases: gShowAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return cmd.Usage()
			}

			page := gCurrentAccount.DefaultRunstatusPage
			service := args[0]

			if gCurrentAccount.DefaultRunstatusPage == "" && len(args) == 1 {
				fmt.Fprintf(os.Stderr, `Error: No default runstat.us page is set:
  Please specify a page in parameter or add it to %q

  `, gConfigFilePath)
				return cmd.Usage()
			}

			if len(args) > 1 {
				page = args[0]
				service = args[1]
			}

			return showRunstatusService(page, service)
		},
	})
}

func showRunstatusService(p, s string) error {
	page, err := csRunstatus.GetRunstatusPage(gContext, egoscale.RunstatusPage{Subdomain: p})
	if err != nil {
		return err
	}

	service, err := getServiceByName(*page, s)
	if err != nil {
		return err
	}

	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{page.Subdomain})

	t.Append([]string{"ID", fmt.Sprint(service.ID)})
	t.Append([]string{"Name", service.Name})
	t.Append([]string{"State", service.State})

	t.Render()

	return nil
}
