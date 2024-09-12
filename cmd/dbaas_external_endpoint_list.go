package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbaasExternalEndpointListItemOutput struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	Type string `json:"type"`
}

type dbaasExternalEndpointListOutput []dbaasExternalEndpointListItemOutput

func (o *dbaasExternalEndpointListOutput) ToJSON()  { output.JSON(o) }
func (o *dbaasExternalEndpointListOutput) ToText()  { output.Text(o) }
func (o *dbaasExternalEndpointListOutput) ToTable() { output.Table(o) }

type dbaasExternalEndpointListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`
}

func (c *dbaasExternalEndpointListCmd) cmdAliases() []string { return gListAlias }
func (c *dbaasExternalEndpointListCmd) cmdShort() string { return "List External Endpoints"}
func (c *dbaasExternalEndpointListCmd) cmdLong() string { return "List External Endpoints"}
func (c *dbaasExternalEndpointListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasExternalEndpointListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := gContext

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	res, err := client.ListDBAASExternalEndpoints(ctx)
	if err != nil {
		return fmt.Errorf("error listing endpoints: %w", err)
	}

	out := make(dbaasExternalEndpointListOutput, 0)

	for _, endpoint := range res.DBAASEndpoints {
		out = append(out, dbaasExternalEndpointListItemOutput{
			Name: *&endpoint.Name,
			ID:   string(*&endpoint.ID),
			Type: string(*&endpoint.Type),
		})
	}

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasExternalEndpointCmd, &dbaasExternalEndpointListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
