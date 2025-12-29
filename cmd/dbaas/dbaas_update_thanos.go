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
			var settings map[string]interface{}

			if err := json.Unmarshal([]byte(c.ThanosSettings), &settings); err != nil {
				return err
			}

			thanosSettings := &v3.JSONSchemaThanos{}

			// Parse compactor settings
			if compactor, ok := settings["compactor"].(map[string]interface{}); ok {
				thanosSettings.Compactor = &v3.JSONSchemaThanosCompactor{
					RetentionDays: int(utils.GetSettingFloat64(compactor, "retention.days")),
				}
			}

			// Parse query settings
			if query, ok := settings["query"].(map[string]interface{}); ok {
				thanosSettings.Query = &v3.JSONSchemaThanosQuery{
					QueryDefaultEvaluationInterval: utils.GetSettingString(query, "query.default-evaluation-interval"),
					QueryLookbackDelta:             utils.GetSettingString(query, "query.lookback-delta"),
					QueryMetadataDefaultTimeRange:  utils.GetSettingString(query, "query.metadata.default-time-range"),
					QueryTimeout:                   utils.GetSettingString(query, "query.timeout"),
					StoreLimitsRequestSamples:      int(utils.GetSettingFloat64(query, "store.limits.request-samples")),
					StoreLimitsRequestSeries:       int(utils.GetSettingFloat64(query, "store.limits.request-series")),
				}
			}

			// Parse query-frontend settings
			if queryFrontend, ok := settings["query-frontend"].(map[string]interface{}); ok {
				alignRangeWithStep := utils.GetSettingBool(queryFrontend, "query-range.align-range-with-step")
				thanosSettings.QueryFrontend = &v3.JSONSchemaThanosQueryFrontend{
					QueryRangeAlignRangeWithStep: &alignRangeWithStep,
				}
			}

			databaseService.ThanosSettings = thanosSettings
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
