package instance

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/cmd/compute/vpc"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceVPCDetachCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"detach"`

	Instance string `cli-arg:"#" cli-usage:"INSTANCE-NAME|ID"`
	VPC      string `cli-arg:"#" cli-usage:"VPC-NAME|ID"`
	Subnet   string `cli-arg:"#" cli-usage:"SUBNET-NAME|ID"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceVPCDetachCmd) CmdAliases() []string { return nil }

func (c *instanceVPCDetachCmd) CmdShort() string {
	return "Detach a Compute instance from a VPC Subnet"
}

func (c *instanceVPCDetachCmd) CmdLong() string {
	return "This command detaches a Compute instance from a VPC Subnet."
}

func (c *instanceVPCDetachCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceVPCDetachCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	instances, err := client.ListInstances(ctx)
	if err != nil {
		return err
	}

	instance, err := findInstance(instances, c.Instance, string(c.Zone))
	if err != nil {
		return err
	}

	vpcEntry, err := vpc.FindVPC(ctx, client, c.VPC)
	if err != nil {
		return err
	}

	subnetEntry, err := vpc.FindSubnet(ctx, client, vpcEntry.ID, c.Subnet)
	if err != nil {
		return err
	}

	req := v3.DetachInstanceFromSubnetRequest{
		Instance: &v3.InstanceRef{ID: instance.ID},
	}

	if err := utils.RunAsync(
		ctx,
		client,
		fmt.Sprintf("Detaching instance %q from Subnet %q...", c.Instance, c.Subnet),
		func(ctx context.Context, client *v3.Client) (*v3.Operation, error) {
			return client.DetachInstanceFromSubnet(ctx, vpcEntry.ID, subnetEntry.ID, req)
		},
	); err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&instanceShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Instance:           instance.ID.String(),
			Zone:               c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceVPCCmd, &instanceVPCDetachCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
