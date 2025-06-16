package dbaas

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

func (c *dbaasServiceUpdateCmd) updatePG(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	databaseService := v3.UpdateDBAASServicePGRequest{}

	settingsSchema, err := client.GetDBAASSettingsPG(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve database Service settings: %w", err)
	}

	if c.PGBackupSchedule != "" {
		bh, bm, err := parseDatabaseBackupSchedule(c.PGBackupSchedule)
		if err != nil {
			return err
		}

		databaseService.BackupSchedule = &v3.UpdateDBAASServicePGRequestBackupSchedule{
			BackupHour:   bh,
			BackupMinute: bm,
		}

		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.PGIPFilter)) {
		databaseService.IPFilter = c.PGIPFilter
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Plan)) {
		databaseService.Plan = c.Plan
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.TerminationProtection)) {
		databaseService.TerminationProtection = &c.TerminationProtection
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.MaintenanceDOW)) &&
		cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.MaintenanceTime)) {
		databaseService.Maintenance = &v3.UpdateDBAASServicePGRequestMaintenance{
			Dow:  v3.UpdateDBAASServicePGRequestMaintenanceDow(c.MaintenanceDOW),
			Time: c.MaintenanceTime,
		}
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.PGBouncerSettings)) {
		_, err := validateDatabaseServiceSettings(
			c.PGBouncerSettings,
			settingsSchema.Settings.Pgbouncer,
		)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		settings := &v3.JSONSchemaPgbouncer{}
		if err = json.Unmarshal([]byte(c.PGBouncerSettings), settings); err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}

		databaseService.PgbouncerSettings = settings
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.PGLookoutSettings)) {
		_, err := validateDatabaseServiceSettings(
			c.PGLookoutSettings,
			settingsSchema.Settings.Pglookout,
		)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		settings := &v3.JSONSchemaPglookout{}
		if err = json.Unmarshal([]byte(c.PGLookoutSettings), settings); err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		databaseService.PglookoutSettings = settings
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.PGSettings)) {
		_, err := validateDatabaseServiceSettings(
			c.PGSettings,
			settingsSchema.Settings.PG,
		)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		settings := &v3.JSONSchemaPG{}
		if err = json.Unmarshal([]byte(c.PGSettings), settings); err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		databaseService.PGSettings = settings
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.PGMigrationHost)) {
		databaseService.Migration = &v3.UpdateDBAASServicePGRequestMigration{
			Host: c.PGMigrationHost,
			Port: c.PGMigrationPort,
		}
		if c.PGMigrationPassword != "" {
			databaseService.Migration.Password = c.PGMigrationPassword
		}
		if c.PGMigrationUsername != "" {
			databaseService.Migration.Username = c.PGMigrationUsername
		}
		if c.PGMigrationDBName != "" {
			databaseService.Migration.Dbname = c.PGMigrationDBName
		}
		if c.PGMigrationSSL {
			databaseService.Migration.SSL = &c.PGMigrationSSL
		}
		if c.PGMigrationMethod != "" {
			method := v3.EnumMigrationMethod(c.PGMigrationMethod)
			databaseService.Migration.Method = method
		}
		if len(c.MysqlMigrationIgnoreDbs) > 0 {
			dbsJoin := strings.Join(c.PGMigrationIgnoreDbs, ",")
			databaseService.Migration.IgnoreDbs = dbsJoin
		}
		updated = true
	}

	if updated {
		op, err := client.UpdateDBAASServicePG(ctx, c.Name, databaseService)
		if err != nil {
			if errors.Is(err, v3.ErrNotFound) {
				return fmt.Errorf("resource not found in zone %q", c.Zone)
			}
			return err
		}

		utils.DecorateAsyncOperation(fmt.Sprintf("Updating Database Service %q...", c.Name), func() {
			_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		})
		if err != nil {
			return err
		}
	}

	if !globalstate.Quiet {
		return c.OutputFunc((&dbaasServiceShowCmd{
			Name: c.Name,
			Zone: c.Zone,
		}).showDatabaseServicePG(ctx))
	}

	return nil
}
