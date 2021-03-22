package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var runstatusCreateCmd = &cobra.Command{
	Use:     "create NAME",
	Short:   "Create Runstat.us page",
	Aliases: gCreateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		dark, err := cmd.Flags().GetBool("dark")
		if err != nil {
			return err
		}

		for _, arg := range args {
			result, err := csRunstatus.CreateRunstatusPage(gContext, egoscale.RunstatusPage{
				Name:      arg,
				Subdomain: arg,
				DarkTheme: dark,
			})
			if err != nil {
				return err
			}

			if !gQuiet {
				fmt.Printf("Runstat.us page %q created:\n - %s\n", result.Subdomain, result.PublicURL)
			}
		}

		return nil
	},
}

func init() {
	runstatusCmd.AddCommand(runstatusCreateCmd)
	runstatusCreateCmd.Flags().BoolP("dark", "d", false, "Enable status page dark mode")
}
