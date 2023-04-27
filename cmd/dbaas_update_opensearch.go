package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
	"github.com/spf13/cobra"
)

func (c *dbaasServiceUpdateCmd) updateOpensearch(cmd *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	updated := false
	db := oapi.UpdateDbaasServiceOpensearchJSONRequestBody{
		IndexTemplate: &struct {
			MappingNestedObjectsLimit *int64 "json:\"mapping-nested-objects-limit,omitempty\""
			NumberOfReplicas          *int64 "json:\"number-of-replicas,omitempty\""
			NumberOfShards            *int64 "json:\"number-of-shards,omitempty\""
		}{},
		OpensearchDashboards: &struct {
			Enabled                  *bool  "json:\"enabled,omitempty\""
			MaxOldSpaceSize          *int64 "json:\"max-old-space-size,omitempty\""
			OpensearchRequestTimeout *int64 "json:\"opensearch-request-timeout,omitempty\""
		}{},
	}

	if len(c.OpensearchIPFilter) > 0 {
		db.IpFilter = &c.OpensearchIPFilter
		updated = true
	}

	if c.MaintenanceDOW != "" && c.MaintenanceTime != "" {
		db.Maintenance = &struct {
			Dow  oapi.UpdateDbaasServiceOpensearchJSONBodyMaintenanceDow `json:"dow"`
			Time string                                                  `json:"time"`
		}{
			Dow:  oapi.UpdateDbaasServiceOpensearchJSONBodyMaintenanceDow(c.MaintenanceDOW),
			Time: c.MaintenanceTime,
		}
		updated = true
	}

	if c.OpensearchSettings != "" {
		settingsSchema, err := globalstate.EgoscaleClient.GetDbaasSettingsOpensearchWithResponse(ctx)
		if err != nil {
			return fmt.Errorf("unable to retrieve Database Service settings: %w", err)
		}
		if settingsSchema.StatusCode() != http.StatusOK {
			return fmt.Errorf("API request error: unexpected status %s", settingsSchema.Status())
		}

		settings, err := validateDatabaseServiceSettings(
			c.OpensearchSettings,
			settingsSchema.JSON200.Settings.Opensearch,
		)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		db.OpensearchSettings = &settings
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.OpensearchMaxIndexCount)) {
		db.MaxIndexCount = &c.OpensearchMaxIndexCount
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.OpensearchDashboardEnabled)) {
		db.OpensearchDashboards.Enabled = &c.OpensearchDashboardEnabled
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.OpensearchDashboardRequestTimeout)) {
		db.OpensearchDashboards.OpensearchRequestTimeout = &c.OpensearchDashboardRequestTimeout
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.OpensearchDashboardRequestTimeout)) {
		db.OpensearchDashboards.MaxOldSpaceSize = &c.OpensearchDashboardMaxOldSpaceSize
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.OpensearchIndexTemplateMappingNestedObjectsLimit)) {
		db.IndexTemplate.MappingNestedObjectsLimit = &c.OpensearchIndexTemplateMappingNestedObjectsLimit
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.OpensearchIndexTemplateNumberOfReplicas)) {
		db.IndexTemplate.NumberOfReplicas = &c.OpensearchIndexTemplateNumberOfReplicas
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.OpensearchIndexTemplateNumberOfShards)) {
		db.IndexTemplate.NumberOfShards = &c.OpensearchIndexTemplateNumberOfShards
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Plan)) {
		db.Plan = &c.Plan
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.TerminationProtection)) {
		db.TerminationProtection = &c.TerminationProtection
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.OpensearchKeepIndexRefreshInterval)) {
		db.KeepIndexRefreshInterval = &c.OpensearchKeepIndexRefreshInterval
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.OpensearchIndexPatterns)) {
		db.IndexPatterns = &[]struct {
			MaxIndexCount    *int64                                                                  `json:"max-index-count,omitempty"`
			Pattern          *string                                                                 `json:"pattern,omitempty"`
			SortingAlgorithm *oapi.UpdateDbaasServiceOpensearchJSONBodyIndexPatternsSortingAlgorithm `json:"sorting-algorithm,omitempty"`
		}{}

		err := json.Unmarshal([]byte(c.OpensearchIndexPatterns), db.IndexPatterns)
		if err != nil {
			return fmt.Errorf("failed to decode Opensearch index patterns JSON: %w", err)
		}
		updated = true
	}

	var err error
	if updated {
		var res *oapi.UpdateDbaasServiceOpensearchResponse
		decorateAsyncOperation(fmt.Sprintf("Updating Database Service %q...", c.Name), func() {
			res, err = globalstate.EgoscaleClient.UpdateDbaasServiceOpensearchWithResponse(ctx, oapi.DbaasServiceName(c.Name), db)
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
		}).showDatabaseServiceOpensearch(ctx))
	}

	return nil
}
