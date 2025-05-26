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

func (c *dbaasServiceUpdateCmd) updateMysql(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := GContext

	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	databaseService := v3.UpdateDBAASServiceMysqlRequest{}

	settingsSchema, err := client.GetDBAASSettingsMysql(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve database Service settings: %w", err)
	}

	if c.MysqlBackupSchedule != "" {
		bh, bm, err := parseDatabaseBackupSchedule(c.MysqlBackupSchedule)
		if err != nil {
			return err
		}

		databaseService.BackupSchedule = &v3.UpdateDBAASServiceMysqlRequestBackupSchedule{
			BackupHour:   bh,
			BackupMinute: bm,
		}

		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.MysqlIPFilter)) {
		databaseService.IPFilter = c.MysqlIPFilter
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.Plan)) {
		databaseService.Plan = c.Plan
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.TerminationProtection)) {
		databaseService.TerminationProtection = &c.TerminationProtection
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.MaintenanceDOW)) &&
		cmd.Flags().Changed(MustCLICommandFlagName(c, &c.MaintenanceTime)) {
		databaseService.Maintenance = &v3.UpdateDBAASServiceMysqlRequestMaintenance{
			Dow:  v3.UpdateDBAASServiceMysqlRequestMaintenanceDow(c.MaintenanceDOW),
			Time: c.MaintenanceTime,
		}
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.MysqlSettings)) {
		_, err := validateDatabaseServiceSettings(c.MysqlSettings, settingsSchema.Settings.Mysql)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		settings := &v3.JSONSchemaMysql{}
		if err = json.Unmarshal([]byte(c.MysqlSettings), settings); err != nil {
			return fmt.Errorf("invalid settings: %w", err)

		}
		databaseService.MysqlSettings = settings
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.MysqlMigrationHost)) {
		databaseService.Migration = &v3.UpdateDBAASServiceMysqlRequestMigration{
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
			method := v3.EnumMigrationMethod(c.MysqlMigrationMethod)
			databaseService.Migration.Method = method
		}
		if len(c.MysqlMigrationIgnoreDbs) > 0 {
			dbsJoin := strings.Join(c.MysqlMigrationIgnoreDbs, ",")
			databaseService.Migration.IgnoreDbs = dbsJoin
		}
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.MysqlBinlogRetentionPeriod)) {
		databaseService.BinlogRetentionPeriod = c.MysqlBinlogRetentionPeriod
		updated = true
	}

	if updated {
		op, err := client.UpdateDBAASServiceMysql(ctx, c.Name, databaseService)
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
		}).showDatabaseServiceMysql(ctx))
	}

	return nil
}
