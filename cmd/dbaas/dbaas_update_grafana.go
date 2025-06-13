package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
<<<<<<< Updated upstream:cmd/dbaas_update_grafana.go
	v3 "github.com/exoscale/egoscale/v3"
=======
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_grafana.go
)

func (c *dbaasServiceUpdateCmd) updateGrafana(cmd *cobra.Command, _ []string) error {
	var updated bool

<<<<<<< Updated upstream:cmd/dbaas_update_grafana.go
	ctx := gContext
	var err error
=======
	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_grafana.go

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	databaseService := v3.UpdateDBAASServiceGrafanaRequest{}

	settingsSchema, err := client.GetDBAASSettingsGrafana(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve database Service settings: %w", err)
	}

<<<<<<< Updated upstream:cmd/dbaas_update_grafana.go
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.GrafanaIPFilter)) {
		databaseService.IPFilter = c.GrafanaIPFilter
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Plan)) {
		databaseService.Plan = c.Plan
=======
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.GrafanaIPFilter)) {
		databaseService.IpFilter = &c.GrafanaIPFilter
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Plan)) {
		databaseService.Plan = &c.Plan
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_grafana.go
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.TerminationProtection)) {
		databaseService.TerminationProtection = &c.TerminationProtection
		updated = true
	}

<<<<<<< Updated upstream:cmd/dbaas_update_grafana.go
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceDOW)) &&
		cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceTime)) {
		databaseService.Maintenance = &v3.UpdateDBAASServiceGrafanaRequestMaintenance{
=======
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.MaintenanceDOW)) &&
		cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.MaintenanceTime)) {
		databaseService.Maintenance = &struct {
			Dow  oapi.UpdateDbaasServiceGrafanaJSONBodyMaintenanceDow `json:"dow"`
			Time string                                               `json:"time"`
		}{
			Dow:  oapi.UpdateDbaasServiceGrafanaJSONBodyMaintenanceDow(c.MaintenanceDOW),
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_grafana.go
			Time: c.MaintenanceTime,
			Dow:  v3.UpdateDBAASServiceGrafanaRequestMaintenanceDow(c.MaintenanceDOW),
		}
		updated = true
	}

<<<<<<< Updated upstream:cmd/dbaas_update_grafana.go
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.GrafanaSettings)) {
		_, err := validateDatabaseServiceSettings(
=======
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.GrafanaSettings)) {
		settings, err := validateDatabaseServiceSettings(
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_grafana.go
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
<<<<<<< Updated upstream:cmd/dbaas_update_grafana.go

		op, err := client.UpdateDBAASServiceGrafana(ctx, c.Name, databaseService)
=======
		var res *oapi.UpdateDbaasServiceGrafanaResponse
		utils.DecorateAsyncOperation(fmt.Sprintf("Updating Database Service %q...", c.Name), func() {
			res, err = globalstate.EgoscaleClient.UpdateDbaasServiceGrafanaWithResponse(ctx, oapi.DbaasServiceName(c.Name), databaseService)
		})
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_grafana.go
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
