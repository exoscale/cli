package cmd

import (
	"fmt"
	"net/http"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
	"github.com/spf13/cobra"
)

func (c *dbaasServiceUpdateCmd) updatePG(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	databaseService := oapi.UpdateDbaasServicePgJSONRequestBody{}

	settingsSchema, err := cs.GetDbaasSettingsPgWithResponse(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve database Service settings: %w", err)
	}
	if settingsSchema.StatusCode() != http.StatusOK {
		return fmt.Errorf("API request error: unexpected status %s", settingsSchema.Status())
	}

	if c.PGBackupSchedule != "" {
		bh, bm, err := parseDatabaseBackupSchedule(c.PGBackupSchedule)
		if err != nil {
			return err
		}

		databaseService.BackupSchedule = &struct {
			BackupHour   *int64 `json:"backup-hour,omitempty"`
			BackupMinute *int64 `json:"backup-minute,omitempty"`
		}{
			BackupHour:   &bh,
			BackupMinute: &bm,
		}

		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.PGIPFilter)) {
		databaseService.IpFilter = &c.PGIPFilter
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Plan)) {
		databaseService.Plan = &c.Plan
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.TerminationProtection)) {
		databaseService.TerminationProtection = &c.TerminationProtection
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceDOW)) &&
		cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceTime)) {
		databaseService.Maintenance = &struct {
			Dow  oapi.UpdateDbaasServicePgJSONBodyMaintenanceDow `json:"dow"`
			Time string                                          `json:"time"`
		}{
			Dow:  oapi.UpdateDbaasServicePgJSONBodyMaintenanceDow(c.MaintenanceDOW),
			Time: c.MaintenanceTime,
		}
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.PGBouncerSettings)) {
		settings, err := validateDatabaseServiceSettings(
			c.PGBouncerSettings,
			settingsSchema.JSON200.Settings.Pgbouncer,
		)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		databaseService.PgbouncerSettings = &settings
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.PGLookoutSettings)) {
		settings, err := validateDatabaseServiceSettings(
			c.PGLookoutSettings,
			settingsSchema.JSON200.Settings.Pglookout,
		)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		databaseService.PglookoutSettings = &settings
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.PGSettings)) {
		settings, err := validateDatabaseServiceSettings(
			c.PGSettings,
			settingsSchema.JSON200.Settings.Pg,
		)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		databaseService.PgSettings = &settings
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.PGMigrationHost)) {
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
			Password: nonEmptyStringPtr(c.PGMigrationPassword),
			Username: nonEmptyStringPtr(c.PGMigrationUsername),
			Dbname:   nonEmptyStringPtr(c.PGMigrationDbName),
		}
		if c.PGMigrationSSL {
			databaseService.Migration.Ssl = &c.PGMigrationSSL
		}
		if c.PGMigrationMethod != "" {
			method := oapi.EnumMigrationMethod(c.PGMigrationMethod)
			databaseService.Migration.Method = &method
		}
		if len(c.MysqlMigrationIgnoreDbs) > 0 {
			dbsJoin := strings.Join(c.PGMigrationIgnoreDbs, ",")
			databaseService.Migration.IgnoreDbs = &dbsJoin
		}
		updated = true
	}

	if updated {
		var res *oapi.UpdateDbaasServicePgResponse
		decorateAsyncOperation(fmt.Sprintf("Updating Database Service %q...", c.Name), func() {
			res, err = cs.UpdateDbaasServicePgWithResponse(ctx, oapi.DbaasServiceName(c.Name), databaseService)
		})
		if err != nil {
			return err
		}
		if res.StatusCode() != http.StatusOK {
			return fmt.Errorf("API request error: unexpected status %s", res.Status())
		}
	}

	if !gQuiet {
		return c.outputFunc((&dbaasServiceShowCmd{
			Name: c.Name,
			Zone: c.Zone,
		}).showDatabaseServicePG(ctx))
	}

	return nil
}
