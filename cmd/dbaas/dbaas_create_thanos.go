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

func (c *dbaasServiceCreateCmd) createThanos(_ *cobra.Command, _ []string) error {
	var err error

	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))

	if err != nil {
		return fmt.Errorf("unable to create client: %w", err)
	}

	databaseService := v3.CreateDBAASServiceThanosRequest{
		Plan:                  c.Plan,
		TerminationProtection: &c.TerminationProtection,
	}

	if len(c.ThanosIPFilter) > 0 {
		databaseService.IPFilter = c.ThanosIPFilter
	}

	if c.MaintenanceDOW != "" && c.MaintenanceTime != "" {
		databaseService.Maintenance = &v3.CreateDBAASServiceThanosRequestMaintenance{
			Dow:  v3.CreateDBAASServiceThanosRequestMaintenanceDow(c.MaintenanceDOW),
			Time: c.MaintenanceTime,
		}
	}

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
		if err := json.Unmarshal([]byte(c.ThanosSettings), &settings); err != nil {
			return err
		}

		databaseService.ThanosSettings = settings
	}

	op, err := client.CreateDBAASServiceThanos(ctx, c.Name, databaseService)

	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Creating DBaaS Thanos service %q", c.Name), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	if err != nil {
		return err
	}

	serviceName := op.Reference.ID.String()

	if !globalstate.Quiet {
		return c.OutputFunc((&dbaasServiceShowCmd{
			Name: serviceName,
			Zone: c.Zone,
		}).showDatabaseServiceThanos(ctx))
	}

	return nil
}
