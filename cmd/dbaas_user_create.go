package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type dbaasUserCreateCmd struct {
	CliCommandSettings `cli-cmd:"-"`

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

func (c *dbaasUserCreateCmd) CmdAliases() []string { return nil }

func (c *dbaasUserCreateCmd) CmdShort() string { return "Create DBAAS user" }

func (c *dbaasUserCreateCmd) CmdLong() string {
	return `This command creates a DBAAS user for the specified service.`
}

func (c *dbaasUserCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	switch {

	case cmd.Flags().Changed("help-mysql"):
		cmdShowHelpFlags(cmd.Flags(), "mysql-")
		os.Exit(0)
	case cmd.Flags().Changed("help-pg"):
		cmdShowHelpFlags(cmd.Flags(), "pg-")
		os.Exit(0)
	}

	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasUserCreateCmd) CmdRun(cmd *cobra.Command, args []string) error {

	ctx := GContext
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
	case "valkey":
		return c.createValkey(cmd, args)
	default:
		return fmt.Errorf("creating user unsupported for service of type %q", db.Type)
	}

}

func init() {
	cobra.CheckErr(RegisterCLICommand(dbaasUserCmd, &dbaasUserCreateCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
