package dbaas

import (
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

func (c *dbaasUserListCmd) listClickhouse(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	users, err := client.ListDBAASClickhouseUsers(ctx, c.Name)
	if err != nil {
		return err
	}

	res := make(dbaasUsersListOutput, 0, len(users.Users))
	for _, u := range users.Users {
		res = append(res, dbaasUsersListItemOutput{
			Username: string(u.Username),
		})
	}

	return c.OutputFunc(&res, nil)
}
