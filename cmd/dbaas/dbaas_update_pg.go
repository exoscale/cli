package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

func (c *dbaasServiceUpdateCmd) updatePG(cmd *cobra.Command, _ []string) error {
	var updated bool

<<<<<<< Updated upstream:cmd/dbaas_update_pg.go
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}
=======
	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_pg.go

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

<<<<<<< Updated upstream:cmd/dbaas_update_pg.go
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.PGIPFilter)) {
		databaseService.IPFilter = c.PGIPFilter
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Plan)) {
		databaseService.Plan = c.Plan
=======
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.PGIPFilter)) {
		databaseService.IpFilter = &c.PGIPFilter
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Plan)) {
		databaseService.Plan = &c.Plan
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_pg.go
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.TerminationProtection)) {
		databaseService.TerminationProtection = &c.TerminationProtection
		updated = true
	}

<<<<<<< Updated upstream:cmd/dbaas_update_pg.go
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceDOW)) &&
		cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceTime)) {
		databaseService.Maintenance = &v3.UpdateDBAASServicePGRequestMaintenance{
			Dow:  v3.UpdateDBAASServicePGRequestMaintenanceDow(c.MaintenanceDOW),
=======
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.MaintenanceDOW)) &&
		cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.MaintenanceTime)) {
		databaseService.Maintenance = &struct {
			Dow  oapi.UpdateDbaasServicePgJSONBodyMaintenanceDow `json:"dow"`
			Time string                                          `json:"time"`
		}{
			Dow:  oapi.UpdateDbaasServicePgJSONBodyMaintenanceDow(c.MaintenanceDOW),
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_pg.go
			Time: c.MaintenanceTime,
		}
		updated = true
	}

<<<<<<< Updated upstream:cmd/dbaas_update_pg.go
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.PGBouncerSettings)) {
		_, err := validateDatabaseServiceSettings(
=======
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.PGBouncerSettings)) {
		settings, err := validateDatabaseServiceSettings(
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_pg.go
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

<<<<<<< Updated upstream:cmd/dbaas_update_pg.go
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.PGLookoutSettings)) {
		_, err := validateDatabaseServiceSettings(
=======
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.PGLookoutSettings)) {
		settings, err := validateDatabaseServiceSettings(
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_pg.go
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

<<<<<<< Updated upstream:cmd/dbaas_update_pg.go
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.PGSettings)) {
		_, err := validateDatabaseServiceSettings(
=======
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.PGSettings)) {
		settings, err := validateDatabaseServiceSettings(
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_pg.go
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
		databaseService.PGSettings = *settings
		updated = true
	}

<<<<<<< Updated upstream:cmd/dbaas_update_pg.go
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.PGMigrationHost)) {
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
=======
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.PGMigrationHost)) {
		databaseService.Migration = &struct {
			Dbname    *string                   `json:"dbname,omitempty"`
			Host      string                    `json:"host"`
			IgnoreDbs *string                   `json:"ignore-dbs,omitempty"`
			Method    *oapi.EnumMigrationMethod `json:"method,omitempty"`
			Password  *string                   `json:"password,omitempty"`
			Port      int64                     `json:"port"`
			Ssl       *bool                     `json:"ssl,omitempty"`
			Username  *string                   `json:"username,omitempty"`
		}{
			Host:     c.PGMigrationHost,
			Port:     c.PGMigrationPort,
			Password: utils.NonEmptyStringPtr(c.PGMigrationPassword),
			Username: utils.NonEmptyStringPtr(c.PGMigrationUsername),
			Dbname:   utils.NonEmptyStringPtr(c.PGMigrationDBName),
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_pg.go
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
<<<<<<< Updated upstream:cmd/dbaas_update_pg.go
		op, err := client.UpdateDBAASServicePG(ctx, c.Name, databaseService)
=======
		var res *oapi.UpdateDbaasServicePgResponse
		utils.DecorateAsyncOperation(fmt.Sprintf("Updating Database Service %q...", c.Name), func() {
			res, err = globalstate.EgoscaleClient.UpdateDbaasServicePgWithResponse(ctx, oapi.DbaasServiceName(c.Name), databaseService)
		})
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_pg.go
		if err != nil {
			if errors.Is(err, v3.ErrNotFound) {
				return fmt.Errorf("resource not found in zone %q", c.Zone)
			}
			return err
		}

		decorateAsyncOperation(fmt.Sprintf("Updating Database Service %q...", c.Name), func() {
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
