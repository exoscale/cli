package vpc

import (
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type vpcRouteDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	VPC   string `cli-arg:"#" cli-usage:"VPC-NAME|ID"`
	Route string `cli-arg:"#" cli-usage:"ROUTE-ID"`

	Subnet string      `cli-usage:"Subnet the route belongs to (NAME|ID)"`
	Force  bool        `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone   v3.ZoneName `cli-short:"z" cli-usage:"VPC zone"`
}

func (c *vpcRouteDeleteCmd) CmdAliases() []string { return exocmd.GRemoveAlias }

func (c *vpcRouteDeleteCmd) CmdShort() string { return "Delete a VPC route" }

func (c *vpcRouteDeleteCmd) CmdLong() string {
	return `This command deletes a route from a VPC Subnet.

Routes associated to a VPC directly can't be deleted, so --subnet is required.`
}

func (c *vpcRouteDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	if err := exocmd.CliCommandDefaultPreRun(c, cmd, args); err != nil {
		return err
	}

	return exocmd.CmdCheckRequiredFlags(cmd, []string{"subnet"})
}

func (c *vpcRouteDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	vpcEntry, err := FindVPC(ctx, client, c.VPC)
	if err != nil {
		return err
	}

	subnetEntry, err := FindSubnet(ctx, client, vpcEntry.ID, c.Subnet)
	if err != nil {
		return err
	}

	if !c.Force {
		if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to delete route %s?", c.Route)) {
			return nil
		}
	}

	if _, err := client.DeleteRoute(ctx, vpcEntry.ID, subnetEntry.ID, v3.UUID(c.Route)); err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(vpcRouteCmd, &vpcRouteDeleteCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
