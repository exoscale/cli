package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// runstatusServiceShowCmd represents the show command
var runstatusServiceShowCmd = &cobra.Command{
	Use:     "show [page name] <service name>",
	Short:   "Show a service",
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

		service, err := getServiceByName(*runstatusPage, serviceName)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)
		fmt.Fprintf(w, "Name:\t%s\n", service.Name)   // nolint: errcheck
		fmt.Fprintf(w, "State:\t%s\n", service.State) // nolint: errcheck
		fmt.Fprintf(w, "URL:\t%s\n", service.URL)     // nolint: errcheck

		return w.Flush()
	},
}

func init() {
	runstatusServiceCmd.AddCommand(runstatusServiceShowCmd)
}
