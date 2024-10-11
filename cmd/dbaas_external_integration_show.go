package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbaasExternalIntegrationShowOutput struct {
	Description       string `json:"description"`
	DestEndpointName  string `json:"dest-endpoint-name"`
	DestEndpointID    string `json:"dest-endpoint-id"`
	IntegrationID     string `json:"id"`
	Status            string `json:"status"`
	SourceServiceName string `json:"source-service-name"`
	SourceServiceType string `json:"source-service-type"`
	Type              string `json:"type"`
}

func (o *dbaasExternalIntegrationShowOutput) ToJSON()  { output.JSON(o) }
func (o *dbaasExternalIntegrationShowOutput) ToText()  { output.Text(o) }
func (o *dbaasExternalIntegrationShowOutput) ToTable() { output.Table(o) }

type dbaasExternalIntegrationShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	IntegrationID string `cli-arg:"#"`
}

func (c *dbaasExternalIntegrationShowCmd) showExternalIntegration() (output.Outputter, error) {
	ctx := gContext

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return nil, err
	}

	integrationID, err := v3.ParseUUID(c.IntegrationID)
	if err != nil {
		return nil, fmt.Errorf("invalid integration ID: %w", err)
	}

	res, err := client.GetDBAASExternalIntegration(ctx, integrationID)
	if err != nil {
		return nil, fmt.Errorf("error showing integration: %w", err)
	}

	out := &dbaasExternalIntegrationShowOutput{
		IntegrationID:     string(res.IntegrationID),
		Type:              string(res.Type),
		Description:       res.Description,
		DestEndpointName:  res.DestEndpointName,
		DestEndpointID:    res.DestEndpointID,
		Status:            res.Status,
		SourceServiceName: res.SourceServiceName,
		SourceServiceType: string(res.SourceServiceType),
	}
	return out, nil
}

func (c *dbaasExternalIntegrationShowCmd) cmdAliases() []string { return gShowAlias }
func (c *dbaasExternalIntegrationShowCmd) cmdShort() string     { return "Show External Integration" }
func (c *dbaasExternalIntegrationShowCmd) cmdLong() string      { return "Show External Integration" }
func (c *dbaasExternalIntegrationShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasExternalIntegrationShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	return c.outputFunc(c.showExternalIntegration())
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasExternalIntegrationCmd, &dbaasExternalIntegrationShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
