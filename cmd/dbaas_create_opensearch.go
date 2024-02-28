package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
)

func (c *dbaasServiceCreateCmd) createOpensearch(cmd *cobra.Command, _ []string) error {
	var err error

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	db := oapi.CreateDbaasServiceOpensearchJSONRequestBody{
		ForkFromService:          (*oapi.DbaasServiceName)(utils.NonEmptyStringPtr(c.OpensearchForkFromService)),
		KeepIndexRefreshInterval: &c.OpensearchKeepIndexRefreshInterval,
		Plan:                     c.Plan,
		RecoveryBackupName:       utils.NonEmptyStringPtr(c.OpensearchRecoveryBackupName),
		Version:                  utils.NonEmptyStringPtr(c.OpensearchVersion),
		OpensearchDashboards: &struct {
			Enabled                  *bool  "json:\"enabled,omitempty\""
			MaxOldSpaceSize          *int64 "json:\"max-old-space-size,omitempty\""
			OpensearchRequestTimeout *int64 "json:\"opensearch-request-timeout,omitempty\""
		}{},
		IndexTemplate: &struct {
			MappingNestedObjectsLimit *int64 "json:\"mapping-nested-objects-limit,omitempty\""
			NumberOfReplicas          *int64 "json:\"number-of-replicas,omitempty\""
			NumberOfShards            *int64 "json:\"number-of-shards,omitempty\""
		}{},
	}

	if len(c.OpensearchIPFilter) > 0 {
		db.IpFilter = &c.OpensearchIPFilter
	}

	if c.MaintenanceDOW != "" && c.MaintenanceTime != "" {
		db.Maintenance = &struct {
			Dow  oapi.CreateDbaasServiceOpensearchJSONBodyMaintenanceDow `json:"dow"`
			Time string                                                  `json:"time"`
		}{
			Dow:  oapi.CreateDbaasServiceOpensearchJSONBodyMaintenanceDow(c.MaintenanceDOW),
			Time: c.MaintenanceTime,
		}
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
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.OpensearchMaxIndexCount)) {
		db.MaxIndexCount = &c.OpensearchMaxIndexCount
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.OpensearchDashboardEnabled)) {
		db.OpensearchDashboards.Enabled = &c.OpensearchDashboardEnabled
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.OpensearchDashboardRequestTimeout)) {
		db.OpensearchDashboards.OpensearchRequestTimeout = &c.OpensearchDashboardRequestTimeout
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.OpensearchDashboardRequestTimeout)) {
		db.OpensearchDashboards.MaxOldSpaceSize = &c.OpensearchDashboardMaxOldSpaceSize
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.OpensearchIndexTemplateMappingNestedObjectsLimit)) {
		db.IndexTemplate.MappingNestedObjectsLimit = &c.OpensearchIndexTemplateMappingNestedObjectsLimit
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.OpensearchIndexTemplateNumberOfReplicas)) {
		db.IndexTemplate.NumberOfReplicas = &c.OpensearchIndexTemplateNumberOfReplicas
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.OpensearchIndexTemplateNumberOfShards)) {
		db.IndexTemplate.NumberOfShards = &c.OpensearchIndexTemplateNumberOfShards
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.OpensearchIndexPatterns)) {
		db.IndexPatterns = &[]struct {
			MaxIndexCount    *int64                                                                  `json:"max-index-count,omitempty"`
			Pattern          *string                                                                 `json:"pattern,omitempty"`
			SortingAlgorithm *oapi.CreateDbaasServiceOpensearchJSONBodyIndexPatternsSortingAlgorithm `json:"sorting-algorithm,omitempty"`
		}{}

		err := json.Unmarshal([]byte(c.OpensearchIndexPatterns), db.IndexPatterns)
		if err != nil {
			return fmt.Errorf("failed to decode Opensearch index patterns JSON: %s", err)
		}
	}

	var res *oapi.CreateDbaasServiceOpensearchResponse
	decorateAsyncOperation(fmt.Sprintf("Creating Database Service %q...", c.Name), func() {
		res, err = globalstate.EgoscaleClient.CreateDbaasServiceOpensearchWithResponse(ctx, oapi.DbaasServiceName(c.Name), db)
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
		}).showDatabaseServiceOpensearch(ctx))
	}

	return nil
}
