package cmd

import (
	"fmt"
	"net/http"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
	"github.com/spf13/cobra"
)

func (c *dbServiceUpdateCmd) updateRedis(cmd *cobra.Command, _ []string) error {
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
		}{}
		if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceDOW)) {
			databaseService.Maintenance.Dow = oapi.UpdateDbaasServiceRedisJSONBodyMaintenanceDow(c.MaintenanceDOW)
		}
		if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceTime)) {
			databaseService.Maintenance.Time = c.MaintenanceTime
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

	if updated {
		fmt.Printf("Updating Database Service %q...\n", c.Name)

		res, err := cs.UpdateDbaasServiceRedisWithResponse(ctx, oapi.DbaasServiceName(c.Name), databaseService)
		if err != nil {
			return err
		}
		if res.StatusCode() != http.StatusOK {
			return fmt.Errorf("API request error: unexpected status %s", res.Status())
		}
	}

	if !gQuiet {
		return output((&dbServiceShowCmd{Zone: c.Zone, Name: c.Name}).showDatabaseServiceRedis(ctx))
	}

	return nil
}
