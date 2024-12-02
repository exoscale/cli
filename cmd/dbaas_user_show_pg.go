package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbaasPGUserShowOutput struct {
	AllowReplication *bool `json:"allow-replication,omitempty"`
}

func (o *dbaasPGUserShowOutput) formatUser(t *table.Table) {
	if o.AllowReplication != nil {
		t.Append([]string{"Allow Replication", strconv.FormatBool(*o.AllowReplication)})
	}
}

func (c *dbaasUserShowCmd) showPG(ctx context.Context) (output.Outputter, error) {

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return &dbaasUserShowOutput{}, err
	}

	s, err := client.GetDBAASServicePG(ctx, c.Name)
	if err != nil {
		return &dbaasUserShowOutput{}, err
	}

	for _, u := range s.Users {

		if u.Username == c.Username {
			return &dbaasUserShowOutput{
				Username: c.Username,
				Type:     u.Type,
				PG: &dbaasPGUserShowOutput{
					AllowReplication: u.AllowReplication,
				},
			}, nil
		}

	}

	return &dbaasUserShowOutput{}, fmt.Errorf("user %q not found for service %q", c.Username, c.Name)
}
