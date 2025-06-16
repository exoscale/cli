package dbaas

import (
	"fmt"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/spf13/cobra"
)

type dbaasUserDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

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
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasUserDeleteCmd) CmdRun(cmd *cobra.Command, args []string) error {

	ctx := exocmd.GContext
	db, err := dbaasGetV3(ctx, c.Name, c.Zone)
	if err != nil {
		return err
	}

	switch db.Type {
	case "mysql":
		return c.deleteMysql(cmd, args)
	case "kafka":
		return c.deleteKafka(cmd, args)
	case "pg":
		return c.deletePg(cmd, args)
	case "opensearch":
		return c.deleteOpensearch(cmd, args)
	case "valkey":
		return c.deleteValkey(cmd, args)
	default:
		return fmt.Errorf("deleting user unsupported for service of type %q", db.Type)
	}

}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasUserCmd, &dbaasUserDeleteCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
