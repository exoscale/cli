package cmd

import (
	"fmt"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

func (c *dbaasUserResetCmd) resetValkey(cmd *cobra.Command, _ []string) error {

	ctx := gContext

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
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

	decorateAsyncOperation(fmt.Sprintf("Resetting DBaaS user %q", c.Username), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return c.outputFunc((&dbaasUserShowCmd{
			Name:     c.Name,
			Zone:     c.Zone,
			Username: c.Username,
		}).showValkey(ctx))
	}

	return nil
}
