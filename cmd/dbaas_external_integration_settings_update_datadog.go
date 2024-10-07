package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

func (c *dbaasExternalIntegrationSettingsUpdateCmd) updateDatadog(cmd *cobra.Command, _ []string) error {
	ctx := gContext

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
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

	decorateAsyncOperation(fmt.Sprintf("Updating DBaaS Datadog external integration settings %q", c.IntegrationID), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&dbaasExternalIntegrationSettingsShowCmd{
			cliCommandSettings: defaultCLICmdSettings(),
			IntegrationID:      string(integrationID),
			Type:               "datadog",
		}).cmdRun(nil, nil)
	}

	return nil
}
