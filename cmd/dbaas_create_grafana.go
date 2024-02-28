package cmd

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
)

func (c *dbaasServiceCreateCmd) createGrafana(_ *cobra.Command, _ []string) error {
	var err error

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	databaseService := oapi.CreateDbaasServiceGrafanaJSONRequestBody{
		Plan: c.Plan,
	}

	settingsSchema, err := globalstate.EgoscaleClient.GetDbaasSettingsGrafanaWithResponse(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve Database Service settings: %w", err)
	}
	if settingsSchema.StatusCode() != http.StatusOK {
		return fmt.Errorf("API request error: unexpected status %s", settingsSchema.Status())
	}

	if len(c.GrafanaIPFilter) > 0 {
		databaseService.IpFilter = &c.GrafanaIPFilter
	}

	if c.MaintenanceDOW != "" && c.MaintenanceTime != "" {
		databaseService.Maintenance = &struct {
			Dow  oapi.CreateDbaasServiceGrafanaJSONBodyMaintenanceDow `json:"dow"`
			Time string                                               `json:"time"`
		}{
			Dow:  oapi.CreateDbaasServiceGrafanaJSONBodyMaintenanceDow(c.MaintenanceDOW),
			Time: c.MaintenanceTime,
		}
	}

	if c.GrafanaSettings != "" {
		settings, err := validateDatabaseServiceSettings(
			c.GrafanaSettings,
			settingsSchema.JSON200.Settings.Grafana,
		)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		databaseService.GrafanaSettings = &settings
	}

	var res *oapi.CreateDbaasServiceGrafanaResponse
	decorateAsyncOperation(fmt.Sprintf("Creating Database Service %q...", c.Name), func() {
		res, err = globalstate.EgoscaleClient.CreateDbaasServiceGrafanaWithResponse(ctx, oapi.DbaasServiceName(c.Name), databaseService)
	})
	if err != nil {
		return err
	}
	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("API request error: unexpected status %s", res.Status())
	}

	if !globalstate.Quiet {
		return c.outputFunc((&dbaasServiceShowCmd{
			Name: c.Name,
			Zone: c.Zone,
		}).showDatabaseServiceGrafana(ctx))
	}

	return nil
}
