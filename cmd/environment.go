package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:    "env",
	Hidden: true,
	Run: func(cmd *cobra.Command, _ []string) {
		vars := map[string]string{
			"EXOSCALE_API_KEY":         gCurrentAccount.Key,
			"EXOSCALE_API_SECRET":      gCurrentAccount.Secret,
			"EXOSCALE_API_ENDPOINT":    gCurrentAccount.Endpoint,
			"EXOSCALE_API_ENVIRONMENT": gCurrentAccount.Environment,
		}

		unset, _ := cmd.Flags().GetBool("unset")

		for k, v := range vars {
			if unset {
				fmt.Printf("unset %s\n", k)
			} else {
				fmt.Printf("export %s=%q\n", k, v)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(&cobra.Command{
		Use:   "environment",
		Short: "Environment variables usage",
		Long: `The exo CLI tool allows users to override some account configuration settings
by specifying shell environment variables. Here is the list of environment
variables supported:

  * EXOSCALE_API_KEY: the Exoscale client API key
  * EXOSCALE_API_SECRET: the Exoscale client API secret
  * EXOSCALE_API_ENDPOINT: the Exoscale (Compute) API endpoint to use

Note: to override the current profile API credentials, *both* EXOSCALE_API_KEY
and EXOSCALE_API_SECRET variables have to be set.
`,
	},
	)

	envCmd.Flags().BoolP("unset", "u", false, "unset EXOSCALE_* environment variables")
	RootCmd.AddCommand(envCmd)
}
