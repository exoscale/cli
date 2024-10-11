package cmd

import (
	"fmt"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

func (c *dbaasExternalEndpointUpdateCmd) updateOpensearch(_ *cobra.Command, _ []string) error {
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	opensearchRequestPayload := v3.DBAASEndpointOpensearchInputUpdate{
		Settings: &v3.DBAASEndpointOpensearchInputUpdateSettings{},
	}

	if c.OpensearchURL != "" {
		opensearchRequestPayload.Settings.URL = c.OpensearchURL
	}
	if c.OpensearchIndexPrefix != "" {
		opensearchRequestPayload.Settings.IndexPrefix = c.OpensearchIndexPrefix
	}
	if c.OpensearchCA != "" {
		opensearchRequestPayload.Settings.CA = c.OpensearchCA
	}
	if c.OpensearchIndexDaysMax != 0 {
		opensearchRequestPayload.Settings.IndexDaysMax = c.OpensearchIndexDaysMax
	}
	if c.OpensearchTimeout != 0 {
		opensearchRequestPayload.Settings.Timeout = c.OpensearchTimeout
	}

	op, err := client.UpdateDBAASExternalEndpointOpensearch(ctx, v3.UUID(c.ID), opensearchRequestPayload)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Updating DBaaS OpenSearch external Endpoint %q", c.ID), func() {
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
			Type:               "opensearch",
		}).cmdRun(nil, nil)
	}
	return nil
}
