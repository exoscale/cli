package cmd

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbaasRedisUserShowOutput struct {
	AccessControl v3.DBAASServiceRedisUsersAccessControl `json:"access-control,omitempty"`
}

func (o *dbaasRedisUserShowOutput) formatUser(t *table.Table) {
	t.Append([]string{"Access Control", func() string {

		buf := bytes.NewBuffer(nil)
		ct := table.NewEmbeddedTable(buf)

		ct.Append([]string{"Categories", strings.Join(o.AccessControl.Categories, ",")})
		ct.Append([]string{"Channels", strings.Join(o.AccessControl.Channels, ",")})
		ct.Append([]string{"Commands", strings.Join(o.AccessControl.Commands, ",")})
		ct.Append([]string{"Keys", strings.Join(o.AccessControl.Keys, ",")})

		ct.Render()

		return buf.String()
	}()})
}

func (c *dbaasUserShowCmd) showRedis(ctx context.Context) (output.Outputter, error) {

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return &dbaasUserShowOutput{}, err
	}

	s, err := client.GetDBAASServiceRedis(ctx, c.Name)
	if err != nil {
		return &dbaasUserShowOutput{}, err
	}

	for _, u := range s.Users {

		if u.Username == c.Username {
			o := &dbaasUserShowOutput{
				Username: c.Username,
				Password: u.Password,
				Type:     u.Type,
			}

			if u.AccessControl != nil {
				o.Redis = &dbaasRedisUserShowOutput{
					AccessControl: *u.AccessControl,
				}
			}

			return o, nil
		}

	}

	return &dbaasUserShowOutput{}, fmt.Errorf("user %q not found for service %q", c.Username, c.Name)
}
