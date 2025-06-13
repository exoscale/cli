package dbaas

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

func (c *dbaasServiceCreateCmd) createPG(_ *cobra.Command, _ []string) error {
	var err error

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	databaseService := v3.CreateDBAASServicePGRequest{
		Plan:                  c.Plan,
		TerminationProtection: &c.TerminationProtection,
	}
	if c.PGVersion != "" {
		databaseService.Version = v3.DBAASPGTargetVersions(c.PGVersion)
	}

	settingsSchema, err := client.GetDBAASSettingsPG(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve Database Service settings: %w", err)
	}

	if c.PGForkFrom != "" {
		databaseService.ForkFromService = v3.DBAASServiceName(c.PGForkFrom)
		if c.PGRecoveryBackupTime != "" {
			databaseService.RecoveryBackupTime = c.PGRecoveryBackupTime
		}
	}

	if c.PGAdminPassword != "" {
		databaseService.AdminPassword = c.PGAdminPassword
	}

	if c.PGAdminUsername != "" {
		databaseService.AdminUsername = c.PGAdminUsername
	}

	if c.PGBackupSchedule != "" {
		bh, bm, err := parseDatabaseBackupSchedule(c.PGBackupSchedule)
		if err != nil {
			return err
		}

		databaseService.BackupSchedule = &v3.CreateDBAASServicePGRequestBackupSchedule{
			BackupHour:   bh,
			BackupMinute: bm,
		}
	}

	if len(c.PGIPFilter) > 0 {
		databaseService.IPFilter = c.PGIPFilter
	}

	if c.MaintenanceDOW != "" && c.MaintenanceTime != "" {
		databaseService.Maintenance = &v3.CreateDBAASServicePGRequestMaintenance{
			Dow:  v3.CreateDBAASServicePGRequestMaintenanceDow(c.MaintenanceDOW),
			Time: c.MaintenanceTime,
		}
	}

	if c.PGBouncerSettings != "" {
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
	}

	if c.PGLookoutSettings != "" {
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
	}

	if c.PGSettings != "" {
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
	}

	if c.PGMigrationHost != "" {
		databaseService.Migration = &v3.CreateDBAASServicePGRequestMigration{
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
			method := c.PGMigrationMethod
			databaseService.Migration.Method = v3.EnumMigrationMethod(method)
		}
		if len(c.PGMigrationIgnoreDbs) > 0 {
			databaseService.Migration.IgnoreDbs = strings.Join(c.PGMigrationIgnoreDbs, ",")

		}
	}

	op, err := client.CreateDBAASServicePG(ctx, c.Name, databaseService)
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Creating Database Service %q...", c.Name), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return c.OutputFunc((&dbaasServiceShowCmd{
			Name: c.Name,
			Zone: c.Zone,
		}).showDatabaseServicePG(ctx))
	}

	return nil
}
