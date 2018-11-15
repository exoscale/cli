package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// runstatusMaintenanceShowCmd represents the show command
var runstatusMaintenanceShowCmd = &cobra.Command{
	Use:     "show [page name] <maintenance name>",
	Short:   "Show a maintenance",
	Aliases: gShowAlias,
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

		maintenance, err := getMaintenanceByNameOrID(*runstatusPage, serviceName)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)
		fmt.Fprintf(w, "Title:\t%s\n", maintenance.Title)  // nolint: errcheck
		fmt.Fprintf(w, "State:\t%s\n", maintenance.Status) // nolint: errcheck
		fmt.Fprintf(w, "URL:\t%s\n", maintenance.URL)      // nolint: errcheck

		return w.Flush()
	},
}

func init() {
	runstatusMaintenanceCmd.AddCommand(runstatusMaintenanceShowCmd)
}
