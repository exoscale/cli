package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

func (c *dbaasExternalEndpointCreateCmd) createDatadog(_ *cobra.Command, _ []string) error {
	ctx := gContext

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	var datadogTags []v3.DBAASDatadogTag
	if c.DatadogTags != "" {
		if err := json.Unmarshal([]byte(c.DatadogTags), &datadogTags); err != nil {
			return fmt.Errorf("failed to parse datadog tags: %v", err)
		}
	}

	datadogRequestPayload := v3.CreateDBAASExternalEndpointDatadogRequest{
		Settings: &v3.DBAASEndpointDatadog{},
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
	if c.DatadogDisableConsumerStats {
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

	op, err := client.CreateDBAASExternalEndpointDatadog(ctx, c.Name, datadogRequestPayload)

	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Creating DBaaS Datadog external Endpoint %q", c.Name), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	if err != nil {
		return err
	}

	// endpointID := op.Reference.ID.String()

	// if !globalstate.Quiet {
	// 	return (&dbaasExternalEndpointShowCmd{
	// 		cliCommandSettings: defaultCLICmdSettings(),
	// 		ExternalEndpointID: endpointID,
	// 	}).cmdRun(nil, nil)
	// }

	return nil
}
