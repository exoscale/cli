package user

import (
	"fmt"

	exocmd "github.com/exoscale/cli/cmd"
	dbaasutils "github.com/exoscale/cli/cmd/dbaas/internal"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/spf13/cobra"
)

type dbaasUserDeleteCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Name     string `cli-arg:"#"`
	Username string `cli-arg:"#"`
	Zone     string `cli-short:"z" cli-usage:"Database Service zone"`

	Force bool `cli-short:"f" cli-usage:"don't prompt for confirmation"`
}

func (c *dbaasUserDeleteCmd) CmdAliases() []string { return nil }

func (c *dbaasUserDeleteCmd) CmdShort() string { return "Delete DBAAS user" }

func (c *dbaasUserDeleteCmd) CmdLong() string {
	return `This command deletes a DBAAS user for the specified service.`
}

func (c *dbaasUserDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	clicmd.CmdSetZoneFlagFromDefault(cmd)
	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasUserDeleteCmd) CmdRun(cmd *cobra.Command, args []string) error {

	ctx := exocmd.GContext
	db, err := dbaasutils.GetV3(ctx, c.Name, c.Zone)
	if err != nil {
		return err
	}

	switch db.Type {
	case "mysql":
		return c.deleteMysql()
	case "kafka":
		return c.deleteKafka()
	case "pg":
		return c.deletePg()
	case "opensearch":
		return c.deleteOpensearch()
	case "valkey":
		return c.deleteValkey()
	case "clickhouse":
		return c.deleteClickhouse()
	default:
		return fmt.Errorf("deleting user unsupported for service of type %q", db.Type)
	}

}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &dbaasUserDeleteCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
