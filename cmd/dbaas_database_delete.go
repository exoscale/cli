package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type dbaasDatabaseDeleteCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Name     string `cli-arg:"#"`
	Database string `cli-arg:"#"`

	Zone string `cli-short:"z" cli-usage:"Database Service zone"`

	Force bool `cli-short:"f" cli-usage:"don't prompt for confirmation"`
}

func (c *dbaasDatabaseDeleteCmd) CmdAliases() []string { return nil }

func (c *dbaasDatabaseDeleteCmd) CmdShort() string { return "Delete DBAAS database" }

func (c *dbaasDatabaseDeleteCmd) CmdLong() string {
	return `This command deletes a DBAAS database for the specified service.`
}

func (c *dbaasDatabaseDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {

	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasDatabaseDeleteCmd) CmdRun(cmd *cobra.Command, args []string) error {

	ctx := GContext
	db, err := dbaasGetV3(ctx, c.Name, c.Zone)
	if err != nil {
		return err
	}

	switch db.Type {
	case "mysql":
		return c.deleteMysql(cmd, args)
	case "pg":
		return c.deletePg(cmd, args)
	default:
		return fmt.Errorf("creating database unsupported for service of type %q", db.Type)
	}

}

func init() {
	cobra.CheckErr(RegisterCLICommand(dbaasDatabaseCmd, &dbaasDatabaseDeleteCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
