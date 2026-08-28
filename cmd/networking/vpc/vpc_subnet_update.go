package vpc

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type vpcSubnetUpdateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	VPC    string `cli-arg:"#" cli-usage:"VPC-NAME|ID"`
	Subnet string `cli-arg:"#" cli-usage:"SUBNET-NAME|ID"`

	Name        string            `cli-usage:"Subnet name"`
	Description string            `cli-usage:"Subnet description"`
	IPv4Block   string            `cli-flag:"ipv4-block" cli-usage:"Subnet IPv4 CIDR (e.g. 10.0.0.0/24)"`
	Labels      map[string]string `cli-flag:"label" cli-usage:"Subnet label (format: key=value), clearing the labels is possible by passing [=]"`
	Zone        v3.ZoneName       `cli-short:"z" cli-usage:"VPC zone"`
}

func (c *vpcSubnetUpdateCmd) CmdAliases() []string { return nil }

func (c *vpcSubnetUpdateCmd) CmdShort() string { return "Update a VPC Subnet" }

func (c *vpcSubnetUpdateCmd) CmdLong() string {
	return fmt.Sprintf(`This command updates a VPC Subnet.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&vpcSubnetShowOutput{}), ", "))
}

func (c *vpcSubnetUpdateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *vpcSubnetUpdateCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	var updated bool

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

	req := v3.UpdateSubnetRequest{}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Name)) {
		req.Name = &c.Name
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Description)) {
		req.Description = &c.Description
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.IPv4Block)) {
		req.Ipv4Block = &c.IPv4Block
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Labels)) {
		req.Labels = exocmd.ConvertIfSpecialEmptyMap(c.Labels)
		updated = true
	}

	if updated {
		if _, err := client.UpdateSubnet(ctx, vpcEntry.ID, subnetEntry.ID, req); err != nil {
			return err
		}
	}

	if !globalstate.Quiet {
		return (&vpcSubnetShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			VPC:                vpcEntry.ID.String(),
			Subnet:             subnetEntry.ID.String(),
			Zone:               c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(vpcSubnetCmd, &vpcSubnetUpdateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
