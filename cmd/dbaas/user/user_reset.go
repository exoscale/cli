package user

import (
	"fmt"
	"os"

	exocmd "github.com/exoscale/cli/cmd"
	dbaasutils "github.com/exoscale/cli/cmd/dbaas/internal"
	"github.com/exoscale/cli/pkg/clicmd"

	"github.com/spf13/cobra"
)

type dbaasUserResetCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"reset-credentials"`

	Name     string `cli-arg:"#"`
	Username string `cli-arg:"#"`
	Zone     string `cli-short:"z" cli-usage:"Database Service zone"`

	Password string `cli-flag:"password" cli-usage:"Use a specific password instead of an automatically generated one"`

	HelpMysql bool `cli-usage:"show usage for flags specific to the mysql type"`

	// "mysql" type specific flags
	MysqlAuthenticationMethod string `cli-flag:"mysql-auhentication-method" cli-usage:"authentication method to be used (\"caching_sha2_password\" or \"mysql_native_password\")." cli-hidden:""`
}

func (c *dbaasUserResetCmd) CmdAliases() []string { return nil }

func (c *dbaasUserResetCmd) CmdShort() string { return "Reset the credentials of a DBAAS user" }

func (c *dbaasUserResetCmd) CmdLong() string {
	return `This command resets the credentials of a DBAAS user for the specified service.`
}

func (c *dbaasUserResetCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	clicmd.CmdSetZoneFlagFromDefault(cmd)

	switch {

	case cmd.Flags().Changed("help-mysql"):
		clicmd.CmdShowHelpFlags(cmd.Flags(), "mysql-")
		os.Exit(0)
	}

	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasUserResetCmd) CmdRun(cmd *cobra.Command, args []string) error {

	ctx := exocmd.GContext
	db, err := dbaasutils.GetV3(ctx, c.Name, c.Zone)
	if err != nil {
		return err
	}

	switch db.Type {
	case "mysql":
		return c.resetMysql()
	case "kafka":
		return c.resetKafka()
	case "pg":
		return c.resetPG()
	case "opensearch":
		return c.resetOpensearch()
	case "grafana":
		return c.resetGrafana()
	case "valkey":
		return c.resetValkey()
	case "clickhouse":
		return c.resetClickhouse()
	default:
		return fmt.Errorf("reseting user credentials unsupported for service of type %q", db.Type)
	}

}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &dbaasUserResetCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
