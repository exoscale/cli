// AI-modified by hermes-agent - not reviewed yet
package dbaas

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

func (c *dbaasServiceCreateCmd) createClickhouse(_ *cobra.Command, _ []string) error {
	var err error

	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return fmt.Errorf("unable to create client: %w", err)
	}

	databaseService := v3.CreateDBAASServiceClickhouseRequest{
		Plan:                  c.Plan,
		TerminationProtection: &c.TerminationProtection,
		Version:               c.ClickhouseVersion,
	}

	if c.ClickhouseForkFrom != "" {
		databaseService.ForkFromService = v3.DBAASServiceName(c.ClickhouseForkFrom)
		if c.ClickhouseRecoveryBackupName != "" {
			databaseService.RecoveryBackupName = c.ClickhouseRecoveryBackupName
		}
	}

	if len(c.ClickhouseIPFilter) > 0 {
		databaseService.IPFilter = c.ClickhouseIPFilter
	}

	if c.MaintenanceDOW != "" && c.MaintenanceTime != "" {
		databaseService.Maintenance = &v3.CreateDBAASServiceClickhouseRequestMaintenance{
			Dow:  v3.CreateDBAASServiceClickhouseRequestMaintenanceDow(c.MaintenanceDOW),
			Time: c.MaintenanceTime,
		}
	}

	if c.ClickhouseSettings != "" {
		settingsSchema, err := client.GetDBAASSettingsClickhouse(ctx)
		if err != nil {
			return fmt.Errorf("unable to retrieve Database Service settings: %w", err)
		}
		_, err = validateDatabaseServiceSettings(
			c.ClickhouseSettings,
			settingsSchema.Settings.Clickhouse.Properties,
		)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}

		settings := &v3.JSONSchemaClickhouse{}
		if err := json.Unmarshal([]byte(c.ClickhouseSettings), &settings); err != nil {
			return err
		}

		databaseService.ClickhouseSettings = settings
	}

	op, err := client.CreateDBAASServiceClickhouse(ctx, c.Name, databaseService)
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Creating DBaaS ClickHouse service %q", c.Name), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	if err != nil {
		return err
	}

	serviceName := op.Reference.ID.String()

	if !globalstate.Quiet {
		return c.OutputFunc((&dbaasServiceShowCmd{
			Name: serviceName,
			Zone: c.Zone,
		}).showDatabaseServiceClickhouse(ctx))
	}

	return nil
}