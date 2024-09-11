package cmd

import (
	"fmt"
	"os"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
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
	default:
		return fmt.Errorf("unsupported external endpoint type %q", c.Type)
	}
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasExternalEndpointCmd, &dbaasExternalEndpointShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}

type externalEndpointShowOutput struct {
	ID       string                       `json:"id"`
	Name     string                       `json:"name"`
	Type     v3.EnumExternalEndpointTypes `json:"type"`
	// Settings any                          `json:"settings"`
}

func (o *externalEndpointShowOutput) ToJSON() { output.JSON(o) }

func (o *externalEndpointShowOutput) ToText() { output.Text(o) }

func (o *externalEndpointShowOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"DBaaS External Endpoint"})
	defer t.Render()

	t.Append([]string{"Name", o.Name})
	t.Append([]string{"ID", string(o.ID)})
	t.Append([]string{"Type", string(o.Type)})
}

