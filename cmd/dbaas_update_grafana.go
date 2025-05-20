package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

func (c *dbaasServiceUpdateCmd) updateGrafana(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := GContext
	var err error

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	databaseService := v3.UpdateDBAASServiceGrafanaRequest{}

	settingsSchema, err := client.GetDBAASSettingsGrafana(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve database Service settings: %w", err)
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.GrafanaIPFilter)) {
		databaseService.IPFilter = c.GrafanaIPFilter
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
		databaseService.Maintenance = &v3.UpdateDBAASServiceGrafanaRequestMaintenance{
			Time: c.MaintenanceTime,
			Dow:  v3.UpdateDBAASServiceGrafanaRequestMaintenanceDow(c.MaintenanceDOW),
		}
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.GrafanaSettings)) {
		_, err := validateDatabaseServiceSettings(
			c.GrafanaSettings,
			settingsSchema.Settings.Grafana,
		)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}

		settings := &v3.JSONSchemaGrafana{}

		if err = json.Unmarshal([]byte(c.GrafanaSettings), settings); err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}

		databaseService.GrafanaSettings = settings
		updated = true
	}

	if updated {

		op, err := client.UpdateDBAASServiceGrafana(ctx, c.Name, databaseService)
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
		}).showDatabaseServiceGrafana(ctx))
	}

	return nil
}
