package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
)

type datadogOutput struct {
	ID       string                                        `json:"id"`
	Name     string                                        `json:"name"`
	Type     string                                        `json:"type"`
	Settings v3.DBAASExternalEndpointDatadogOutputSettings `json:"settings"`
}

func (o *datadogOutput) ToJSON() { output.JSON(o) }

func (o *datadogOutput) ToText() { output.Text(o) }

func (o *datadogOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Datadog External Endpoint"})
	defer t.Render()

	t.Append([]string{"Endpoint ID", o.ID})
	t.Append([]string{"Endpoint Name", o.Name})
	t.Append([]string{"Endpoint Type", o.Type})

	settings := o.Settings
	t.Append([]string{"Site", string(settings.Site)})

	disableConsumerStats := "false"
	if settings.DisableConsumerStats != nil && *settings.DisableConsumerStats {
		disableConsumerStats = "true"
	}

	t.Append([]string{"Disable Consumer Stats", disableConsumerStats})
	t.Append([]string{"Kafka Consumer Check Instances", strconv.FormatInt(settings.KafkaConsumerCheckInstances, 10)})
	t.Append([]string{"Kafka Consumer Stats Timeout", strconv.FormatInt(settings.KafkaConsumerStatsTimeout, 10)})
	t.Append([]string{"Max Partition Contexts", strconv.FormatInt(settings.MaxPartitionContexts, 10)})

	if len(settings.DatadogTags) > 0 {
		tagLines := make([]string, len(settings.DatadogTags))
		for i, tag := range settings.DatadogTags {
			tagLines[i] = fmt.Sprintf("%s (%s)", tag.Tag, tag.Comment)
		}
		t.Append([]string{"Datadog Tags", tagLines[0]})
		for _, line := range tagLines[1:] {
			t.Append([]string{"", line})
		}
	} else {
		t.Append([]string{"Datadog Tags", "None"})
	}
}

func (c *dbaasExternalEndpointShowCmd) showDatadog() (output.Outputter, error) {
	ctx := GContext

	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return nil, err
	}

	endpointUUID, err := v3.ParseUUID(c.EndpointID)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint ID: %w", err)
	}
	endpointResponse, err := client.GetDBAASExternalEndpointDatadog(ctx, endpointUUID)
	if err != nil {
		return nil, fmt.Errorf("error getting Datadog external endpoint: %w", err)
	}

	output := &datadogOutput{
		ID:       endpointResponse.ID.String(),
		Name:     endpointResponse.Name,
		Type:     string(endpointResponse.Type),
		Settings: *endpointResponse.Settings,
	}

	return output, nil
}
