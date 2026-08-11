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

type dbaasRoleListOutput struct {
	Roles []dbaasRoleOutput `json:"roles"`
}

func (o *dbaasRoleListOutput) ToJSON() { output.JSON(o) }
func (o *dbaasRoleListOutput) ToText() { output.Text(o) }

type dbaasRoleOutput struct {
	Name         string                  `json:"name"`
	UUID         string                  `json:"uuid,omitempty"`
	Privileges   []dbaasRolePrivOutput   `json:"privileges,omitempty"`
	GrantedRoles []dbaasGrantedRoleOutput `json:"granted-roles,omitempty"`
}

type dbaasRolePrivOutput struct {
	Name            string `json:"name"`
	Database        string `json:"database,omitempty"`
	Table           string `json:"table,omitempty"`
	Column          string `json:"column,omitempty"`
	GrantOption     bool   `json:"grant-option,omitempty"`
	IsPartialRevoke bool   `json:"is-partial-revoke,omitempty"`
}

type dbaasGrantedRoleOutput struct {
	Name            string `json:"name"`
	UUID            string `json:"uuid,omitempty"`
	IsDefault       bool   `json:"is-default,omitempty"`
	WithAdminOption bool   `json:"with-admin-option,omitempty"`
}

type dbaasRoleListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_    bool `cli-cmd:"list"`
	Name string `cli-arg:"#" cli-usage:"NAME"`
	Zone string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasRoleListCmd) CmdAliases() []string { return nil }
func (c *dbaasRoleListCmd) CmdShort() string     { return "List ClickHouse roles" }
func (c *dbaasRoleListCmd) CmdLong() string      { return "List roles for a ClickHouse DBaaS service." }

func (c *dbaasRoleListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasRoleListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	roles, err := client.ListDBAASClickhouseRoles(ctx, c.Name)
	if err != nil {
		return err
	}

	out := &dbaasRoleListOutput{}
	for _, r := range roles.Roles {
		roleOut := dbaasRoleOutput{
			Name: string(r.Name),
			UUID: r.Uuid,
		}
		for _, p := range r.Privileges {
			roleOut.Privileges = append(roleOut.Privileges, dbaasRolePrivOutput{
				Name:            string(p.Name),
				Database:        p.Database,
				Table:           p.Table,
				Column:          p.Column,
				GrantOption:     utils.DefaultBool(p.GrantOption, false),
				IsPartialRevoke: utils.DefaultBool(p.ISPartialRevoke, false),
			})
		}
		for _, g := range r.GrantedRoles {
			roleOut.GrantedRoles = append(roleOut.GrantedRoles, dbaasGrantedRoleOutput{
				Name:            g.Name,
				UUID:            g.Uuid,
				IsDefault:       utils.DefaultBool(g.ISDefault, false),
				WithAdminOption: utils.DefaultBool(g.WithAdminOption, false),
			})
		}
		out.Roles = append(out.Roles, roleOut)
	}

	return c.OutputFunc(out, nil)
}

func (o *dbaasRoleListOutput) ToTable() {
	t := table.NewTable(nil)
	defer t.Render()

	if len(o.Roles) == 0 {
		t.Append([]string{"No roles found", ""})
		return
	}

	for _, r := range o.Roles {
		t.Append([]string{"Role", r.Name})
		t.Append([]string{"UUID", r.UUID})

		// Privileges as embedded table
		buf := bytes.NewBuffer(nil)
		privsTable := table.NewEmbeddedTable(buf)
		privsTable.SetHeader([]string{"Access", "Database", "Table", "Column", "Grant"})
		for _, p := range r.Privileges {
			privsTable.Append([]string{
				p.Name,
				p.Database,
				p.Table,
				p.Column,
				fmt.Sprintf("%v", p.GrantOption),
			})
		}
		privsTable.Render()
		t.Append([]string{"Privileges", buf.String()})

		// Granted roles as embedded table
		buf.Reset()
		grantedTable := table.NewEmbeddedTable(buf)
		grantedTable.SetHeader([]string{"Role", "Default", "Admin"})
		for _, g := range r.GrantedRoles {
			grantedTable.Append([]string{
				g.Name,
				fmt.Sprintf("%v", g.IsDefault),
				fmt.Sprintf("%v", g.WithAdminOption),
			})
		}
		grantedTable.Render()
		t.Append([]string{"Granted Roles", buf.String()})

		t.Append([]string{"", ""})
	}
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasRoleCmd, &dbaasRoleListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
