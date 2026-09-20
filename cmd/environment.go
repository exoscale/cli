package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
)

var envCmd = &cobra.Command{
	Use:    "env",
	Hidden: true,
	Run: func(cmd *cobra.Command, _ []string) {
		vars := map[string]string{
			"EXOSCALE_API_KEY":         account.CurrentAccount.Key,
			"EXOSCALE_API_SECRET":      account.CurrentAccount.Secret,
			"EXOSCALE_API_ENVIRONMENT": account.CurrentAccount.Environment,
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
		Long: "The exo CLI tool allows users to override some account configuration settings\n" +
			"by specifying shell environment variables. Here is the list of environment\n" +
			"variables supported:\n\n" +
			"  * `EXOSCALE_ACCOUNT`: account profile to use\n" +
			"  * `EXOSCALE_API_ENDPOINT`: Exoscale API endpoint\n" +
			"  * `EXOSCALE_API_ENVIRONMENT`: legacy API environment name\n" +
			"  * `EXOSCALE_API_KEY`: Exoscale API key\n" +
			"  * `EXOSCALE_API_SECRET`: Exoscale API secret\n" +
			"  * `EXOSCALE_API_TIMEOUT`: legacy API timeout in minutes\n" +
			"  * `EXOSCALE_CONFIG`: path to an alternate configuration file\n" +
			"  * `EXOSCALE_STORAGE_API_ENDPOINT`: SOS endpoint\n" +
			"  * `EXOSCALE_TIMEOUT`: per-zone timeout for list operations (e.g. 15s, 1m),\n" +
			"    or -1s to disable the timeout\n" +
			"  * `EXOSCALE_TRACE`: enable HTTP tracing (unset it to disable tracing). Trace\n" +
			"    output may contain sensitive information\n" +
			"  * `EXOSCALE_ZONE`: default zone\n\n" +
			"Command-line flags take precedence over environment variables, which take\n" +
			"precedence over the selected configuration profile. To override profile API\n" +
			"credentials, both `EXOSCALE_API_KEY` and `EXOSCALE_API_SECRET` must be set.\n\n" +
			"Compatibility aliases:\n" +
			"  * API key: `EXOSCALE_KEY`, `CLOUDSTACK_KEY`, `CLOUDSTACK_API_KEY`\n" +
			"  * API secret: `EXOSCALE_SECRET`, `EXOSCALE_SECRET_KEY`, `CLOUDSTACK_SECRET`,\n" +
			"    `CLOUDSTACK_SECRET_KEY`\n" +
			"  * SOS endpoint: `EXOSCALE_SOS_ENDPOINT`\n",
	},
	)

	envCmd.Flags().BoolP("unset", "u", false, "unset EXOSCALE_* environment variables")
	RootCmd.AddCommand(envCmd)
}
