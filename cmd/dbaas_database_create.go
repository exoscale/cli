package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type dbaasDatabaseCreateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name     string `cli-arg:"#"`
	Database string `cli-arg:"#"`

	HelpPg bool   `cli-usage:"show usage for flags specific to the pg type"`
	Zone   string `cli-short:"z" cli-usage:"Database Service zone"`

	// "pg" type specific flags
	PgLcCollate string `cli-usage:"Default string sort order (LC_COLLATE) for PostgreSQL database" cli-hidden:""`
	PgLcCtype   string `cli-usage:"Default character classification (LC_CTYPE) for PostgreSQL database" cli-hidden:""`
}

func (c *dbaasDatabaseCreateCmd) cmdAliases() []string { return nil }

func (c *dbaasDatabaseCreateCmd) cmdShort() string { return "Create DBAAS database" }

func (c *dbaasDatabaseCreateCmd) cmdLong() string {
	return `This command creates a DBAAS database for the specified service.`
}

func (c *dbaasDatabaseCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	switch {

	case cmd.Flags().Changed("help-mysql"):
		cmdShowHelpFlags(cmd.Flags(), "mysql-")
		os.Exit(0)
	case cmd.Flags().Changed("help-pg"):
		cmdShowHelpFlags(cmd.Flags(), "pg-")
		os.Exit(0)
	}

	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasDatabaseCreateCmd) cmdRun(cmd *cobra.Command, args []string) error {

	ctx := gContext
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
	cobra.CheckErr(registerCLICommand(dbaasDatabaseCmd, &dbaasDatabaseCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
