package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
	"github.com/spf13/cobra"
)

func (c *dbaasServiceUpdateCmd) updateRedis(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	databaseService := oapi.UpdateDbaasServiceRedisJSONRequestBody{}

	settingsSchema, err := cs.GetDbaasSettingsRedisWithResponse(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve database Service settings: %w", err)
	}
	if settingsSchema.StatusCode() != http.StatusOK {
		return fmt.Errorf("API request error: unexpected status %s", settingsSchema.Status())
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.RedisIPFilter)) {
		databaseService.IpFilter = &c.RedisIPFilter
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
			Dow  oapi.UpdateDbaasServiceRedisJSONBodyMaintenanceDow `json:"dow"`
			Time string                                             `json:"time"`
		}{
			Dow:  oapi.UpdateDbaasServiceRedisJSONBodyMaintenanceDow(c.MaintenanceDOW),
			Time: c.MaintenanceTime,
		}
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.RedisSettings)) {
		settings, err := validateDatabaseServiceSettings(
			c.RedisSettings,
			settingsSchema.JSON200.Settings.Redis,
		)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		databaseService.RedisSettings = &settings
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.RedisMigrationHost)) {
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
			Dbname:   utils.NonEmptyStringPtr(c.RedisMigrationDbName),
		}
		if c.RedisMigrationSSL {
			databaseService.Migration.Ssl = &c.RedisMigrationSSL
		}
		if c.RedisMigrationMethod != "" {
			method := oapi.EnumMigrationMethod(c.RedisMigrationMethod)
			databaseService.Migration.Method = &method
		}
		if len(c.RedisMigrationIgnoreDbs) > 0 {
			dbsJoin := strings.Join(c.RedisMigrationIgnoreDbs, ",")
			databaseService.Migration.IgnoreDbs = &dbsJoin
		}
		updated = true
	}

	if updated {
		var res *oapi.UpdateDbaasServiceRedisResponse
		decorateAsyncOperation(fmt.Sprintf("Updating Database Service %q...", c.Name), func() {
			res, err = cs.UpdateDbaasServiceRedisWithResponse(ctx, oapi.DbaasServiceName(c.Name), databaseService)
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
		}).showDatabaseServiceRedis(ctx))
	}

	return nil
}
