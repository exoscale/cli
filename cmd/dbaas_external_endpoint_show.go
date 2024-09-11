package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type dbaasExternalEndpointShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Type       string `cli-arg:"#"`
	EndpointID string `cli-arg:"#"`
}

func (c *dbaasExternalEndpointShowCmd) cmdAliases() []string { return gShowAlias }

func (c *dbaasExternalEndpointShowCmd) cmdShort() string {
	return "Show a Database External endpoint details"
}

func (c *dbaasExternalEndpointShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows a Database Service external endpoint details.`)
}

func (c *dbaasExternalEndpointShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasExternalEndpointShowCmd) cmdRun(cmd *cobra.Command, args []string) error {

	switch c.Type {
	case "datadog":
		return c.outputFunc(c.showDatadog())
	case "opensearch":
		return c.outputFunc(c.showOpensearch())
	default:
		return fmt.Errorf("unsupported external endpoint type %q", c.Type)
	}
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasExternalEndpointCmd, &dbaasExternalEndpointShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
