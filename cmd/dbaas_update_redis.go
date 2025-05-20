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

	ctx := GContext

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))

	if err != nil {
		return fmt.Errorf("unable to create client: %w", err)
	}

	databaseService := v3.UpdateDBAASServiceRedisRequest{}

	settingsSchema, err := client.GetDBAASSettingsRedis(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve database Service settings: %w", err)
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.RedisIPFilter)) {
		databaseService.IPFilter = c.RedisIPFilter
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
		databaseService.Maintenance = &v3.UpdateDBAASServiceRedisRequestMaintenance{
			Dow:  v3.UpdateDBAASServiceRedisRequestMaintenanceDow(c.MaintenanceDOW),
			Time: c.MaintenanceTime,
		}
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.RedisSettings)) {
		_, err = validateDatabaseServiceSettings(
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

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.RedisMigrationHost)) {
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
		op, err := client.UpdateDBAASServiceRedis(ctx, c.Name, databaseService)
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
