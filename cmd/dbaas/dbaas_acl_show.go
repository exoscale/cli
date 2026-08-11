package dbaas

import (
	"bytes"
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbaasAclShowOutput struct {
	Users []dbaasAclUserOutput `json:"users"`
}

func (o *dbaasAclShowOutput) ToJSON() { output.JSON(o) }
func (o *dbaasAclShowOutput) ToText() { output.Text(o) }

type dbaasAclUserOutput struct {
	Username   string                `json:"username"`
	Roles      []dbaasAclRoleOutput  `json:"roles"`
	Privileges []dbaasPrivOutput     `json:"privileges"`
}

type dbaasAclRoleOutput struct {
	Name            string `json:"name"`
	Default         bool   `json:"default,omitempty"`
	WithAdminOption bool   `json:"with-admin-option,omitempty"`
}

type dbaasPrivOutput struct {
	AccessType   string `json:"access-type"`
	Database     string `json:"database,omitempty"`
	Table        string `json:"table,omitempty"`
	Column       string `json:"column,omitempty"`
	GrantOption  bool   `json:"grant-option,omitempty"`
	PartialRevoke bool  `json:"partial-revoke,omitempty"`
}

type dbaasAclShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_  bool `cli-cmd:"show"`
	Name string `cli-arg:"#" cli-usage:"NAME"`
	Zone string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasAclShowCmd) CmdAliases() []string { return nil }
func (c *dbaasAclShowCmd) CmdShort() string     { return "Show ClickHouse ACL configuration" }
func (c *dbaasAclShowCmd) CmdLong() string      { return "Show the current ClickHouse ACL configuration for a DBaaS service." }

func (c *dbaasAclShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasAclShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	acl, err := client.GetDBAASClickhouseAclConfig(ctx, c.Name)
	if err != nil {
		return err
	}

	out := &dbaasAclShowOutput{}
	for _, u := range acl.Users {
		userOut := dbaasAclUserOutput{
			Username: string(u.Username),
		}
		for _, r := range u.Roles {
			userOut.Roles = append(userOut.Roles, dbaasAclRoleOutput{
				Name:            r.Name,
				Default:         utils.DefaultBool(r.Default, false),
				WithAdminOption: utils.DefaultBool(r.WithAdminOption, false),
			})
		}
		for _, p := range u.Privileges {
			userOut.Privileges = append(userOut.Privileges, dbaasPrivOutput{
				AccessType:  p.AccessType,
				Database:    p.Database,
				Table:       p.Table,
				Column:      p.Column,
				GrantOption: utils.DefaultBool(p.GrantOption, false),
				PartialRevoke: utils.DefaultBool(p.PartialRevoke, false),
			})
		}
		out.Users = append(out.Users, userOut)
	}

	return c.OutputFunc(out, nil)
}

func (o *dbaasAclShowOutput) ToTable() {
	t := table.NewTable(nil)
	defer t.Render()

	for _, u := range o.Users {
		t.Append([]string{"User", u.Username})

		// Roles
		buf := bytes.NewBuffer(nil)
		rolesTable := table.NewEmbeddedTable(buf)
		rolesTable.SetHeader([]string{"Role", "Default", "Admin"})
		for _, r := range u.Roles {
			rolesTable.Append([]string{
				r.Name,
				fmt.Sprintf("%v", r.Default),
				fmt.Sprintf("%v", r.WithAdminOption),
			})
		}
		rolesTable.Render()
		t.Append([]string{"Roles", buf.String()})

		// Privileges
		buf.Reset()
		privsTable := table.NewEmbeddedTable(buf)
		privsTable.SetHeader([]string{"Access", "Database", "Table", "Column", "Grant", "Partial"})
		for _, p := range u.Privileges {
			privsTable.Append([]string{
				p.AccessType,
				p.Database,
				p.Table,
				p.Column,
				fmt.Sprintf("%v", p.GrantOption),
				fmt.Sprintf("%v", p.PartialRevoke),
			})
		}
		privsTable.Render()
		t.Append([]string{"Privileges", buf.String()})
		t.Append([]string{"", ""})
	}
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasAclCmd, &dbaasAclShowCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}