package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

func (c *dbaasServiceCreateCmd) createGrafana(_ *cobra.Command, _ []string) error {
	var err error

	ctx := GContext
	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	databaseService := v3.CreateDBAASServiceGrafanaRequest{
		Plan:                  c.Plan,
		TerminationProtection: &c.TerminationProtection,
	}

	settingsSchema, err := client.GetDBAASSettingsGrafana(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve Database Service settings: %w", err)
	}

	if len(c.GrafanaIPFilter) > 0 {
		databaseService.IPFilter = c.GrafanaIPFilter
	}

	if c.MaintenanceDOW != "" && c.MaintenanceTime != "" {
		databaseService.Maintenance = &v3.CreateDBAASServiceGrafanaRequestMaintenance{
			Time: c.MaintenanceTime,
			Dow:  v3.CreateDBAASServiceGrafanaRequestMaintenanceDow(c.MaintenanceDOW),
		}
	}

	if c.GrafanaSettings != "" {

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
	}

	op, err := client.CreateDBAASServiceGrafana(ctx, c.Name, databaseService)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Creating Database Service %q...", c.Name), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return c.OutputFunc((&dbaasServiceShowCmd{
			Name: c.Name,
			Zone: c.Zone,
		}).showDatabaseServiceGrafana(ctx))
	}

	return nil
}
