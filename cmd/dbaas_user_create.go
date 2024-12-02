package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type dbaasUserCreateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name     string `cli-arg:"#"`
	Username string `cli-arg:"#"`

	HelpMysql bool   `cli-usage:"show usage for flags specific to the mysql type"`
	HelpPg    bool   `cli-usage:"show usage for flags specific to the pg type"`
	Zone      string `cli-short:"z" cli-usage:"Database Service zone"`

	// "mysql" type specific flags
	MysqlAuthenticationMethod string `cli-flag:"mysql-authentication-method" cli-usage:"authentication method to be used (\"caching_sha2_password\" or \"mysql_native_password\")." cli-hidden:""`

	// "kafka" type specific flags
	PostgresAllowReplication bool `cli-flag:"pg-allow-replication" cli-usage:"" cli-hidden:""`
}

func (c *dbaasUserCreateCmd) cmdAliases() []string { return nil }

func (c *dbaasUserCreateCmd) cmdShort() string { return "Create DBAAS user" }

func (c *dbaasUserCreateCmd) cmdLong() string {
	return `This command creates a DBAAS user for the specified service.`
}

func (c *dbaasUserCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
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

func (c *dbaasUserCreateCmd) cmdRun(cmd *cobra.Command, args []string) error {

	ctx := gContext
	db, err := dbaasGetV3(ctx, c.Name, c.Zone)
	if err != nil {
		return err
	}

	switch db.Type {
	case "mysql":
		return c.createMysql(cmd, args)
	case "kafka":
		return c.createKafka(cmd, args)
	case "pg":
		return c.createPg(cmd, args)
	case "opensearch":
		return c.createOpensearch(cmd, args)
	case "redis":
		return c.createRedis(cmd, args)
	default:
		return fmt.Errorf("creating user unsupported for service of type %q", db.Type)
	}

}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasUserCmd, &dbaasUserCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
