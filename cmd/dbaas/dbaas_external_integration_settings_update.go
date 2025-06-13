package dbaas

import (
	"fmt"
	"os"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/spf13/cobra"
)

type dbaasExternalIntegrationSettingsUpdateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	Type          string `cli-arg:"#"`
	IntegrationID string `cli-arg:"#"`

	HelpDatadog bool `cli-usage:"show usage for flags specific to the datadog external integration type"`

	DatadogDbmEnabled       bool `cli-flag:"datadog-dbm-enabled" cli-usage:"Enable/Disable pg stats with Datadog"`
	DatadogPgbouncerEnabled bool `cli-flag:"datadog-pgbouncer-enabled" cli-usage:"Enable/Disable pgbouncer stats with Datadog"`
}

func (c *dbaasExternalIntegrationSettingsUpdateCmd) CmdAliases() []string { return nil }
func (c *dbaasExternalIntegrationSettingsUpdateCmd) CmdShort() string {
	return "Update external integration settings"
}
func (c *dbaasExternalIntegrationSettingsUpdateCmd) CmdLong() string {
	return "Update external integration settings"
}
func (c *dbaasExternalIntegrationSettingsUpdateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	switch {
	case cmd.Flags().Changed("help-datadog"):
		exocmd.CmdShowHelpFlags(cmd.Flags(), "datadog-")
		os.Exit(0)
	}

	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasExternalIntegrationSettingsUpdateCmd) CmdRun(cmd *cobra.Command, args []string) error {
	switch c.Type {
	case ExternalEndpointTypeDatadog:
		return c.updateDatadog(cmd, args)
	default:
		return fmt.Errorf("unsupported updating external integration settings for type %q", c.Type)
	}
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasExternalIntegrationSettingsCmd, &dbaasExternalIntegrationSettingsUpdateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
