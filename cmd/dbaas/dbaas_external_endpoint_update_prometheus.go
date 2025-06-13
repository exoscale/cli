package cmd

import (
	"fmt"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

func (c *dbaasExternalEndpointUpdateCmd) updatePrometheus(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
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

	op, err := client.UpdateDBAASExternalEndpointPrometheus(ctx, v3.UUID(c.ID), prometheusRequestPayload)
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Updating DBaaS Prometheus external Endpoint %q", c.ID), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	endpointID := op.Reference.ID.String()
	if !globalstate.Quiet {
		return (&dbaasExternalEndpointShowCmd{
			cliCommandSettings: exocmd.DefaultCLICmdSettings(),
			EndpointID:         endpointID,
			Type:               "prometheus",
		}).CmdRun(nil, nil)
	}
	return nil
}
