package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

func (c *dbaasServiceCreateCmd) createValkey(_ *cobra.Command, _ []string) error {
	var err error

	ctx := exocmd.GContext

	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))

	if err != nil {
		return fmt.Errorf("unable to create client: %w", err)
	}

	databaseService := v3.CreateDBAASServiceValkeyRequest{
		Plan:                  c.Plan,
		TerminationProtection: &c.TerminationProtection,
	}

	if c.ValkeyForkFrom != "" {
		databaseService.ForkFromService = v3.DBAASServiceName(c.ValkeyForkFrom)
		if c.ValkeyRecoveryBackupName != "" {
			databaseService.RecoveryBackupName = c.ValkeyRecoveryBackupName
		}
	}

	if len(c.ValkeyIPFilter) > 0 {
		databaseService.IPFilter = c.ValkeyIPFilter
	}

	if c.MaintenanceDOW != "" && c.MaintenanceTime != "" {
		databaseService.Maintenance = &v3.CreateDBAASServiceValkeyRequestMaintenance{
			Dow:  v3.CreateDBAASServiceValkeyRequestMaintenanceDow(c.MaintenanceDOW),
			Time: c.MaintenanceTime,
		}
	}

	if c.ValkeySettings != "" {

		settingsSchema, err := client.GetDBAASSettingsValkey(ctx)
		if err != nil {
			return fmt.Errorf("unable to retrieve Database Service settings: %w", err)
		}
		_, err = validateDatabaseServiceSettings(
			c.ValkeySettings,
			settingsSchema.Settings.Valkey,
		)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}

		settings := &v3.JSONSchemaValkey{}
		if err := json.Unmarshal([]byte(c.ValkeySettings), &settings); err != nil {
			return err
		}

		databaseService.ValkeySettings = settings
	}

	if c.ValkeyMigrationHost != "" {
		databaseService.Migration = &v3.CreateDBAASServiceValkeyRequestMigration{
			Host: c.ValkeyMigrationHost,
			Port: c.ValkeyMigrationPort,
		}
		if c.ValkeyMigrationPassword != "" {
			databaseService.Migration.Password = c.ValkeyMigrationPassword
		}
		if c.ValkeyMigrationUsername != "" {
			databaseService.Migration.Username = c.ValkeyMigrationUsername
		}
		if c.ValkeyMigrationDBName != "" {
			databaseService.Migration.Dbname = c.ValkeyMigrationDBName
		}
		if c.ValkeyMigrationSSL {
			databaseService.Migration.SSL = &c.ValkeyMigrationSSL
		}
		if c.ValkeyMigrationMethod != "" {
			databaseService.Migration.Method = v3.EnumMigrationMethod(c.ValkeyMigrationMethod)
		}
		if len(c.ValkeyMigrationIgnoreDbs) > 0 {
			databaseService.Migration.IgnoreDbs = strings.Join(c.ValkeyMigrationIgnoreDbs, ",")
		}
	}

	op, err := client.CreateDBAASServiceValkey(ctx, c.Name, databaseService)

	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Creating DBaaS Datadog external Endpoint %q", c.Name), func() {
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
		}).showDatabaseServiceValkey(ctx))
	}

	return nil
}
