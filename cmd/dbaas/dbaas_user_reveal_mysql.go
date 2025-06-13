package cmd

import (
	"context"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

func (c *dbaasUserRevealCmd) revealMysql(ctx context.Context) (output.Outputter, error) {

	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return &dbaasUserRevealOutput{}, err
	}

	s, err := client.RevealDBAASMysqlUserPassword(ctx, c.Name, c.Username)
	if err != nil {
		return &dbaasUserRevealOutput{}, err
	}

	return &dbaasUserRevealOutput{
		Password: s.Password,
	}, nil

}
