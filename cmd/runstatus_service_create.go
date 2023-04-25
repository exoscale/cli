package cmd

import (
	"fmt"
	"os"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/egoscale"

	"github.com/spf13/cobra"
)

var runstatusServiceCreateCmd = &cobra.Command{
	Use:     "create [PAGE] SERVICE-NAME",
	Short:   "Create a service",
	Aliases: gCreateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		pageName := gCurrentAccount.DefaultRunstatusPage
		serviceName := args[0]

		if gCurrentAccount.DefaultRunstatusPage == "" && len(args) == 1 {
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

		s, err := csRunstatus.CreateRunstatusService(gContext, egoscale.RunstatusService{
			PageURL: runstatusPage.URL,
			Name:    serviceName,
		})
		if err != nil {
			return err
		}

		if !globalstate.Quiet {
			fmt.Printf("Service %q successfully created\n", s.Name)
		}

		return nil
	},
}

func init() {
	runstatusServiceCmd.AddCommand(runstatusServiceCreateCmd)
}
