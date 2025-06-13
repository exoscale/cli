package cmd

import (
	"fmt"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

func (c *dbaasUserCreateCmd) createPg(cmd *cobra.Command, _ []string) error {

	ctx := exocmd.GContext

	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	s, err := client.GetDBAASServicePG(ctx, c.Name)
	if err != nil {
		return err
	}

	if len(s.Users) == 0 {
		return fmt.Errorf("service %q is not ready for user creation", c.Name)
	}

	req := v3.CreateDBAASPostgresUserRequest{Username: v3.DBAASUserUsername(c.Username), AllowReplication: &c.PostgresAllowReplication}

	op, err := client.CreateDBAASPostgresUser(ctx, c.Name, req)

	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Creating DBaaS user %q", c.Username), func() {
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
		}).showPG(ctx))
	}

	return nil
}
