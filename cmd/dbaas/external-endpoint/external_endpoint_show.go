package externalendpoint

import (
	"fmt"

	"github.com/exoscale/cli/cmd/dbaas/internal"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/spf13/cobra"
)

type dbaasExternalEndpointShowCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Type       string `cli-arg:"#"`
	EndpointID string `cli-arg:"#"`
}

func (c *dbaasExternalEndpointShowCmd) CmdAliases() []string { return clicmd.GShowAlias }

func (c *dbaasExternalEndpointShowCmd) CmdShort() string {
	return "Show a Database External endpoint details"
}

func (c *dbaasExternalEndpointShowCmd) CmdLong() string {
	return "This command shows a Database Service external endpoint details."
}

func (c *dbaasExternalEndpointShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasExternalEndpointShowCmd) CmdRun(cmd *cobra.Command, args []string) error {
	switch c.Type {
	case internal.ExternalEndpointTypeDatadog:
		return c.OutputFunc(c.showDatadog())
	case internal.ExternalEndpointTypeOpensearch:
		return c.OutputFunc(c.showOpensearch())
	case internal.ExternalEndpointTypeElasticsearch:
		return c.OutputFunc(c.showElasticsearch())
	case internal.ExternalEndpointTypePrometheus:
		return c.OutputFunc(c.showPrometheus())
	case internal.ExternalEndpointTypeRsyslog:
		return c.OutputFunc(c.showRsyslog())
	default:
		return fmt.Errorf("unsupported external endpoint type %q", c.Type)
	}
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &dbaasExternalEndpointShowCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
