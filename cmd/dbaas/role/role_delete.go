package role

import (
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbaasRoleDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_        bool   `cli-cmd:"delete"`
	Name     string `cli-arg:"#" cli-usage:"NAME"`
	RoleUUID string `cli-arg:"#" cli-usage:"ROLE-UUID"`
	Zone     string `cli-short:"z" cli-usage:"Database Service zone"`

	Force bool `cli-short:"f" cli-usage:"don't prompt for confirmation"`
}

func (c *dbaasRoleDeleteCmd) CmdAliases() []string { return nil }
func (c *dbaasRoleDeleteCmd) CmdShort() string     { return "Delete a ClickHouse role" }
func (c *dbaasRoleDeleteCmd) CmdLong() string {
	return "Delete a role from a ClickHouse DBaaS service by UUID."
}

func (c *dbaasRoleDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasRoleDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	if !c.Force {
		if !utils.AskQuestion(
			ctx,
			fmt.Sprintf(
				"Are you sure you want to delete role %q from service %q?",
				c.RoleUUID,
				c.Name,
			),
		) {
			return nil
		}
	}

	op, err := client.DeleteDBAASClickhouseRole(ctx, c.Name, v3.UUID(c.RoleUUID))
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Deleting role %q from service %q", c.RoleUUID, c.Name), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	return err
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &dbaasRoleDeleteCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
