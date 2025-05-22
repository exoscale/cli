package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

func (c *dbaasServiceUpdateCmd) updateOpensearch(cmd *cobra.Command, _ []string) error {
	ctx := GContext

	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	updated := false

	db := v3.UpdateDBAASServiceOpensearchRequest{}

	if len(c.OpensearchIPFilter) > 0 {
		db.IPFilter = c.OpensearchIPFilter
		updated = true
	}

	if c.MaintenanceDOW != "" && c.MaintenanceTime != "" {
		db.Maintenance = &v3.UpdateDBAASServiceOpensearchRequestMaintenance{
			Dow:  v3.UpdateDBAASServiceOpensearchRequestMaintenanceDow(c.MaintenanceDOW),
			Time: c.MaintenanceTime,
		}
		updated = true
	}

	if c.OpensearchSettings != "" {
		settingsSchema, err := client.GetDBAASSettingsOpensearch(ctx)
		if err != nil {
			return fmt.Errorf("unable to retrieve Database Service settings: %w", err)
		}

		_, err = validateDatabaseServiceSettings(
			c.OpensearchSettings,
			settingsSchema.Settings.Opensearch,
		)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}

		settings := &v3.JSONSchemaOpensearch{}
		if err = json.Unmarshal([]byte(c.OpensearchSettings), settings); err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}

		db.OpensearchSettings = *settings
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.OpensearchDashboardEnabled)) {
		db.OpensearchDashboards.Enabled = &c.OpensearchDashboardEnabled
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.OpensearchDashboardRequestTimeout)) {
		db.OpensearchDashboards.OpensearchRequestTimeout = c.OpensearchDashboardRequestTimeout
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.OpensearchDashboardRequestTimeout)) {
		db.OpensearchDashboards.MaxOldSpaceSize = c.OpensearchDashboardMaxOldSpaceSize
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.OpensearchIndexTemplateMappingNestedObjectsLimit)) {
		db.IndexTemplate.MappingNestedObjectsLimit = c.OpensearchIndexTemplateMappingNestedObjectsLimit
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.OpensearchIndexTemplateNumberOfReplicas)) {
		db.IndexTemplate.NumberOfReplicas = c.OpensearchIndexTemplateNumberOfReplicas
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.OpensearchIndexTemplateNumberOfShards)) {
		db.IndexTemplate.NumberOfShards = c.OpensearchIndexTemplateNumberOfShards
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.Plan)) {
		db.Plan = c.Plan
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.TerminationProtection)) {
		db.TerminationProtection = &c.TerminationProtection
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.OpensearchKeepIndexRefreshInterval)) {
		db.KeepIndexRefreshInterval = &c.OpensearchKeepIndexRefreshInterval
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.OpensearchIndexPatterns)) {
		db.IndexPatterns = make([]v3.UpdateDBAASServiceOpensearchRequestIndexPatterns, 0)
		err := json.Unmarshal([]byte(c.OpensearchIndexPatterns), &db.IndexPatterns)
		if err != nil {
			return fmt.Errorf("failed to decode Opensearch index patterns JSON: %w", err)
		}
		updated = true
	}

	if updated {
		op, err := client.UpdateDBAASServiceOpensearch(ctx, c.Name, db)
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
		}).showDatabaseServiceOpensearch(ctx))
	}

	return nil
}
