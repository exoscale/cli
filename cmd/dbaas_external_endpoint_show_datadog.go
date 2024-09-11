package cmd

import (
	"fmt"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type datadogOutput struct {
	output.Outputter
	ID       string                  `json:"id"`
	Name     string                  `json:"name"`
	Type     string                  `json:"type"`
	// Settings v3.DBAASEndpointDatadog `json:"settings"`
}

func (c *dbaasExternalEndpointShowCmd) showDatadog() (output.Outputter, error) {
	ctx := gContext

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
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
		// Settings: *endpointResponse.Settings,
	}

	return output, nil
}
