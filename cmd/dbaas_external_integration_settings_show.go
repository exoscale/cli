package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

type dbaasExternalIntegrationSettingsShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Type          string `cli-arg:"#"`
	IntegrationID string `cli-arg:"#"`
}

func (c *dbaasExternalIntegrationSettingsShowCmd) cmdAliases() []string { return gShowAlias }
func (c *dbaasExternalIntegrationSettingsShowCmd) cmdShort() string {
	return "Show External Integration Settings"
}
func (c *dbaasExternalIntegrationSettingsShowCmd) cmdLong() string {
	return "Show External Integration Settings"
}
func (c *dbaasExternalIntegrationSettingsShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasExternalIntegrationSettingsShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	switch c.Type {
	case "datadog":
		return c.outputFunc(c.showDatadog())
	default:
		return fmt.Errorf("unsupported external integration settings for type %q", c.Type)
	}
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasExternalIntegrationSettingsCmd, &dbaasExternalIntegrationSettingsShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
