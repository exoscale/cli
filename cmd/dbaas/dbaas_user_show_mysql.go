package cmd

import (
	"context"
	"fmt"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbaasMysqlUserShowOutput struct {
	Authentication string `json:"authentication,omitempty"`
}

func (o *dbaasMysqlUserShowOutput) formatUser(t *table.Table) {
	t.Append([]string{"Authentication", o.Authentication})
}

func (c *dbaasUserShowCmd) showMysql(ctx context.Context) (output.Outputter, error) {

	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return &dbaasUserShowOutput{}, err
	}

	s, err := client.GetDBAASServiceMysql(ctx, c.Name)
	if err != nil {
		return &dbaasUserShowOutput{}, err
	}

	for _, u := range s.Users {

		if u.Username == c.Username {
			return &dbaasUserShowOutput{
				Username: c.Username,

				Type: u.Type,
				MySQL: &dbaasMysqlUserShowOutput{
					Authentication: u.Authentication,
				},
			}, nil
		}

	}

	return &dbaasUserShowOutput{}, fmt.Errorf("user %q not found for service %q", c.Username, c.Name)
}
