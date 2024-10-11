package cmd

import (
	"fmt"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

func (c *dbaasExternalEndpointUpdateCmd) updateElasticsearch(_ *cobra.Command, _ []string) error {
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	elasticsearchRequestPayload := v3.DBAASEndpointElasticsearchInputUpdate{
		Settings: &v3.DBAASEndpointElasticsearchInputUpdateSettings{},
	}

	if c.ElasticsearchURL != "" {
		elasticsearchRequestPayload.Settings.URL = c.ElasticsearchURL
	}
	if c.ElasticsearchIndexPrefix != "" {
		elasticsearchRequestPayload.Settings.IndexPrefix = c.ElasticsearchIndexPrefix
	}
	if c.ElasticsearchCA != "" {
		elasticsearchRequestPayload.Settings.CA = c.ElasticsearchCA
	}
	if c.ElasticsearchIndexDaysMax != 0 {
		elasticsearchRequestPayload.Settings.IndexDaysMax = c.ElasticsearchIndexDaysMax
	}
	if c.ElasticsearchTimeout != 0 {
		elasticsearchRequestPayload.Settings.Timeout = c.ElasticsearchTimeout
	}

	op, err := client.UpdateDBAASExternalEndpointElasticsearch(ctx, v3.UUID(c.ID), elasticsearchRequestPayload)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Updating DBaaS ElasticSearch external Endpoint %q", c.ID), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	endpointID := op.Reference.ID.String()
	if !globalstate.Quiet {
		return (&dbaasExternalEndpointShowCmd{
			cliCommandSettings: defaultCLICmdSettings(),
			EndpointID:         endpointID,
			Type:               "elasticsearch",
		}).cmdRun(nil, nil)
	}
	return nil
}
