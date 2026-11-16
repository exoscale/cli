package user

import (
	"fmt"

	exocmd "github.com/exoscale/cli/cmd"
	dbaasutils "github.com/exoscale/cli/cmd/dbaas/internal"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/output"
	"github.com/spf13/cobra"
)

type dbaasUsersListItemOutput struct {
	Username string `json:"username,omitempty"`
	Type     string `json:"type,omitempty"`
}
type dbaasUsersListOutput []dbaasUsersListItemOutput

func (o *dbaasUsersListOutput) ToJSON() { output.JSON(o) }
func (o *dbaasUsersListOutput) ToText() { output.Text(o) }

func (o *dbaasUsersListOutput) ToTable() {
	output.Table(o)
}

type dbaasUserListCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Name string `cli-arg:"#"`
	Zone string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasUserListCmd) CmdAliases() []string { return nil }

func (c *dbaasUserListCmd) CmdShort() string { return "List users of a DBAAS service" }

func (c *dbaasUserListCmd) CmdLong() string {
	return `This command list users and their role for a specified DBAAS service.`
}

func (c *dbaasUserListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	clicmd.CmdSetZoneFlagFromDefault(cmd)

	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasUserListCmd) CmdRun(cmd *cobra.Command, args []string) error {

	ctx := exocmd.GContext
	db, err := dbaasutils.GetV3(ctx, c.Name, c.Zone)
	if err != nil {
		return err
	}

	switch db.Type {
	case "mysql":
		return c.listMysql()
	case "kafka":
		return c.listKafka()
	case "pg":
		return c.listPG()
	case "opensearch":
		return c.listOpensearch()
	case "grafana":
		return c.listGrafana()
	case "valkey":
		return c.listValkey()
	case "thanos":
		return c.listThanos()
	case "clickhouse":
		return c.listClickhouse()
	default:
		return fmt.Errorf("listing users unsupported for service of type %q", db.Type)
	}

}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &dbaasUserListCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
