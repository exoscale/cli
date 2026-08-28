package vpc

import (
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type vpcSubnetDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	VPC    string `cli-arg:"#" cli-usage:"VPC-NAME|ID"`
	Subnet string `cli-arg:"#" cli-usage:"SUBNET-NAME|ID"`

	Force bool        `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  v3.ZoneName `cli-short:"z" cli-usage:"VPC zone"`
}

func (c *vpcSubnetDeleteCmd) CmdAliases() []string { return exocmd.GRemoveAlias }

func (c *vpcSubnetDeleteCmd) CmdShort() string { return "Delete a VPC Subnet" }

func (c *vpcSubnetDeleteCmd) CmdLong() string { return "" }

func (c *vpcSubnetDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *vpcSubnetDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
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
		if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to delete Subnet %s?", c.Subnet)) {
			return nil
		}
	}

	if _, err := client.DeleteSubnet(ctx, vpcEntry.ID, subnetEntry.ID); err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(vpcSubnetCmd, &vpcSubnetDeleteCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
