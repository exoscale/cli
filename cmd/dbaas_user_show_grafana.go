package cmd

import (
	"fmt"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

func (c *dbaasUserShowCmd) showGrafana(cmd *cobra.Command, _ []string) (output.Outputter, error) {

	ctx := gContext

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return &dbaasUserShowOutput{}, err
	}

	s, err := client.GetDBAASServiceGrafana(ctx, c.Name)
	if err != nil {
		return &dbaasUserShowOutput{}, err
	}

	for _, u := range s.Users {

		if u.Username == c.Username {
			return &dbaasUserShowOutput{
				Username: c.Username,
				Password: u.Password,
				Type:     u.Type,
			}, nil
		}

	}

	return &dbaasUserShowOutput{}, fmt.Errorf("user %q not found for service %q", c.Username, c.Name)
}
