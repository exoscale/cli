package cmd

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/exoscale/cli/utils"
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
		Version:               utils.NonEmptyStringPtr(c.MysqlVersion),
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
		databaseService.Maintenance = &struct {
			Dow  oapi.CreateDbaasServiceMysqlJSONBodyMaintenanceDow `json:"dow"`
			Time string                                             `json:"time"`
		}{
			Dow:  oapi.CreateDbaasServiceMysqlJSONBodyMaintenanceDow(c.MaintenanceDOW),
			Time: c.MaintenanceTime,
		}
	}

	if c.MysqlSettings != "" {
		settings, err := validateDatabaseServiceSettings(c.MysqlSettings, settingsSchema.JSON200.Settings.Mysql)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		databaseService.MysqlSettings = &settings
	}

	if c.MysqlMigrationHost != "" {
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
			Host:     c.MysqlMigrationHost,
			Port:     c.MysqlMigrationPort,
			Password: utils.NonEmptyStringPtr(c.MysqlMigrationPassword),
			Username: utils.NonEmptyStringPtr(c.MysqlMigrationUsername),
			Dbname:   utils.NonEmptyStringPtr(c.MysqlMigrationDbName),
		}
		if c.MysqlMigrationSSL {
			databaseService.Migration.Ssl = &c.MysqlMigrationSSL
		}
		if c.MysqlMigrationMethod != "" {
			method := oapi.EnumMigrationMethod(c.MysqlMigrationMethod)
			databaseService.Migration.Method = &method
		}
		if len(c.MysqlMigrationIgnoreDbs) > 0 {
			dbsJoin := strings.Join(c.MysqlMigrationIgnoreDbs, ",")
			databaseService.Migration.IgnoreDbs = &dbsJoin
		}
	}

	if c.MysqlBinlogRetentionPeriod > 0 {
		databaseService.BinlogRetentionPeriod = &c.MysqlBinlogRetentionPeriod
	}

	var res *oapi.CreateDbaasServiceMysqlResponse
	decorateAsyncOperation(fmt.Sprintf("Creating Database Service %q...", c.Name), func() {
		res, err = cs.CreateDbaasServiceMysqlWithResponse(ctx, oapi.DbaasServiceName(c.Name), databaseService)
	})
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
