package cmd

import (
	"fmt"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

func (c *dbaasExternalEndpointCreateCmd) createOpensearch(_ *cobra.Command, _ []string) error {
	ctx := GContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	opensearchRequestPayload := v3.DBAASEndpointOpensearchInputCreate{
		Settings: &v3.DBAASEndpointOpensearchInputCreateSettings{},
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

	op, err := client.CreateDBAASExternalEndpointOpensearch(ctx, c.Name, opensearchRequestPayload)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Creating DBaaS OpenSearch external Endpoint %q", c.Name), func() {
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
			Type:               "opensearch",
		}).CmdRun(nil, nil)
	}
	return nil
}
