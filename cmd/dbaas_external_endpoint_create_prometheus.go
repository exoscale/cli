package cmd

import (
	"fmt"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

func (c *dbaasExternalEndpointCreateCmd) createPrometheus(_ *cobra.Command, _ []string) error {
	ctx := GContext
	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	prometheusRequestPayload := v3.DBAASEndpointPrometheusPayload{
		Settings: &v3.DBAASEndpointPrometheusPayloadSettings{},
	}

	if c.PrometheusBasicAuthPassword != "" {
		prometheusRequestPayload.Settings.BasicAuthPassword = c.PrometheusBasicAuthPassword
	}
	if c.PrometheusBasicAuthUsername != "" {
		prometheusRequestPayload.Settings.BasicAuthUsername = c.PrometheusBasicAuthUsername
	}

	op, err := client.CreateDBAASExternalEndpointPrometheus(ctx, c.Name, prometheusRequestPayload)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Creating DBaaS Prometheus external Endpoint %q", c.Name), func() {
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
			Type:               "prometheus",
		}).CmdRun(nil, nil)
	}
	return nil
}
