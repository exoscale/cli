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

func (c *dbaasServiceUpdateCmd) updateRedis(cmd *cobra.Command, _ []string) error {
	var updated bool

<<<<<<< Updated upstream:cmd/dbaas_update_redis.go
	ctx := gContext
=======
	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_redis.go

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))

	if err != nil {
		return fmt.Errorf("unable to create client: %w", err)
	}

	databaseService := v3.UpdateDBAASServiceRedisRequest{}

	settingsSchema, err := client.GetDBAASSettingsRedis(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve database Service settings: %w", err)
	}

<<<<<<< Updated upstream:cmd/dbaas_update_redis.go
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.RedisIPFilter)) {
		databaseService.IPFilter = c.RedisIPFilter
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Plan)) {
		databaseService.Plan = c.Plan
=======
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.RedisIPFilter)) {
		databaseService.IpFilter = &c.RedisIPFilter
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Plan)) {
		databaseService.Plan = &c.Plan
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_redis.go
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.TerminationProtection)) {
		databaseService.TerminationProtection = &c.TerminationProtection
		updated = true
	}

<<<<<<< Updated upstream:cmd/dbaas_update_redis.go
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceDOW)) &&
		cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceTime)) {
		databaseService.Maintenance = &v3.UpdateDBAASServiceRedisRequestMaintenance{
			Dow:  v3.UpdateDBAASServiceRedisRequestMaintenanceDow(c.MaintenanceDOW),
=======
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.MaintenanceDOW)) &&
		cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.MaintenanceTime)) {
		databaseService.Maintenance = &struct {
			Dow  oapi.UpdateDbaasServiceRedisJSONBodyMaintenanceDow `json:"dow"`
			Time string                                             `json:"time"`
		}{
			Dow:  oapi.UpdateDbaasServiceRedisJSONBodyMaintenanceDow(c.MaintenanceDOW),
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_redis.go
			Time: c.MaintenanceTime,
		}
		updated = true
	}

<<<<<<< Updated upstream:cmd/dbaas_update_redis.go
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.RedisSettings)) {
		_, err = validateDatabaseServiceSettings(
=======
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.RedisSettings)) {
		settings, err := validateDatabaseServiceSettings(
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_redis.go
			c.RedisSettings,
			settingsSchema.Settings.Redis,
		)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}

		settings := &v3.JSONSchemaRedis{}
		if err := json.Unmarshal([]byte(c.RedisSettings), &settings); err != nil {
			return err
		}

		databaseService.RedisSettings = settings
		updated = true
	}

<<<<<<< Updated upstream:cmd/dbaas_update_redis.go
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.RedisMigrationHost)) {
		databaseService.Migration = &v3.UpdateDBAASServiceRedisRequestMigration{
			Host: c.RedisMigrationHost,
			Port: c.RedisMigrationPort,
		}
		if c.RedisMigrationPassword != "" {
			databaseService.Migration.Password = c.RedisMigrationPassword
		}
		if c.RedisMigrationUsername != "" {
			databaseService.Migration.Username = c.RedisMigrationUsername
		}
		if c.RedisMigrationDBName != "" {
			databaseService.Migration.Dbname = c.RedisMigrationDBName
=======
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.RedisMigrationHost)) {
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
			Host:     c.RedisMigrationHost,
			Port:     c.RedisMigrationPort,
			Password: utils.NonEmptyStringPtr(c.RedisMigrationPassword),
			Username: utils.NonEmptyStringPtr(c.RedisMigrationUsername),
			Dbname:   utils.NonEmptyStringPtr(c.RedisMigrationDBName),
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_redis.go
		}
		if c.RedisMigrationSSL {
			databaseService.Migration.SSL = &c.RedisMigrationSSL
		}
		if c.RedisMigrationMethod != "" {
			method := v3.EnumMigrationMethod(c.RedisMigrationMethod)
			databaseService.Migration.Method = method
		}
		if len(c.RedisMigrationIgnoreDbs) > 0 {
			dbsJoin := strings.Join(c.RedisMigrationIgnoreDbs, ",")
			databaseService.Migration.IgnoreDbs = dbsJoin
		}
		updated = true
	}

	if updated {
<<<<<<< Updated upstream:cmd/dbaas_update_redis.go
		op, err := client.UpdateDBAASServiceRedis(ctx, c.Name, databaseService)
=======
		var res *oapi.UpdateDbaasServiceRedisResponse
		utils.DecorateAsyncOperation(fmt.Sprintf("Updating Database Service %q...", c.Name), func() {
			res, err = globalstate.EgoscaleClient.UpdateDbaasServiceRedisWithResponse(ctx, oapi.DbaasServiceName(c.Name), databaseService)
		})
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_redis.go
		if err != nil {
			if errors.Is(err, v3.ErrNotFound) {
				return fmt.Errorf("resource not found in zone %q", c.Zone)
			}
			return err
		}
		decorateAsyncOperation(fmt.Sprintf("Updating Database Service %q...", c.Name), func() {
			op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		})
		if err != nil {
			return err
		}

	}

	if !globalstate.Quiet {
		return c.OutputFunc((&dbaasServiceShowCmd{
			Name: c.Name,
			Zone: c.Zone,
		}).showDatabaseServiceRedis(ctx))
	}

	return nil
}
