package cmd

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
	"github.com/spf13/cobra"
)

func (c *dbaasServiceCreateCmd) createPG(_ *cobra.Command, _ []string) error {
	var err error

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	databaseService := oapi.CreateDbaasServicePgJSONRequestBody{
		Plan:                  c.Plan,
		TerminationProtection: &c.TerminationProtection,
		Version:               utils.NonEmptyStringPtr(c.PGVersion),
	}

	settingsSchema, err := globalstate.GlobalEgoscaleClient.GetDbaasSettingsPgWithResponse(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve Database Service settings: %w", err)
	}
	if settingsSchema.StatusCode() != http.StatusOK {
		return fmt.Errorf("API request error: unexpected status %s", settingsSchema.Status())
	}

	if c.PGForkFrom != "" {
		databaseService.ForkFromService = (*oapi.DbaasServiceName)(&c.PGForkFrom)
		if c.PGRecoveryBackupTime != "" {
			databaseService.RecoveryBackupTime = &c.PGRecoveryBackupTime
		}
	}

	if c.PGAdminPassword != "" {
		databaseService.AdminPassword = &c.PGAdminPassword
	}

	if c.PGAdminUsername != "" {
		databaseService.AdminUsername = &c.PGAdminUsername
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
	}

	if len(c.PGIPFilter) > 0 {
		databaseService.IpFilter = &c.PGIPFilter
	}

	if c.MaintenanceDOW != "" && c.MaintenanceTime != "" {
		databaseService.Maintenance = &struct {
			Dow  oapi.CreateDbaasServicePgJSONBodyMaintenanceDow `json:"dow"`
			Time string                                          `json:"time"`
		}{
			Dow:  oapi.CreateDbaasServicePgJSONBodyMaintenanceDow(c.MaintenanceDOW),
			Time: c.MaintenanceTime,
		}
	}

	if c.PGBouncerSettings != "" {
		settings, err := validateDatabaseServiceSettings(
			c.PGBouncerSettings,
			settingsSchema.JSON200.Settings.Pgbouncer,
		)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		databaseService.PgbouncerSettings = &settings
	}

	if c.PGLookoutSettings != "" {
		settings, err := validateDatabaseServiceSettings(
			c.PGLookoutSettings,
			settingsSchema.JSON200.Settings.Pglookout,
		)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		databaseService.PglookoutSettings = &settings
	}

	if c.PGSettings != "" {
		settings, err := validateDatabaseServiceSettings(
			c.PGSettings,
			settingsSchema.JSON200.Settings.Pg,
		)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		databaseService.PgSettings = &settings
	}

	if c.PGMigrationHost != "" {
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
			Dbname:   utils.NonEmptyStringPtr(c.PGMigrationDbName),
		}
		if c.PGMigrationSSL {
			databaseService.Migration.Ssl = &c.PGMigrationSSL
		}
		if c.PGMigrationMethod != "" {
			method := oapi.EnumMigrationMethod(c.PGMigrationMethod)
			databaseService.Migration.Method = &method
		}
		if len(c.PGMigrationIgnoreDbs) > 0 {
			dbsJoin := strings.Join(c.PGMigrationIgnoreDbs, ",")
			databaseService.Migration.IgnoreDbs = &dbsJoin
		}
	}

	var res *oapi.CreateDbaasServicePgResponse
	decorateAsyncOperation(fmt.Sprintf("Creating Database Service %q...", c.Name), func() {
		res, err = globalstate.GlobalEgoscaleClient.CreateDbaasServicePgWithResponse(ctx, oapi.DbaasServiceName(c.Name), databaseService)
	})
	if err != nil {
		return err
	}
	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("API request error: unexpected status %s", res.Status())
	}

	if !globalstate.Quiet {
		return c.outputFunc((&dbaasServiceShowCmd{
			Name: c.Name,
			Zone: c.Zone,
		}).showDatabaseServicePG(ctx))
	}

	return nil
}
