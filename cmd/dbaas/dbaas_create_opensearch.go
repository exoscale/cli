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

func (c *dbaasServiceCreateCmd) createOpensearch(cmd *cobra.Command, _ []string) error {
	var err error

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	db := v3.CreateDBAASServiceOpensearchRequest{
		KeepIndexRefreshInterval: &c.OpensearchKeepIndexRefreshInterval,
		Plan:                     c.Plan,
		TerminationProtection:    &c.TerminationProtection,
		OpensearchDashboards:     &v3.CreateDBAASServiceOpensearchRequestOpensearchDashboards{},
		IndexTemplate:            &v3.CreateDBAASServiceOpensearchRequestIndexTemplate{},
	}

	if c.OpensearchForkFromService != "" {
		db.ForkFromService = v3.DBAASServiceName(c.OpensearchForkFromService)
	}
	if c.OpensearchRecoveryBackupName != "" {
		db.RecoveryBackupName = c.OpensearchRecoveryBackupName
	}
	if db.Version != "" {
		db.Version = c.OpensearchVersion
	}

	if len(c.OpensearchIPFilter) > 0 {
		db.IPFilter = c.OpensearchIPFilter
	}

	if c.MaintenanceDOW != "" && c.MaintenanceTime != "" {
		db.Maintenance = &v3.CreateDBAASServiceOpensearchRequestMaintenance{

			Dow:  v3.CreateDBAASServiceOpensearchRequestMaintenanceDow(c.MaintenanceDOW),
			Time: c.MaintenanceTime,
		}
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

		db.OpensearchSettings = settings
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.OpensearchDashboardEnabled)) {
		db.OpensearchDashboards.Enabled = &c.OpensearchDashboardEnabled
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.OpensearchDashboardRequestTimeout)) {
		db.OpensearchDashboards.OpensearchRequestTimeout = c.OpensearchDashboardRequestTimeout
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.OpensearchDashboardRequestTimeout)) {
		db.OpensearchDashboards.MaxOldSpaceSize = c.OpensearchDashboardMaxOldSpaceSize
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.OpensearchIndexTemplateMappingNestedObjectsLimit)) {
		db.IndexTemplate.MappingNestedObjectsLimit = c.OpensearchIndexTemplateMappingNestedObjectsLimit
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.OpensearchIndexTemplateNumberOfReplicas)) {
		db.IndexTemplate.NumberOfReplicas = c.OpensearchIndexTemplateNumberOfReplicas
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.OpensearchIndexTemplateNumberOfShards)) {
		db.IndexTemplate.NumberOfShards = c.OpensearchIndexTemplateNumberOfShards
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.OpensearchIndexPatterns)) {
		db.IndexPatterns = make([]v3.CreateDBAASServiceOpensearchRequestIndexPatterns, 0)
		err := json.Unmarshal([]byte(c.OpensearchIndexPatterns), &db.IndexPatterns)
		if err != nil {
			return fmt.Errorf("failed to decode Opensearch index patterns JSON: %s", err)
		}
	}

	op, err := client.CreateDBAASServiceOpensearch(ctx, c.Name, db)
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Creating Database Service %q...", c.Name), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return c.OutputFunc((&dbaasServiceShowCmd{
			Name: c.Name,
			Zone: c.Zone,
		}).showDatabaseServiceOpensearch(ctx))
	}

	return nil
}
