package cmd

import (
	"fmt"
	"os"

	"github.com/exoscale/egoscale"

	"github.com/spf13/cobra"
)

// runstatusServiceCreateCmd represents the create command
var runstatusServiceCreateCmd = &cobra.Command{
	Use:     "create [page name] <name>",
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

		err = csRunstatus.CreateRunstatusService(gContext, *runstatusPage, egoscale.RunstatusService{
			Name: serviceName,
		})
		if err != nil {
			return err
		}
		fmt.Printf("Service %q successfully created\n", serviceName)

		return nil
	},
}

func init() {
	runstatusServiceCmd.AddCommand(runstatusServiceCreateCmd)
}
