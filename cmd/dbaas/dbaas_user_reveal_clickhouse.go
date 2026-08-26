package dbaas

import (
	"context"
	"fmt"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

func (c *dbaasUserRevealCmd) revealClickhouse(ctx context.Context) (output.Outputter, error) {
	if c.Username != "avnadmin" {
		return nil, fmt.Errorf("reveal password is only supported for the avnadmin user")
	}

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return &dbaasUserRevealOutput{}, err
	}

	s, err := client.RevealDBAASClickhouseUserPassword(ctx, c.Name, c.Username)
	if err != nil {
		return &dbaasUserRevealOutput{}, err
	}

	return &dbaasUserRevealOutput{
		Username: c.Username,
		Password: s.Password,
	}, nil
}
