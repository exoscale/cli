package cmd

import (
	"fmt"

	"github.com/exoscale/cli/pkg/account"
	exoapi "github.com/exoscale/egoscale/v2/api"
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
	return fmt.Sprintf(`This command deletes a DBAAS user for the specified service.`)
}

func (c *dbaasUserDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasUserDeleteCmd) cmdRun(cmd *cobra.Command, args []string) error {

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))
	dbType, err := dbaasGetType(ctx, c.Name, c.Zone)
	if err != nil {
		return err
	}

	switch dbType {
	case "mysql":
		return c.deleteMysql(cmd, args)
	case "kafka":
		return c.deleteKafka(cmd, args)
	case "pg":
		return c.deletePg(cmd, args)
	case "opensearch":
		return c.deleteOpensearch(cmd, args)
	case "redis":
		return c.deleteRedis(cmd, args)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasUserCmd, &dbaasUserDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
