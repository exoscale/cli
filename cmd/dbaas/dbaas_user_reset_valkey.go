package dbaas

import (
	"fmt"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

func (c *dbaasUserResetCmd) resetValkey(cmd *cobra.Command, _ []string) error {

	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	req := v3.ResetDBAASValkeyUserPasswordRequest{}
	if c.Password != "" {
		req.Password = v3.DBAASUserPassword(c.Password)
	}

	op, err := client.ResetDBAASValkeyUserPassword(ctx, c.Name, c.Username, req)

	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Resetting DBaaS user %q", c.Username), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return c.OutputFunc((&dbaasUserShowCmd{
			Name:     c.Name,
			Zone:     c.Zone,
			Username: c.Username,
		}).showValkey(ctx))
	}

	return nil
}
