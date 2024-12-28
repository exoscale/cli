package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type dbaasUserDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Name     string `cli-arg:"#"`
	Username string `cli-arg:"#"`
	Zone     string `cli-short:"z" cli-usage:"Database Service zone"`

	Force bool `cli-short:"f" cli-usage:"don't prompt for confirmation"`
}

func (c *dbaasUserDeleteCmd) cmdAliases() []string { return nil }

func (c *dbaasUserDeleteCmd) cmdShort() string { return "Delete DBAAS user" }

func (c *dbaasUserDeleteCmd) cmdLong() string {
	return `This command deletes a DBAAS user for the specified service.`
}

func (c *dbaasUserDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasUserDeleteCmd) cmdRun(cmd *cobra.Command, args []string) error {

	ctx := gContext
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
	default:
		return fmt.Errorf("deleting user unsupported for service of type %q", db.Type)
	}

}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasUserCmd, &dbaasUserDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
