package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

func init() {
	runstatusMaintenanceCmd.AddCommand(
		&cobra.Command{
			Use:     "show [page name] <maintenance name|id>",
			Short:   "Show a runstat.us page maintenance details",
			Aliases: gShowAlias,
			RunE: func(cmd *cobra.Command, args []string) error {
				if len(args) < 1 {
					return cmd.Usage()
				}

				page := gCurrentAccount.DefaultRunstatusPage
				maintenance := args[0]

				if gCurrentAccount.DefaultRunstatusPage == "" && len(args) == 1 {
					fmt.Fprintf(os.Stderr, `Error: No default runstat.us page is set:
  Please specify a page in parameter or add it to %q

  `, gConfigFilePath)
					return cmd.Usage()
				}

				if len(args) > 1 {
					page = args[0]
					maintenance = args[1]
				}

				return showRunstatusMaintenance(page, maintenance)
			},
		})
}

func showRunstatusMaintenance(p, m string) error {
	page, err := csRunstatus.GetRunstatusPage(gContext, egoscale.RunstatusPage{Subdomain: p})
	if err != nil {
		return err
	}

	maintenance, err := getMaintenanceByNameOrID(*page, m)
	if err != nil {
		return err
	}

	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{page.Subdomain})

	t.Append([]string{"ID", fmt.Sprint(maintenance.ID)})
	t.Append([]string{"Title", maintenance.Title})
	t.Append([]string{"Description", maintenance.Description})
	t.Append([]string{"State", maintenance.Status})
	t.Append([]string{"Start Date", fmt.Sprint(maintenance.StartDate)})
	t.Append([]string{"End Date", fmt.Sprint(maintenance.EndDate)})
	t.Append([]string{"Affected Services", strings.Join(maintenance.Services, "\n")})

	t.Render()

	return nil
}
