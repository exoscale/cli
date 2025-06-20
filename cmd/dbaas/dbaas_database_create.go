package dbaas

import (
	"fmt"
	"os"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/spf13/cobra"
)

type dbaasDatabaseCreateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name     string `cli-arg:"#"`
	Database string `cli-arg:"#"`

	HelpPg bool   `cli-usage:"show usage for flags specific to the pg type"`
	Zone   string `cli-short:"z" cli-usage:"Database Service zone"`

	// "pg" type specific flags
	PgLcCollate string `cli-usage:"Default string sort order (LC_COLLATE) for PostgreSQL database" cli-hidden:""`
	PgLcCtype   string `cli-usage:"Default character classification (LC_CTYPE) for PostgreSQL database" cli-hidden:""`
}

func (c *dbaasDatabaseCreateCmd) CmdAliases() []string { return nil }

func (c *dbaasDatabaseCreateCmd) CmdShort() string { return "Create DBAAS database" }

func (c *dbaasDatabaseCreateCmd) CmdLong() string {
	return `This command creates a DBAAS database for the specified service.`
}

func (c *dbaasDatabaseCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	switch {

	case cmd.Flags().Changed("help-mysql"):
		exocmd.CmdShowHelpFlags(cmd.Flags(), "mysql-")
		os.Exit(0)
	case cmd.Flags().Changed("help-pg"):
		exocmd.CmdShowHelpFlags(cmd.Flags(), "pg-")
		os.Exit(0)
	}

	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasDatabaseCreateCmd) CmdRun(cmd *cobra.Command, args []string) error {

	ctx := exocmd.GContext
	db, err := dbaasGetV3(ctx, c.Name, c.Zone)
	if err != nil {
		return err
	}

	switch db.Type {
	case "mysql":
		return c.createMysql(cmd, args)
	case "pg":
		return c.createPg(cmd, args)
	default:
		return fmt.Errorf("creating database unsupported for service of type %q", db.Type)
	}

}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasDatabaseCmd, &dbaasDatabaseCreateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
