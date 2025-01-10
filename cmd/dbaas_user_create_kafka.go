package cmd

import (
	"fmt"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

func (c *dbaasUserCreateCmd) createKafka(cmd *cobra.Command, _ []string) error {

	ctx := gContext

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
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

	req := v3.CreateDBAASKafkaUserRequest{Username: v3.DBAASUserUsername(c.Username)}

	op, err := client.CreateDBAASKafkaUser(ctx, c.Name, req)

	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Creating DBaaS user %q", c.Username), func() {
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
		}).showKafka(ctx))

	}

	return nil

}
