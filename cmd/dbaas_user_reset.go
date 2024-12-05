package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type dbaasUserResetCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"reset-credentials"`

	Name     string `cli-arg:"#"`
	Username string `cli-arg:"#"`
	Zone     string `cli-short:"z" cli-usage:"Database Service zone"`

	Password string `cli-flag:"password" cli-usage:"Use a specific password instead of an automatically generated one"`

	HelpMysql bool `cli-usage:"show usage for flags specific to the mysql type"`

	// "mysql" type specific flags
	MysqlAuthenticationMethod string `cli-flag:"mysql-auhentication-method" cli-usage:"authentication method to be used (\"caching_sha2_password\" or \"mysql_native_password\")." cli-hidden:""`
}

func (c *dbaasUserResetCmd) cmdAliases() []string { return nil }

func (c *dbaasUserResetCmd) cmdShort() string { return "Reset the credentials of a DBAAS user" }

func (c *dbaasUserResetCmd) cmdLong() string {
	return `This command resets the credentials of a DBAAS user for the specified service.`
}

func (c *dbaasUserResetCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)

	switch {

	case cmd.Flags().Changed("help-mysql"):
		cmdShowHelpFlags(cmd.Flags(), "mysql-")
		os.Exit(0)
	}

	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasUserResetCmd) cmdRun(cmd *cobra.Command, args []string) error {

	ctx := gContext
	db, err := dbaasGetV3(ctx, c.Name, c.Zone)
	if err != nil {
		return err
	}

	switch db.Type {
	case "mysql":
		return c.resetMysql(cmd, args)
	case "kafka":
		return c.resetKafka(cmd, args)
	case "pg":
		return c.resetPG(cmd, args)
	case "opensearch":
		return c.resetOpensearch(cmd, args)
	case "grafana":
		return c.resetGrafana(cmd, args)
	default:
		return fmt.Errorf("reseting user credentials unsupported for service of type %q", db.Type)
	}

}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasUserCmd, &dbaasUserResetCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}