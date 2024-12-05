package cmd

import (
	"fmt"

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
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Name string `cli-arg:"#"`
	Zone string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasUserListCmd) cmdAliases() []string { return nil }

func (c *dbaasUserListCmd) cmdShort() string { return "List users of a DBAAS service" }

func (c *dbaasUserListCmd) cmdLong() string {
	return `This command list users and their role for a specified DBAAS service.`
}

func (c *dbaasUserListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)

	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasUserListCmd) cmdRun(cmd *cobra.Command, args []string) error {

	ctx := gContext
	db, err := dbaasGetV3(ctx, c.Name, c.Zone)
	if err != nil {
		return err
	}

	switch db.Type {
	case "mysql":
		return c.listMysql(cmd, args)
	case "kafka":
		return c.listKafka(cmd, args)
	case "pg":
		return c.listPG(cmd, args)
	case "opensearch":
		return c.listOpensearch(cmd, args)
	case "grafana":
		return c.listGrafana(cmd, args)
	default:
		return fmt.Errorf("listing users unsupported for service of type %q", db.Type)

	}

}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasUserCmd, &dbaasUserListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
