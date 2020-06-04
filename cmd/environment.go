package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

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
`},
	)

	RootCmd.AddCommand(&cobra.Command{
		Use:    "env",
		Hidden: true,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("export EXOSCALE_API_KEY=%q\n", gCurrentAccount.Key)
			fmt.Printf("export EXOSCALE_API_SECRET=%q\n", gCurrentAccount.Secret)
			fmt.Printf("export EXOSCALE_API_ENDPOINT=%q\n", gCurrentAccount.Endpoint)
		},
	})
}
