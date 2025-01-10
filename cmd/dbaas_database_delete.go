package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type dbaasDatabaseDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Name     string `cli-arg:"#"`
	Database string `cli-arg:"#"`

	Zone string `cli-short:"z" cli-usage:"Database Service zone"`

	Force bool `cli-short:"f" cli-usage:"don't prompt for confirmation"`
}

func (c *dbaasDatabaseDeleteCmd) cmdAliases() []string { return nil }

func (c *dbaasDatabaseDeleteCmd) cmdShort() string { return "Delete DBAAS database" }

func (c *dbaasDatabaseDeleteCmd) cmdLong() string {
	return `This command deletes a DBAAS database for the specified service.`
}

func (c *dbaasDatabaseDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {

	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasDatabaseDeleteCmd) cmdRun(cmd *cobra.Command, args []string) error {

	ctx := gContext
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
	cobra.CheckErr(registerCLICommand(dbaasDatabaseCmd, &dbaasDatabaseDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
