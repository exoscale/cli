package dbaas

import (
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

func (c *dbaasUserResetCmd) resetClickhouse(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	req := v3.ResetDBAASClickhouseUserPasswordRequest{}
	if c.Password != "" {
		req.Password = v3.DBAASUserPassword(c.Password)
	}

	// ClickHouse reset is synchronous and returns secrets directly
	secrets, err := client.ResetDBAASClickhouseUserPassword(ctx, c.Name, c.Username, req)
	if err != nil {
		return err
	}

	return c.OutputFunc(&dbaasUserClickhouseSecretsOutput{
		Username: c.Username,
		Password: secrets.Password,
	}, nil)
}
