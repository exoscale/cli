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

func (c *dbaasServiceCreateCmd) createMysql(_ *cobra.Command, _ []string) error {
	var err error

	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	databaseService := v3.CreateDBAASServiceMysqlRequest{
		Plan:                  c.Plan,
		TerminationProtection: &c.TerminationProtection,
	}

	if c.MysqlVersion != "" {
		databaseService.Version = c.MysqlVersion
	}

	settingsSchema, err := client.GetDBAASSettingsMysql(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve Database Service settings: %w", err)
	}

	if c.MysqlForkFrom != "" {
		databaseService.ForkFromService = v3.DBAASServiceName(c.MysqlForkFrom)
		if c.MysqlRecoveryBackupTime != "" {
			databaseService.RecoveryBackupTime = c.MysqlRecoveryBackupTime
		}
	}

	if c.MysqlAdminPassword != "" {
		databaseService.AdminPassword = c.MysqlAdminPassword
	}

	if c.MysqlAdminUsername != "" {
		databaseService.AdminUsername = c.MysqlAdminUsername
	}

	if c.MysqlBackupSchedule != "" {
		bh, bm, err := parseDatabaseBackupSchedule(c.MysqlBackupSchedule)
		if err != nil {
			return err
		}

		databaseService.BackupSchedule = &v3.CreateDBAASServiceMysqlRequestBackupSchedule{
			BackupHour:   *bh,
			BackupMinute: *bm,
		}
	}

	if len(c.MysqlIPFilter) > 0 {
		databaseService.IPFilter = c.MysqlIPFilter
	}

	if c.MaintenanceDOW != "" && c.MaintenanceTime != "" {
		databaseService.Maintenance = &v3.CreateDBAASServiceMysqlRequestMaintenance{
			Dow:  v3.CreateDBAASServiceMysqlRequestMaintenanceDow(c.MaintenanceDOW),
			Time: c.MaintenanceTime,
		}
	}

	if c.MysqlSettings != "" {
		_, err := validateDatabaseServiceSettings(c.MysqlSettings, settingsSchema.Settings.Mysql)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		settings := &v3.JSONSchemaMysql{}
		if err = json.Unmarshal([]byte(c.MysqlSettings), settings); err != nil {
			return fmt.Errorf("invalid settings: %w", err)

		}
		databaseService.MysqlSettings = settings
	}

	if c.MysqlMigrationHost != "" {
		databaseService.Migration = &v3.CreateDBAASServiceMysqlRequestMigration{
			Host: c.MysqlMigrationHost,
			Port: c.MysqlMigrationPort,
		}
		if c.MysqlMigrationPassword != "" {
			databaseService.Migration.Password = c.MysqlMigrationPassword
		}
		if c.MysqlMigrationUsername != "" {
			databaseService.Migration.Username = c.MysqlMigrationUsername
		}
		if c.MysqlMigrationDBName != "" {
			databaseService.Migration.Dbname = c.MysqlMigrationDBName
		}
		if c.MysqlMigrationSSL {
			databaseService.Migration.SSL = &c.MysqlMigrationSSL
		}
		if c.MysqlMigrationMethod != "" {
			databaseService.Migration.Method = v3.EnumMigrationMethod(c.MysqlMigrationMethod)
		}
		if len(c.MysqlMigrationIgnoreDbs) > 0 {
			databaseService.Migration.IgnoreDbs = strings.Join(c.MysqlMigrationIgnoreDbs, ",")
		}
	}

	if c.MysqlBinlogRetentionPeriod > 0 {
		databaseService.BinlogRetentionPeriod = c.MysqlBinlogRetentionPeriod
	}

	op, err := client.CreateDBAASServiceMysql(ctx, c.Name, databaseService)
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
		}).showDatabaseServiceMysql(ctx))
	}

	return nil
}
