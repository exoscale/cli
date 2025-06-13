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

<<<<<<< Updated upstream:cmd/dbaas_update_mysql.go
	ctx := gContext
=======
	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_mysql.go

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
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

<<<<<<< Updated upstream:cmd/dbaas_update_mysql.go
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MysqlIPFilter)) {
		databaseService.IPFilter = c.MysqlIPFilter
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Plan)) {
		databaseService.Plan = c.Plan
=======
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.MysqlIPFilter)) {
		databaseService.IpFilter = &c.MysqlIPFilter
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Plan)) {
		databaseService.Plan = &c.Plan
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_mysql.go
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.TerminationProtection)) {
		databaseService.TerminationProtection = &c.TerminationProtection
		updated = true
	}

<<<<<<< Updated upstream:cmd/dbaas_update_mysql.go
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceDOW)) &&
		cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceTime)) {
		databaseService.Maintenance = &v3.UpdateDBAASServiceMysqlRequestMaintenance{
			Dow:  v3.UpdateDBAASServiceMysqlRequestMaintenanceDow(c.MaintenanceDOW),
=======
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.MaintenanceDOW)) &&
		cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.MaintenanceTime)) {
		databaseService.Maintenance = &struct {
			Dow  oapi.UpdateDbaasServiceMysqlJSONBodyMaintenanceDow `json:"dow"`
			Time string                                             `json:"time"`
		}{
			Dow:  oapi.UpdateDbaasServiceMysqlJSONBodyMaintenanceDow(c.MaintenanceDOW),
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_mysql.go
			Time: c.MaintenanceTime,
		}
		updated = true
	}

<<<<<<< Updated upstream:cmd/dbaas_update_mysql.go
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MysqlSettings)) {
		_, err := validateDatabaseServiceSettings(c.MysqlSettings, settingsSchema.Settings.Mysql)
=======
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.MysqlSettings)) {
		settings, err := validateDatabaseServiceSettings(
			c.MysqlSettings,
			settingsSchema.JSON200.Settings.Mysql,
		)
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_mysql.go
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		settings := &v3.JSONSchemaMysql{}
		if err = json.Unmarshal([]byte(c.MysqlSettings), settings); err != nil {
			return fmt.Errorf("invalid settings: %w", err)

		}
		databaseService.MysqlSettings = *settings
		updated = true
	}

<<<<<<< Updated upstream:cmd/dbaas_update_mysql.go
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MysqlMigrationHost)) {
		databaseService.Migration = &v3.UpdateDBAASServiceMysqlRequestMigration{
			Host: c.MysqlMigrationHost,
			Port: c.MysqlMigrationPort,
=======
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.MysqlMigrationHost)) {
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
			Dbname:   utils.NonEmptyStringPtr(c.MysqlMigrationDBName),
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_mysql.go
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

<<<<<<< Updated upstream:cmd/dbaas_update_mysql.go
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MysqlBinlogRetentionPeriod)) {
		databaseService.BinlogRetentionPeriod = c.MysqlBinlogRetentionPeriod
=======
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.MysqlBinlogRetentionPeriod)) {
		databaseService.BinlogRetentionPeriod = &c.MysqlBinlogRetentionPeriod
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_mysql.go
		updated = true
	}

	if updated {
<<<<<<< Updated upstream:cmd/dbaas_update_mysql.go
		op, err := client.UpdateDBAASServiceMysql(ctx, c.Name, databaseService)
=======
		var res *oapi.UpdateDbaasServiceMysqlResponse
		utils.DecorateAsyncOperation(fmt.Sprintf("Updating Database Service %q...", c.Name), func() {
			res, err = globalstate.EgoscaleClient.UpdateDbaasServiceMysqlWithResponse(ctx, oapi.DbaasServiceName(c.Name), databaseService)
		})
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_mysql.go
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
