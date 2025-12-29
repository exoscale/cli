package dbaas

import (
	"context"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

func (c *dbaasUserRevealCmd) revealThanos(ctx context.Context) (output.Outputter, error) {

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return &dbaasUserRevealOutput{}, err
	}

	s, err := client.RevealDBAASThanosUserPassword(ctx, c.Name, c.Username)
	if err != nil {
		return &dbaasUserRevealOutput{}, err
	}

	return &dbaasUserRevealOutput{
		Password: s.Password,
	}, nil

}
