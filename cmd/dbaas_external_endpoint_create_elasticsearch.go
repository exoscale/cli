package cmd

import (
	"fmt"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

func (c *dbaasExternalEndpointCreateCmd) createElasticsearch(_ *cobra.Command, _ []string) error {
	ctx := GContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	elasticsearchRequestPayload := v3.DBAASEndpointElasticsearchInputCreate{
		Settings: &v3.DBAASEndpointElasticsearchInputCreateSettings{},
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

	op, err := client.CreateDBAASExternalEndpointElasticsearch(ctx, c.Name, elasticsearchRequestPayload)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Creating DBaaS ElasticSearch external Endpoint %q", c.Name), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	endpointID := op.Reference.ID.String()
	if !globalstate.Quiet {
		return (&dbaasExternalEndpointShowCmd{
			CliCommandSettings: DefaultCLICmdSettings(),
			EndpointID:         endpointID,
			Type:               "elasticsearch",
		}).CmdRun(nil, nil)
	}
	return nil
}
