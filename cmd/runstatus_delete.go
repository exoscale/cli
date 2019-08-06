package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var runstatusDeleteCmd = &cobra.Command{
	Use:     "delete <page name>+",
	Short:   "Delete runstat.us page(s)",
	Aliases: gDeleteAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}
		for _, arg := range args {
			// TODO: add "--force" flag
			if !askQuestion(fmt.Sprintf("sure you want to delete %q runstat.us page", arg)) {
				continue
			}

			runstatusPage, err := csRunstatus.GetRunstatusPage(gContext, egoscale.RunstatusPage{Subdomain: arg})
			if err != nil {
				return err
			}

			if err := csRunstatus.DeleteRunstatusPage(gContext, *runstatusPage); err != nil {
				return err
			}

			if !gQuiet {
				fmt.Printf("Page %q successfully deleted\n", arg)
			}
		}

		return nil
	},
}

func init() {
	runstatusCmd.AddCommand(runstatusDeleteCmd)
}
