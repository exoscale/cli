package cmd

import (
	"fmt"
	"os"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var runstatusServiceDeleteCmd = &cobra.Command{
	Use:     "delete [PAGE] SERVICE-NAME",
	Short:   "Delete a service",
	Aliases: gDeleteAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		pageName := account.CurrentAccount.DefaultRunstatusPage
		serviceName := args[0]

		if account.CurrentAccount.DefaultRunstatusPage == "" && len(args) == 1 {
			fmt.Fprintf(os.Stderr, `Error: No default runstat.us page is set:
  Please specify a page in parameter or add it to %q

  `, gConfigFilePath)
			return cmd.Usage()
		}

		if len(args) > 1 {
			pageName = args[0]
			serviceName = args[1]
		}

		runstatusPage, err := csRunstatus.GetRunstatusPage(gContext, egoscale.RunstatusPage{Subdomain: pageName})
		if err != nil {
			return err
		}

		service, err := getServiceByName(*runstatusPage, serviceName)
		if err != nil {
			return err
		}

		// TODO: add "--force" flag
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete service %q?", serviceName)) {
			return nil
		}

		if err := csRunstatus.DeleteRunstatusService(gContext, *service); err != nil {
			return err
		}
		fmt.Printf("Service %q successfully deleted\n", serviceName)

		return nil
	},
}

func init() {
	runstatusServiceCmd.AddCommand(runstatusServiceDeleteCmd)
}
