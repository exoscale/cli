package cmd

import (
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

func (c *dbaasUserListCmd) listKafka(cmd *cobra.Command, _ []string) error {

	ctx := GContext

	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	s, err := client.GetDBAASServiceKafka(ctx, c.Name)
	if err != nil {
		return err
	}

	res := make(dbaasUsersListOutput, 0)

	for _, u := range s.Users {
		res = append(res, dbaasUsersListItemOutput{
			Username: u.Username,
			Type:     u.Type,
		})
	}

	return c.OutputFunc(&res, nil)
}
