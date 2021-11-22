package cmd

import (
	"fmt"
	"net/http"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
	"github.com/spf13/cobra"
)

func (c *dbaasServiceCreateCmd) createMysql(_ *cobra.Command, _ []string) error {
	var err error

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	databaseService := oapi.CreateDbaasServiceMysqlJSONRequestBody{
		Plan:                  c.Plan,
		TerminationProtection: &c.TerminationProtection,
		Version:               nonEmptyStringPtr(c.MysqlVersion),
	}

	settingsSchema, err := cs.GetDbaasSettingsMysqlWithResponse(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve Database Service settings: %w", err)
	}
	if settingsSchema.StatusCode() != http.StatusOK {
		return fmt.Errorf("API request error: unexpected status %s", settingsSchema.Status())
	}

	if c.MysqlForkFrom != "" {
		databaseService.ForkFromService = (*oapi.DbaasServiceName)(&c.MysqlForkFrom)
		if c.MysqlRecoveryBackupTime != "" {
			databaseService.RecoveryBackupTime = &c.MysqlRecoveryBackupTime
		}
	}

	if c.MysqlAdminPassword != "" {
		databaseService.AdminPassword = &c.MysqlAdminPassword
	}

	if c.MysqlAdminUsername != "" {
		databaseService.AdminUsername = &c.MysqlAdminUsername
	}

	if c.MysqlBackupSchedule != "" {
		bh, bm, err := parseDatabaseBackupSchedule(c.MysqlBackupSchedule)
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

	if len(c.MysqlIPFilter) > 0 {
		databaseService.IpFilter = &c.MysqlIPFilter
	}

	if c.MaintenanceDOW != "" && c.MaintenanceTime != "" {
		databaseService.Maintenance.Dow = oapi.CreateDbaasServiceMysqlJSONBodyMaintenanceDow(c.MaintenanceDOW)
		databaseService.Maintenance.Time = c.MaintenanceTime
	}

	if c.MysqlSettings != "" {
		settings, err := validateDatabaseServiceSettings(c.MysqlSettings, settingsSchema.JSON200.Settings.Mysql)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		databaseService.MysqlSettings = &settings
	}

	fmt.Printf("Creating Database Service %q...\n", c.Name)

	res, err := cs.CreateDbaasServiceMysqlWithResponse(ctx, oapi.DbaasServiceName(c.Name), databaseService)
	if err != nil {
		return err
	}
	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("API request error: unexpected status %s", res.Status())
	}

	if !gQuiet {
		return c.outputFunc((&dbaasServiceShowCmd{
			Name: c.Name,
			Zone: c.Zone,
		}).showDatabaseServiceMysql(ctx))
	}

	return nil
}
