package instance

import (
	"context"
	"fmt"
	"net"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/cmd/compute/vpc"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceVPCAttachCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"attach"`

	Instance string `cli-arg:"#" cli-usage:"INSTANCE-NAME|ID"`
	VPC      string `cli-arg:"#" cli-usage:"VPC-NAME|ID"`
	Subnet   string `cli-arg:"#" cli-usage:"SUBNET-NAME|ID"`

	IPv4 string      `cli-flag:"ipv4" cli-usage:"IPv4 address to assign to the Compute instance in the Subnet"`
	Zone v3.ZoneName `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceVPCAttachCmd) CmdAliases() []string { return nil }

func (c *instanceVPCAttachCmd) CmdShort() string {
	return "Attach a Compute instance to a VPC Subnet"
}

func (c *instanceVPCAttachCmd) CmdLong() string {
	return "This command attaches a Compute instance to a VPC Subnet."
}

func (c *instanceVPCAttachCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceVPCAttachCmd) CmdRun(_ *cobra.Command, _ []string) error {
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

	req := v3.AttachInstanceToSubnetRequest{
		Instance: &v3.InstanceRef{ID: instance.ID},
	}

	if c.IPv4 != "" {
		ip := net.ParseIP(c.IPv4)
		if ip == nil || ip.To4() == nil {
			return fmt.Errorf("invalid IPv4 address: %q", c.IPv4)
		}
		req.Ipv4 = ip
	}

	if err := utils.RunAsync(
		ctx,
		client,
		fmt.Sprintf("Attaching instance %q to Subnet %q...", c.Instance, c.Subnet),
		func(ctx context.Context, client *v3.Client) (*v3.Operation, error) {
			return client.AttachInstanceToSubnet(ctx, vpcEntry.ID, subnetEntry.ID, req)
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
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceVPCCmd, &instanceVPCAttachCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
