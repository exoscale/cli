package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
)

func (c *dbaasServiceUpdateCmd) updateMysql(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	databaseService := oapi.UpdateDbaasServiceMysqlJSONRequestBody{}

	settingsSchema, err := globalstate.EgoscaleClient.GetDbaasSettingsMysqlWithResponse(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve database Service settings: %w", err)
	}
	if settingsSchema.StatusCode() != http.StatusOK {
		return fmt.Errorf("API request error: unexpected status %s", settingsSchema.Status())
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

		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MysqlIPFilter)) {
		databaseService.IpFilter = &c.MysqlIPFilter
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Plan)) {
		databaseService.Plan = &c.Plan
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceDOW)) &&
		cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceTime)) {
		databaseService.Maintenance = &struct {
			Dow  oapi.UpdateDbaasServiceMysqlJSONBodyMaintenanceDow `json:"dow"`
			Time string                                             `json:"time"`
		}{
			Dow:  oapi.UpdateDbaasServiceMysqlJSONBodyMaintenanceDow(c.MaintenanceDOW),
			Time: c.MaintenanceTime,
		}
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MysqlSettings)) {
		settings, err := validateDatabaseServiceSettings(
			c.MysqlSettings,
			settingsSchema.JSON200.Settings.Mysql,
		)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		databaseService.MysqlSettings = &settings
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MysqlMigrationHost)) {
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
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MysqlBinlogRetentionPeriod)) {
		databaseService.BinlogRetentionPeriod = &c.MysqlBinlogRetentionPeriod
		updated = true
	}

	if updated {
		var res *oapi.UpdateDbaasServiceMysqlResponse
		decorateAsyncOperation(fmt.Sprintf("Updating Database Service %q...", c.Name), func() {
			res, err = globalstate.EgoscaleClient.UpdateDbaasServiceMysqlWithResponse(ctx, oapi.DbaasServiceName(c.Name), databaseService)
		})
		if err != nil {
			if errors.Is(err, exoapi.ErrNotFound) {
				return fmt.Errorf("resource not found in zone %q", c.Zone)
			}
			return err
		}
		if res.StatusCode() != http.StatusOK {
			return fmt.Errorf("API request error: unexpected status %s", res.Status())
		}
	}

	if !globalstate.Quiet {
		return c.outputFunc((&dbaasServiceShowCmd{
			Name: c.Name,
			Zone: c.Zone,
		}).showDatabaseServiceMysql(ctx))
	}

	return nil
}
