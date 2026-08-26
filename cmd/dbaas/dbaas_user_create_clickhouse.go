package dbaas

import (
	"fmt"
	"os"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

// dbaasUserClickhouseSecretsOutput carries the user/password pair returned
// synchronously by ClickHouse user creation and password reset.
type dbaasUserClickhouseSecretsOutput struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

func (o *dbaasUserClickhouseSecretsOutput) ToJSON() { output.JSON(o) }
func (o *dbaasUserClickhouseSecretsOutput) ToText() { output.Text(o) }

func (o *dbaasUserClickhouseSecretsOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()

	t.Append([]string{"Username", o.Username})
	t.Append([]string{"Password", o.Password})
}

func (c *dbaasUserCreateCmd) createClickhouse(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	s, err := client.GetDBAASServiceClickhouse(ctx, c.Name)
	if err != nil {
		return err
	}
	if len(s.Users) == 0 {
		return fmt.Errorf("service %q is not ready for user creation", c.Name)
	}

	req := v3.CreateDBAASClickhouseUserRequest{
		Username: v3.DBAASUserUsername(c.Username),
	}

	// ClickHouse user creation is synchronous and returns secrets directly
	secrets, err := client.CreateDBAASClickhouseUser(ctx, c.Name, req)
	if err != nil {
		return err
	}

	return c.OutputFunc(&dbaasUserClickhouseSecretsOutput{
		Username: c.Username,
		Password: secrets.Password,
	}, nil)
}
