// AI-modified by hermes-agent - not reviewed yet
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

func (c *dbaasServiceUpdateCmd) updateClickhouse(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return fmt.Errorf("unable to create client: %w", err)
	}

	databaseService := v3.UpdateDBAASServiceClickhouseRequest{}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.ClickhouseIPFilter)) {
		databaseService.IPFilter = c.ClickhouseIPFilter
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
		databaseService.Maintenance = &v3.UpdateDBAASServiceClickhouseRequestMaintenance{
			Dow:  v3.UpdateDBAASServiceClickhouseRequestMaintenanceDow(c.MaintenanceDOW),
			Time: c.MaintenanceTime,
		}
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.ClickhouseSettings)) {
		if c.ClickhouseSettings != "" {
			settings := &v3.JSONSchemaClickhouse{}
			if err := json.Unmarshal([]byte(c.ClickhouseSettings), settings); err != nil {
				return err
			}
			databaseService.ClickhouseSettings = settings
		}
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.ClickhouseVersion)) {
		databaseService.Version = c.ClickhouseVersion
		updated = true
	}

	if updated {
		op, err := client.UpdateDBAASServiceClickhouse(ctx, c.Name, databaseService)
		if err != nil {
			return err
		}

		utils.DecorateAsyncOperation(fmt.Sprintf("Updating DBaaS ClickHouse service %q", c.Name), func() {
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
		}).showDatabaseServiceClickhouse(ctx))
	}
	return nil
}
