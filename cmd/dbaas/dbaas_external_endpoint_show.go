package dbaas

import (
	"fmt"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/spf13/cobra"
)

const (
	ExternalEndpointTypeDatadog       = "datadog"
	ExternalEndpointTypeOpensearch    = "opensearch"
	ExternalEndpointTypeElasticsearch = "elasticsearch"
	ExternalEndpointTypePrometheus    = "prometheus"
	ExternalEndpointTypeRsyslog       = "rsyslog"
)

type dbaasExternalEndpointShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Type       string `cli-arg:"#"`
	EndpointID string `cli-arg:"#"`
}

func (c *dbaasExternalEndpointShowCmd) CmdAliases() []string { return exocmd.GShowAlias }

func (c *dbaasExternalEndpointShowCmd) CmdShort() string {
	return "Show a Database External endpoint details"
}

func (c *dbaasExternalEndpointShowCmd) CmdLong() string {
	return "This command shows a Database Service external endpoint details."
}

func (c *dbaasExternalEndpointShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasExternalEndpointShowCmd) CmdRun(cmd *cobra.Command, args []string) error {
	switch c.Type {
	case ExternalEndpointTypeDatadog:
		return c.OutputFunc(c.showDatadog())
	case ExternalEndpointTypeOpensearch:
		return c.OutputFunc(c.showOpensearch())
	case ExternalEndpointTypeElasticsearch:
		return c.OutputFunc(c.showElasticsearch())
	case ExternalEndpointTypePrometheus:
		return c.OutputFunc(c.showPrometheus())
	case ExternalEndpointTypeRsyslog:
		return c.OutputFunc(c.showRsyslog())
	default:
		return fmt.Errorf("unsupported external endpoint type %q", c.Type)
	}
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasExternalEndpointCmd, &dbaasExternalEndpointShowCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
