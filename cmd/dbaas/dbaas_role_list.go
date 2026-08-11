package dbaas

import (
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbaasRoleListOutput struct {
	Roles []dbaasRoleOutput `json:"roles"`
}

func (o *dbaasRoleListOutput) ToJSON() { output.JSON(o) }
func (o *dbaasRoleListOutput) ToText() { output.Text(o) }
func (o *dbaasRoleListOutput) ToTable() { output.Table(o) }

type dbaasRoleOutput struct {
	Name         string               `json:"name"`
	UUID         string               `json:"uuid,omitempty"`
	Privileges   []dbaasRolePrivOutput `json:"privileges,omitempty"`
	GrantedRoles []dbaasGrantedRoleOutput `json:"granted-roles,omitempty"`
}

type dbaasRolePrivOutput struct {
	Name          string `json:"name"`
	Database      string `json:"database,omitempty"`
	Table         string `json:"table,omitempty"`
	Column        string `json:"column,omitempty"`
	GrantOption   bool   `json:"grant-option,omitempty"`
	IsPartialRevoke bool `json:"is-partial-revoke,omitempty"`
}

type dbaasGrantedRoleOutput struct {
	Name            string `json:"name"`
	UUID            string `json:"uuid,omitempty"`
	IsDefault       bool   `json:"is-default,omitempty"`
	WithAdminOption bool  `json:"with-admin-option,omitempty"`
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
				Name:          string(p.Name),
				Database:      p.Database,
				Table:         p.Table,
				Column:        p.Column,
				GrantOption:   utilsDefaultBool(p.GrantOption, false),
				IsPartialRevoke: utilsDefaultBool(p.ISPartialRevoke, false),
			})
		}
		for _, g := range r.GrantedRoles {
			roleOut.GrantedRoles = append(roleOut.GrantedRoles, dbaasGrantedRoleOutput{
				Name:            g.Name,
				UUID:            g.Uuid,
				IsDefault:       utilsDefaultBool(g.ISDefault, false),
				WithAdminOption: utilsDefaultBool(g.WithAdminOption, false),
			})
		}
		out.Roles = append(out.Roles, roleOut)
	}

	return c.OutputFunc(out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasRoleCmd, &dbaasRoleListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}

// utilsDefaultBool is a local helper because utils.DefaultBool may not accept *bool
func utilsDefaultBool(v *bool, def bool) bool {
	if v == nil {
		return def
	}
	return *v
}