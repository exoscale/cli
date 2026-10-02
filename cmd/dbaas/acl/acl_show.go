package acl

import (
	"bytes"
	"context"
	"fmt"
	"os"

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
	Username   string               `json:"username"`
	Roles      []dbaasAclRoleOutput `json:"roles"`
	Privileges []dbaasPrivOutput    `json:"privileges"`
}

type dbaasAclRoleOutput struct {
	Name            string `json:"name"`
	Default         bool   `json:"default,omitempty"`
	WithAdminOption bool   `json:"with-admin-option,omitempty"`
}

type dbaasPrivOutput struct {
	AccessType    string `json:"access-type"`
	Database      string `json:"database,omitempty"`
	Table         string `json:"table,omitempty"`
	Column        string `json:"column,omitempty"`
	GrantOption   bool   `json:"grant-option,omitempty"`
	PartialRevoke bool   `json:"partial-revoke,omitempty"`
}

type dbaasAclShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_    bool   `cli-cmd:"show"`
	Name string `cli-arg:"#" cli-usage:"NAME"`
	Zone string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasAclShowCmd) CmdAliases() []string { return nil }
func (c *dbaasAclShowCmd) CmdShort() string {
	return "Show the ACL configuration of a Database Service"
}
func (c *dbaasAclShowCmd) CmdLong() string {
	return "Show the current ACL configuration of a ClickHouse, Kafka or OpenSearch DBaaS service."
}

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

	dbType, err := dbaasServiceType(ctx, client, c.Name, c.Zone)
	if err != nil {
		return err
	}

	switch dbType {
	case "clickhouse":
		return c.OutputFunc(showClickhouseAcl(ctx, client, c.Name))
	case "kafka":
		return c.OutputFunc(listAcl(ctx, client, dbType, c.Name))
	case "opensearch":
		return c.OutputFunc(showOpensearchAcl(ctx, client, c.Name))
	default:
		return fmt.Errorf("showing ACL configuration unsupported for service of type %q", dbType)
	}
}

func showClickhouseAcl(ctx context.Context, client *v3.Client, name string) (output.Outputter, error) {
	acl, err := client.GetDBAASClickhouseAclConfig(ctx, name)
	if err != nil {
		return nil, err
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
				AccessType:    p.AccessType,
				Database:      p.Database,
				Table:         p.Table,
				Column:        p.Column,
				GrantOption:   utils.DefaultBool(p.GrantOption, false),
				PartialRevoke: utils.DefaultBool(p.PartialRevoke, false),
			})
		}
		out.Users = append(out.Users, userOut)
	}

	return out, nil
}

func (o *dbaasAclShowOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()

	if len(o.Users) == 0 {
		t.Append([]string{"No ACL configuration found", ""})
		return
	}

	for _, u := range o.Users {
		t.Append([]string{"User", u.Username})

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

type dbaasAclOpensearchShowOutput struct {
	ACLEnabled         bool               `json:"acl-enabled"`
	ExtendedACLEnabled bool               `json:"extended-acl-enabled"`
	ACL                dbaasAclListOutput `json:"acl"`
}

func (o *dbaasAclOpensearchShowOutput) ToJSON() { output.JSON(o) }
func (o *dbaasAclOpensearchShowOutput) ToText() { output.Text(o) }
func (o *dbaasAclOpensearchShowOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()

	t.Append([]string{"ACL Enabled", fmt.Sprint(o.ACLEnabled)})
	t.Append([]string{"Extended ACL Enabled", fmt.Sprint(o.ExtendedACLEnabled)})

	buf := bytes.NewBuffer(nil)
	at := table.NewEmbeddedTable(buf)
	at.SetHeader([]string{"Username", "Index", "Permission"})
	for _, acl := range o.ACL {
		at.Append([]string{acl.Username, acl.Resource, acl.Permission})
	}
	at.Render()
	t.Append([]string{"ACL", buf.String()})
}

func showOpensearchAcl(ctx context.Context, client *v3.Client, name string) (output.Outputter, error) {
	config, err := client.GetDBAASOpensearchAclConfig(ctx, name)
	if err != nil {
		return nil, err
	}

	return &dbaasAclOpensearchShowOutput{
		ACLEnabled:         utils.DefaultBool(config.AclEnabled, false),
		ExtendedACLEnabled: utils.DefaultBool(config.ExtendedAclEnabled, false),
		ACL:                *opensearchAclListOutput(config),
	}, nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &dbaasAclShowCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
