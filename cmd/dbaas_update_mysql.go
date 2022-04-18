package cmd

import (
	"fmt"
	"net/http"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
	"github.com/spf13/cobra"
)

func (c *dbaasServiceUpdateCmd) updateMysql(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	databaseService := oapi.UpdateDbaasServiceMysqlJSONRequestBody{}

	settingsSchema, err := cs.GetDbaasSettingsMysqlWithResponse(ctx)
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

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.TerminationProtection)) {
		databaseService.TerminationProtection = &c.TerminationProtection
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
			Dbname    *string                     `json:"dbname,omitempty"`
			Host      string                      `json:"host"`
			IgnoreDbs *string                     `json:"ignore-dbs,omitempty"`
			Method    *oapi.EnumPgMigrationMethod `json:"method,omitempty"`
			Password  *string                     `json:"password,omitempty"`
			Port      int64                       `json:"port"`
			Ssl       *bool                       `json:"ssl,omitempty"`
			Username  *string                     `json:"username,omitempty"`
		}{
			Host:     c.MysqlMigrationHost,
			Port:     c.MysqlMigrationPort,
			Password: nonEmptyStringPtr(c.MysqlMigrationPassword),
			Username: nonEmptyStringPtr(c.MysqlMigrationUsername),
			Dbname:   nonEmptyStringPtr(c.MysqlMigrationDbName),
		}
		if c.MysqlMigrationSSL {
			databaseService.Migration.Ssl = &c.MysqlMigrationSSL
		}
		if c.MysqlMigrationMethod != "" {
			method := oapi.EnumPgMigrationMethod(c.MysqlMigrationMethod)
			databaseService.Migration.Method = &method
		}
		if len(c.MysqlMigrationIgnoreDbs) > 0 {
			dbsJoin := strings.Join(c.MysqlMigrationIgnoreDbs, ",")
			databaseService.Migration.IgnoreDbs = &dbsJoin
		}
		updated = true
	}

	if updated {
		var res *oapi.UpdateDbaasServiceMysqlResponse
		decorateAsyncOperation(fmt.Sprintf("Updating Database Service %q...", c.Name), func() {
			res, err = cs.UpdateDbaasServiceMysqlWithResponse(ctx, oapi.DbaasServiceName(c.Name), databaseService)
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
		}).showDatabaseServiceMysql(ctx))
	}

	return nil
}
