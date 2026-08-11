// AI-modified by hermes-agent - not reviewed yet
package dbaas

import (
	"context"
	"fmt"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

func (c *dbaasUserShowCmd) showClickhouse(ctx context.Context) (output.Outputter, error) {
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return &dbaasUserShowOutput{}, err
	}

	s, err := client.GetDBAASServiceClickhouse(ctx, c.Name)
	if err != nil {
		return &dbaasUserShowOutput{}, err
	}

	for _, u := range s.Users {
		if string(u.Username) == c.Username {
			return &dbaasUserShowOutput{
				Username: c.Username,
			}, nil
		}
	}

	return &dbaasUserShowOutput{}, fmt.Errorf("user %q not found for service %q", c.Username, c.Name)
}