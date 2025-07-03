package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	v3 "github.com/exoscale/egoscale/v3"

	"github.com/exoscale/cli/cmd/internal/x"
	"github.com/exoscale/cli/pkg/account"
)

var xCmd *cobra.Command

func init() {
	xCmd = x.InitCommand()
	xCmd.Use = "x"
	xCmd.Aliases = append(xCmd.Aliases, "api")
	xCmd.Hidden = true
	xCmd.Long = `Low-level Exoscale API calls -- don't use this unless you have to.

These commands are automatically generated using openapi-cli-generator[1],
input parameters can be supplied either via stdin or using Shorthands[2].

[1]: https://github.com/exoscale/openapi-cli-generator
[2]: https://github.com/exoscale/openapi-cli-generator/tree/master/shorthand
`

	// This code being executed very early, at this point some information is not ready
	// to be used such as the account's configuration and some global variables, so we
	// we use the subcommand pre-run hook to perform last second changes to the
	// outgoing requests.
	xCmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		// If no value is provided for flag `--server`, infer the server URL from the current
		// CLI profile's default zone (or explicit `--zone` flag if specified) and environment
		// (or explicit `--environment` flag if speficied):
		if server, _ := cmd.Flags().GetString("server"); server == "" {
			zone := account.CurrentAccount.DefaultZone
			if z, _ := cmd.Flags().GetString("zone"); z != "" {
				zone = z
			}

			env := account.CurrentAccount.Environment
			if e, _ := cmd.Flags().GetString("environment"); e != "" {
				env = e
			}

			if err := cmd.Flags().Set("server", buildServerURL(zone, env)); err != nil {
				return err
			}
		}

		x.SetClientUserAgent(fmt.Sprintf(
			"Exoscale-CLI-X/%s (%s) %s",
			GVersion,
			GCommit,
			v3.UserAgent,
		))

		return x.SetClientCredentials(account.CurrentAccount.Key, account.CurrentAccount.Secret)
	}

	RootCmd.AddCommand(xCmd)
}

func buildServerURL(zone, env string) string {
	server := "https://api-ch-gva-2.exoscale.com/v2"

	if zone != "" && env != "" {
		server = fmt.Sprintf("https://%s-%s.exoscale.com/v2", env, zone)
	}

	return server
}
