package cmd

import (
	"fmt"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

func (c *dbaasUserDeleteCmd) deleteOpensearch(cmd *cobra.Command, _ []string) error {

	ctx := GContext

	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	s, err := client.GetDBAASServiceOpensearch(ctx, c.Name)
	if err != nil {
		return err
	}
	userFound := false
	for _, u := range s.Users {
		if u.Username == c.Username {
			userFound = true
			break
		}
	}
	if !userFound {
		return fmt.Errorf("user %q not found for service %q", c.Username, c.Name)
	}
	if !c.Force {
		if !askQuestion(fmt.Sprintf(
			"Are you sure you want to delete user %q", c.Username)) {
			return nil
		}
	}

	op, err := client.DeleteDBAASOpensearchUser(ctx, c.Name, c.Username)

	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Deletng DBaaS user %q", c.Username), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return c.OutputFunc((&dbaasServiceShowCmd{
			Name: c.Name,
			Zone: c.Zone,
		}).showDatabaseServiceOpensearch(ctx))
	}

	return nil

}
