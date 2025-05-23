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

type elasticsearchOutput struct {
	ID       string                                      `json:"id"`
	Name     string                                      `json:"name"`
	Type     string                                      `json:"type"`
	Settings v3.DBAASEndpointElasticsearchOptionalFields `json:"settings"`
}

func (o *elasticsearchOutput) ToJSON() { output.JSON(o) }
func (o *elasticsearchOutput) ToText() { output.Text(o) }
func (o *elasticsearchOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Elasticsearch External Endpoint"})
	defer t.Render()

	t.Append([]string{"Endpoint ID", o.ID})
	t.Append([]string{"Endpoint Name", o.Name})
	t.Append([]string{"Endpoint Type", o.Type})

	settings := o.Settings
	t.Append([]string{"Elasticsearch URL", settings.URL})
	t.Append([]string{"Index Prefix", settings.IndexPrefix})
	t.Append([]string{"Index Days Max", strconv.FormatInt(settings.IndexDaysMax, 10)})
	t.Append([]string{"Timeout", strconv.FormatInt(settings.Timeout, 10)})
}

func (c *dbaasExternalEndpointShowCmd) showElasticsearch() (output.Outputter, error) {
	ctx := GContext
	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return nil, err
	}

	endpointUUID, err := v3.ParseUUID(c.EndpointID)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint ID: %w", err)
	}

	endpointResponse, err := client.GetDBAASExternalEndpointElasticsearch(ctx, endpointUUID)
	if err != nil {
		return nil, fmt.Errorf("error getting ElasticSearch external endpoint: %w", err)
	}

	output := &elasticsearchOutput{
		ID:       endpointResponse.ID.String(),
		Name:     endpointResponse.Name,
		Type:     string(endpointResponse.Type),
		Settings: *endpointResponse.Settings,
	}

	return output, nil
}
