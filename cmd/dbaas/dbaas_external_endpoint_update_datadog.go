package dbaas

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

func (c *dbaasExternalEndpointUpdateCmd) updateDatadog(cmd *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	var datadogTags []v3.DBAASDatadogTag
	if c.DatadogTags != "" {
		if err := json.Unmarshal([]byte(c.DatadogTags), &datadogTags); err != nil {
			return fmt.Errorf("failed to parse datadog tags: %v", err)
		}
	}

	datadogRequestPayload := v3.DBAASEndpointDatadogInputUpdate{
		Settings: &v3.DBAASEndpointDatadogInputUpdateSettings{},
	}

	if c.DatadogAPIKey != "" {
		datadogRequestPayload.Settings.DatadogAPIKey = c.DatadogAPIKey
	}
	if c.DatadogSite != "" {
		datadogRequestPayload.Settings.Site = v3.EnumDatadogSite(c.DatadogSite)
	}
	if c.DatadogTags != "" {
		datadogRequestPayload.Settings.DatadogTags = datadogTags
	}
	if cmd.Flags().Changed("datadog-disable-consumer-stats") {
		datadogRequestPayload.Settings.DisableConsumerStats = v3.Bool(c.DatadogDisableConsumerStats)
	}
	if c.DatadogKafkaConsumerCheckInstances != 0 {
		datadogRequestPayload.Settings.KafkaConsumerCheckInstances = int64(c.DatadogKafkaConsumerCheckInstances)
	}
	if c.DatadogKafkaConsumerStatsTimeout != 0 {
		datadogRequestPayload.Settings.KafkaConsumerStatsTimeout = int64(c.DatadogKafkaConsumerStatsTimeout)
	}
	if c.DatadogMaxPartitionContexts != 0 {
		datadogRequestPayload.Settings.MaxPartitionContexts = int64(c.DatadogMaxPartitionContexts)
	}

	op, err := client.UpdateDBAASExternalEndpointDatadog(ctx, v3.UUID(c.ID), datadogRequestPayload)

	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Updating DBaaS Datadog external Endpoint %q", c.ID), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	if err != nil {
		return err
	}

	endpointID := op.Reference.ID.String()

	if !globalstate.Quiet {
		return (&dbaasExternalEndpointShowCmd{
			CliCommandSettings: exocmd.DefaultCLICmdSettings(),
			EndpointID:         endpointID,
			Type:               "datadog",
		}).CmdRun(nil, nil)
	}

	return nil
}
