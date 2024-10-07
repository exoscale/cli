package cmd
import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
)

type dbaasExternalIntegrationSettingsUpdateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	Type          string `cli-arg:"#"`
	IntegrationID string `cli-arg:"#"`

	HelpDatadog       bool `cli-usage:"show usage for flags specific to the datadog external integration type"`

	DatadogDbmEnabled bool `cli-flag:"datadog-dbm-enabled" cli-usage:"Enable/Disable pg stats with Datadog" cli-hidden:""`
	DatadogPgbouncerEnabled bool `cli-flag:"datadog-pgbouncer-enabled" cli-usage:"Enable/Disable pgbouncer stats with Datadog" cli-hidden:""`
}

func (c *dbaasExternalIntegrationSettingsUpdateCmd) cmdAliases() []string { return nil }
func (c *dbaasExternalIntegrationSettingsUpdateCmd) cmdShort() string { return "Update external integration settings"}
func (c *dbaasExternalIntegrationSettingsUpdateCmd) cmdLong() string { return "Update external integration settings"}
func (c *dbaasExternalIntegrationSettingsUpdateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	switch {
	case cmd.Flags().Changed("help-datadog"):
		cmdShowHelpFlags(cmd.Flags(), "datadog-")
		os.Exit(0)
	}

	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasExternalIntegrationSettingsUpdateCmd) cmdRun(cmd *cobra.Command, args []string) error {
	switch c.Type {
	case "datadog":
		return c.updateDatadog(cmd, args)
	default:
		return fmt.Errorf("unsupported updating external integration settings for type %q", c.Type)
	}
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasExternalIntegrationSettingsCmd, &dbaasExternalIntegrationSettingsUpdateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
