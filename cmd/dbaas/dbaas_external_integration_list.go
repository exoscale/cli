package dbaas

import (
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbaasExternalIntegrationListItemOutput struct {
	Description       string `json:"description"`
	DestEndpointName  string `json:"dest-endpoint-name"`
	DestEndpointID    string `json:"dest-endpoint-id"`
	IntegrationID     string `json:"id"`
	Status            string `json:"status"`
	SourceServiceName string `json:"source-service-name"`
	SourceServiceType string `json:"source-service-type"`
	Type              string `json:"type"`
}

type dbaasExternalIntegrationListOutput []dbaasExternalIntegrationListItemOutput

func (o *dbaasExternalIntegrationListOutput) ToJSON()  { output.JSON(o) }
func (o *dbaasExternalIntegrationListOutput) ToText()  { output.Text(o) }
func (o *dbaasExternalIntegrationListOutput) ToTable() { output.Table(o) }

type dbaasExternalIntegrationListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	ServiceName string `cli-arg:"#"`
}

func (c *dbaasExternalIntegrationListCmd) CmdAliases() []string { return exocmd.GListAlias }
func (c *dbaasExternalIntegrationListCmd) CmdShort() string     { return "List External Integrations" }
func (c *dbaasExternalIntegrationListCmd) CmdLong() string      { return "List External Integrations" }
func (c *dbaasExternalIntegrationListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasExternalIntegrationListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	res, err := client.ListDBAASExternalIntegrations(ctx, c.ServiceName)
	if err != nil {
		return fmt.Errorf("error listing integrations: %w", err)
	}

	out := make(dbaasExternalIntegrationListOutput, 0)

	for _, integration := range res.ExternalIntegrations {
		out = append(out, dbaasExternalIntegrationListItemOutput{
			IntegrationID:     string(integration.IntegrationID),
			Type:              string(integration.Type),
			Description:       integration.Description,
			DestEndpointName:  integration.DestEndpointName,
			DestEndpointID:    integration.DestEndpointID,
			Status:            integration.Status,
			SourceServiceName: integration.SourceServiceName,
			SourceServiceType: string(integration.SourceServiceType),
		})
	}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasExternalIntegrationCmd, &dbaasExternalIntegrationListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
