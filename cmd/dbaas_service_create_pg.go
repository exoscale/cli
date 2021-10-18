package cmd

import (
	"fmt"
	"net/http"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
	"github.com/spf13/cobra"
)

func (c *dbServiceCreateCmd) createPG(_ *cobra.Command, _ []string) error {
	var err error

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	databaseService := oapi.CreateDbaasServicePgJSONRequestBody{
		Plan:                  c.Plan,
		TerminationProtection: &c.TerminationProtection,
	}

	settingsSchema, err := cs.GetDbaasSettingsPgWithResponse(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve Database Service settings: %w", err)
	}
	if settingsSchema.StatusCode() != http.StatusOK {
		return fmt.Errorf("API request error: unexpected status %s", settingsSchema.Status())
	}

	if c.ForkFrom != "" {
		databaseService.ForkFromService = (*oapi.DbaasServiceName)(&c.ForkFrom)
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
		databaseService.Maintenance.Dow = oapi.CreateDbaasServicePgJSONBodyMaintenanceDow(c.MaintenanceDOW)
		databaseService.Maintenance.Time = c.MaintenanceTime
	}

	if c.PGBouncerSettings != "" {
		settings, err := validateDatabaseServiceSettings(
			c.PGBouncerSettings,
			settingsSchema.JSON200.Settings.Pglookout,
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

	if c.PGVersion != "" {
		databaseService.Version = &c.PGVersion
	}

	fmt.Printf("Creating Database Service %q...\n", c.Name)

	res, err := cs.CreateDbaasServicePgWithResponse(ctx, oapi.DbaasServiceName(c.Name), databaseService)
	if err != nil {
		return err
	}
	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("API request error: unexpected status %s", res.Status())
	}

	if !gQuiet {
		return output((&dbServiceShowCmd{Zone: c.Zone, Name: c.Name}).showDatabaseServicePG(ctx))
	}

	return nil
}
