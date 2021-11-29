package cmd

import (
	"fmt"
	"net/http"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
	"github.com/spf13/cobra"
)

func (c *dbaasServiceCreateCmd) createRedis(_ *cobra.Command, _ []string) error {
	var err error

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	databaseService := oapi.CreateDbaasServiceRedisJSONRequestBody{
		Plan:                  c.Plan,
		TerminationProtection: &c.TerminationProtection,
	}

	settingsSchema, err := cs.GetDbaasSettingsRedisWithResponse(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve Database Service settings: %w", err)
	}
	if settingsSchema.StatusCode() != http.StatusOK {
		return fmt.Errorf("API request error: unexpected status %s", settingsSchema.Status())
	}

	if c.RedisForkFrom != "" {
		databaseService.ForkFromService = (*oapi.DbaasServiceName)(&c.RedisForkFrom)
		if c.RedisRecoveryBackupName != "" {
			databaseService.RecoveryBackupName = &c.RedisRecoveryBackupName
		}
	}

	if len(c.RedisIPFilter) > 0 {
		databaseService.IpFilter = &c.RedisIPFilter
	}

	if c.MaintenanceDOW != "" && c.MaintenanceTime != "" {
		databaseService.Maintenance.Dow = oapi.CreateDbaasServiceRedisJSONBodyMaintenanceDow(c.MaintenanceDOW)
		databaseService.Maintenance.Time = c.MaintenanceTime
	}

	if c.RedisSettings != "" {
		settings, err := validateDatabaseServiceSettings(c.RedisSettings, settingsSchema.JSON200.Settings.Redis)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		databaseService.RedisSettings = &settings
	}

	var res *oapi.CreateDbaasServiceRedisResponse
	decorateAsyncOperation(fmt.Sprintf("Creating Database Service %q...", c.Name), func() {
		res, err = cs.CreateDbaasServiceRedisWithResponse(ctx, oapi.DbaasServiceName(c.Name), databaseService)
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
		}).showDatabaseServiceRedis(ctx))
	}

	return nil
}
