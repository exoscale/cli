package dbaas

import (
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

func (c *dbaasExternalIntegrationSettingsUpdateCmd) updateDatadog(cmd *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	integrationID, err := v3.ParseUUID(c.IntegrationID)
	if err != nil {
		return fmt.Errorf("invalid integration ID: %w", err)
	}

	payload := v3.UpdateDBAASExternalIntegrationSettingsDatadogRequest{
		Settings: &v3.DBAASIntegrationSettingsDatadog{},
	}

	if cmd.Flags().Changed("datadog-dbm-enabled") {
		payload.Settings.DatadogDbmEnabled = v3.Bool(c.DatadogDbmEnabled)
	}

	if cmd.Flags().Changed("datadog-pgbouncer-enabled") {
		payload.Settings.DatadogPgbouncerEnabled = v3.Bool(c.DatadogPgbouncerEnabled)
	}

	op, err := client.UpdateDBAASExternalIntegrationSettingsDatadog(ctx, integrationID, payload)
	if err != nil {
		return fmt.Errorf("error updating settings for integration: %w", err)
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Updating DBaaS Datadog external integration settings %q", c.IntegrationID), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&dbaasExternalIntegrationSettingsShowCmd{
			CliCommandSettings: exocmd.DefaultCLICmdSettings(),
			IntegrationID:      string(integrationID),
			Type:               "datadog",
		}).CmdRun(nil, nil)
	}

	return nil
}
