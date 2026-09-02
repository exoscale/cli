package dbaas

import (
	"fmt"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

func (c *dbaasUserDeleteCmd) deleteClickhouse(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	users, err := client.ListDBAASClickhouseUsers(ctx, c.Name)
	if err != nil {
		return err
	}

	userUUID := ""
	for _, u := range users.Users {
		if string(u.Username) == c.Username {
			userUUID = string(u.Uuid)
			break
		}
	}
	if userUUID == "" {
		return fmt.Errorf("user %q not found for service %q", c.Username, c.Name)
	}

	if !c.Force {
		if !utils.AskQuestion(
			ctx,
			fmt.Sprintf(
				"Are you sure you want to delete user %q from service %q?",
				c.Username,
				c.Name,
			),
		) {
			return nil
		}
	}

	op, err := client.DeleteDBAASClickhouseUser(ctx, c.Name, v3.UUID(userUUID))
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Deleting user %q...", c.Username), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	return err
}
