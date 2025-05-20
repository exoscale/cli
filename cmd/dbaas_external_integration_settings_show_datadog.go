package cmd

import (
	"fmt"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbaasExternalIntegrationSettingsShowDatadogOutput struct {
	DatadogDbmEnabled       bool `json:"datadog-dbm-enabled"`
	DatadogPgbouncerEnabled bool `json:"datadog-pgbouncer-enabled"`
}

func (o *dbaasExternalIntegrationSettingsShowDatadogOutput) ToJSON()  { output.JSON(o) }
func (o *dbaasExternalIntegrationSettingsShowDatadogOutput) ToText()  { output.Text(o) }
func (o *dbaasExternalIntegrationSettingsShowDatadogOutput) ToTable() { output.Table(o) }

func (c *dbaasExternalIntegrationSettingsShowCmd) showDatadog() (output.Outputter, error) {
	ctx := GContext

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return nil, err
	}

	integrationID, err := v3.ParseUUID(c.IntegrationID)
	if err != nil {
		return nil, fmt.Errorf("invalid integration ID: %w", err)
	}

	res, err := client.GetDBAASExternalIntegrationSettingsDatadog(ctx, integrationID)
	if err != nil {
		return nil, fmt.Errorf("error getting settings for integration: %w", err)
	}

	out := &dbaasExternalIntegrationSettingsShowDatadogOutput{
		DatadogDbmEnabled:       *res.Settings.DatadogDbmEnabled,
		DatadogPgbouncerEnabled: *res.Settings.DatadogPgbouncerEnabled,
	}
	return out, nil
}
