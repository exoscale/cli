package dbaas

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

func (c *dbaasServiceUpdateCmd) updateThanos(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return fmt.Errorf("unable to create client: %w", err)
	}

	databaseService := v3.UpdateDBAASServiceThanosRequest{}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.ThanosIPFilter)) {
		databaseService.IPFilter = c.ThanosIPFilter
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Plan)) {
		databaseService.Plan = c.Plan
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.TerminationProtection)) {
		databaseService.TerminationProtection = &c.TerminationProtection
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.MaintenanceDOW)) &&
		cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.MaintenanceTime)) {
		databaseService.Maintenance = &v3.UpdateDBAASServiceThanosRequestMaintenance{
			Dow:  v3.UpdateDBAASServiceThanosRequestMaintenanceDow(c.MaintenanceDOW),
			Time: c.MaintenanceTime,
		}
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.ThanosSettings)) {
		if c.ThanosSettings != "" {
			settingsSchema, err := client.GetDBAASSettingsThanos(ctx)
			if err != nil {
				return fmt.Errorf("unable to retrieve Database Service settings: %w", err)
			}

			_, err = validateDatabaseServiceSettings(
				c.ThanosSettings,
				settingsSchema.Settings.Thanos,
			)
			if err != nil {
				return fmt.Errorf("invalid settings: %w", err)
			}

			settings := &v3.JSONSchemaThanos{}
			if err = json.Unmarshal([]byte(c.ThanosSettings), settings); err != nil {
				return fmt.Errorf("invalid settings: %w", err)
			}

			databaseService.ThanosSettings = settings
		}
		updated = true
	}

	if updated {
		op, err := client.UpdateDBAASServiceThanos(ctx, c.Name, databaseService)
		if err != nil {
			return err
		}

		utils.DecorateAsyncOperation(fmt.Sprintf("Updating DBaaS Thanos service %q", c.Name), func() {
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
		}).showDatabaseServiceThanos(ctx))
	}
	return nil
}
