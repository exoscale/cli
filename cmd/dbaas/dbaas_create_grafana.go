package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
<<<<<<< Updated upstream:cmd/dbaas_create_grafana.go
	v3 "github.com/exoscale/egoscale/v3"
=======
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
>>>>>>> Stashed changes:cmd/dbaas/dbaas_create_grafana.go
)

func (c *dbaasServiceCreateCmd) createGrafana(_ *cobra.Command, _ []string) error {
	var err error

<<<<<<< Updated upstream:cmd/dbaas_create_grafana.go
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}
=======
	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))
>>>>>>> Stashed changes:cmd/dbaas/dbaas_create_grafana.go

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

<<<<<<< Updated upstream:cmd/dbaas_create_grafana.go
	op, err := client.CreateDBAASServiceGrafana(ctx, c.Name, databaseService)
=======
	var res *oapi.CreateDbaasServiceGrafanaResponse
	utils.DecorateAsyncOperation(fmt.Sprintf("Creating Database Service %q...", c.Name), func() {
		res, err = globalstate.EgoscaleClient.CreateDbaasServiceGrafanaWithResponse(ctx, oapi.DbaasServiceName(c.Name), databaseService)
	})
>>>>>>> Stashed changes:cmd/dbaas/dbaas_create_grafana.go
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
