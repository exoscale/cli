package dbaas

import (
	"fmt"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/spf13/cobra"
)

type dbaasExternalIntegrationSettingsShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Type          string `cli-arg:"#"`
	IntegrationID string `cli-arg:"#"`
}

func (c *dbaasExternalIntegrationSettingsShowCmd) CmdAliases() []string { return exocmd.GShowAlias }
func (c *dbaasExternalIntegrationSettingsShowCmd) CmdShort() string {
	return "Show External Integration Settings"
}
func (c *dbaasExternalIntegrationSettingsShowCmd) CmdLong() string {
	return "Show External Integration Settings"
}
func (c *dbaasExternalIntegrationSettingsShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasExternalIntegrationSettingsShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	switch c.Type {
	case ExternalEndpointTypeDatadog:
		return c.OutputFunc(c.showDatadog())
	default:
		return fmt.Errorf("unsupported external integration settings for type %q", c.Type)
	}
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasExternalIntegrationSettingsCmd, &dbaasExternalIntegrationSettingsShowCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
