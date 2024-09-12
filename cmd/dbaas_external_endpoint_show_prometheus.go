package cmd

import (
	"fmt"
	"os"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
)

type prometheusOutput struct {
	ID       string                    `json:"id"`
	Name     string                    `json:"name"`
	Type     string                    `json:"type"`
	Settings v3.DBAASEndpointPrometheus `json:"settings"`
}

func (o *prometheusOutput) ToJSON() { output.JSON(o) }
func (o *prometheusOutput) ToText() { output.Text(o) }
func (o *prometheusOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Prometheus External Endpoint"})
	defer t.Render()

	t.Append([]string{"Endpoint ID", o.ID})
	t.Append([]string{"Endpoint Name", o.Name})
	t.Append([]string{"Endpoint Type", o.Type})

	settings := o.Settings
	t.Append([]string{"Basic Auth Username", settings.BasicAuthUsername})
	t.Append([]string{"Basic Auth Password", settings.BasicAuthPassword})
}

func (c *dbaasExternalEndpointShowCmd) showPrometheus() (output.Outputter, error) {
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return nil, err
	}

	endpointUUID, err := v3.ParseUUID(c.EndpointID)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint ID: %w", err)
	}

	endpointResponse, err := client.GetDBAASExternalEndpointPrometheus(ctx, endpointUUID)
	if err != nil {
		return nil, fmt.Errorf("error getting Prometheus external endpoint: %w", err)
	}

	output := &prometheusOutput{
		ID:       endpointResponse.ID.String(),
		Name:     endpointResponse.Name,
		Type:     string(endpointResponse.Type),
		Settings: *endpointResponse.Settings,
	}

	return output, nil
}
