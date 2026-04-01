package dbaas

import (
	"fmt"
	"os"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/spf13/cobra"
)

type dbaasUserShowOutput struct {
	Username string `json:"username,omitempty"`
	Type     string `json:"type,omitempty"`

	// Additional user info for some DBAAS Services
	MySQL *dbaasMysqlUserShowOutput `json:"mysql,omitempty"`
	PG    *dbaasPGUserShowOutput    `json:"pg,omitempty"`
}

func (o *dbaasUserShowOutput) ToJSON() { output.JSON(o) }
func (o *dbaasUserShowOutput) ToText() { output.Text(o) }

func (o *dbaasUserShowOutput) ToTable() {

	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Service User"})
	defer t.Render()

	t.Append([]string{"Username", o.Username})
	t.Append([]string{"Type", o.Type})

	switch {
	case o.MySQL != nil:
		o.MySQL.formatUser(t)
	case o.PG != nil:
		o.PG.formatUser(t)
	}

}

type dbaasUserShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Name     string `cli-arg:"#"`
	Username string `cli-arg:"#"`
	Zone     string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasUserShowCmd) CmdAliases() []string { return nil }

func (c *dbaasUserShowCmd) CmdShort() string { return "Show the details of a user" }

func (c *dbaasUserShowCmd) CmdLong() string {
	return `This command show a user and their details for a specified DBAAS service.`
}

func (c *dbaasUserShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)

	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasUserShowCmd) CmdRun(cmd *cobra.Command, args []string) error {

	ctx := exocmd.GContext
	db, err := dbaasGetV3(ctx, c.Name, c.Zone)
	if err != nil {
		return err
	}

	switch db.Type {
	case "mysql":
		return c.OutputFunc(c.showMysql(ctx))
	case "kafka":
		return c.OutputFunc(c.showKafka(ctx))
	case "pg":
		return c.OutputFunc(c.showPG(ctx))
	case "opensearch":
		return c.OutputFunc(c.showOpensearch(ctx))
	case "grafana":
		return c.OutputFunc(c.showGrafana(ctx))
	case "valkey":
		return c.OutputFunc(c.showValkey(ctx))
	case "thanos":
		return c.OutputFunc(c.showThanos(ctx))
	default:
		return fmt.Errorf("listing users unsupported for service of type %q", db.Type)

	}

}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasUserCmd, &dbaasUserShowCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
