package cmd

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
	"github.com/spf13/cobra"
)

func (c *dbaasServiceUpdateCmd) updateGrafana(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	databaseService := oapi.UpdateDbaasServiceGrafanaJSONRequestBody{}

	settingsSchema, err := globalstate.EgoscaleClient.GetDbaasSettingsGrafanaWithResponse(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve database Service settings: %w", err)
	}
	if settingsSchema.StatusCode() != http.StatusOK {
		return fmt.Errorf("API request error: unexpected status %s", settingsSchema.Status())
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.GrafanaIPFilter)) {
		databaseService.IpFilter = &c.GrafanaIPFilter
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
			Dow  oapi.UpdateDbaasServiceGrafanaJSONBodyMaintenanceDow `json:"dow"`
			Time string                                               `json:"time"`
		}{
			Dow:  oapi.UpdateDbaasServiceGrafanaJSONBodyMaintenanceDow(c.MaintenanceDOW),
			Time: c.MaintenanceTime,
		}
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.GrafanaSettings)) {
		settings, err := validateDatabaseServiceSettings(
			c.GrafanaSettings,
			settingsSchema.JSON200.Settings.Grafana,
		)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		databaseService.GrafanaSettings = &settings
		updated = true
	}

	if updated {
		var res *oapi.UpdateDbaasServiceGrafanaResponse
		decorateAsyncOperation(fmt.Sprintf("Updating Database Service %q...", c.Name), func() {
			res, err = globalstate.EgoscaleClient.UpdateDbaasServiceGrafanaWithResponse(ctx, oapi.DbaasServiceName(c.Name), databaseService)
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
		}).showDatabaseServiceGrafana(ctx))
	}

	return nil
}
