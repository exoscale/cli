package vpc

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type vpcSubnetCreateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	VPC  string `cli-arg:"#" cli-usage:"VPC-NAME|ID"`
	Name string `cli-arg:"#" cli-usage:"NAME"`

	// TODO: I'm not a fan of the usability of this command
	IPv4Block     string            `cli-flag:"ipv4-block" cli-usage:"Subnet IPv4 CIDR (e.g. 10.0.0.0/24)"`
	AddressFamily string            `cli-flag:"address-family" cli-usage:"Subnet address family (currently only \"inet4\" is supported)"`
	AddressSpace  string            `cli-flag:"address-space" cli-usage:"Subnet address space (currently only \"private\" is supported)"`
	Description   string            `cli-usage:"Subnet description"`
	Labels        map[string]string `cli-flag:"label" cli-usage:"Subnet label (format: key=value)"`
	Zone          v3.ZoneName       `cli-short:"z" cli-usage:"VPC zone"`
}

func (c *vpcSubnetCreateCmd) CmdAliases() []string { return exocmd.GCreateAlias }

func (c *vpcSubnetCreateCmd) CmdShort() string { return "Create a VPC Subnet" }

func (c *vpcSubnetCreateCmd) CmdLong() string {
	return fmt.Sprintf(`This command creates a Subnet in a Virtual Private Cloud.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&vpcSubnetShowOutput{}), ", "))
}

func (c *vpcSubnetCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *vpcSubnetCreateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	vpcEntry, err := FindVPC(ctx, client, c.VPC)
	if err != nil {
		return err
	}

	req := v3.CreateSubnetRequest{
		Name:          c.Name,
		Description:   c.Description,
		Ipv4Block:     c.IPv4Block,
		AddressSpace:  v3.CreateSubnetRequestAddressSpace(c.AddressSpace),
		Addressfamily: v3.CreateSubnetRequestAddressfamily(c.AddressFamily),
	}

	if len(c.Labels) > 0 {
		req.Labels = c.Labels
	}

	op, err := client.CreateSubnet(ctx, vpcEntry.ID, req)
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Creating Subnet %q...", c.Name), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&vpcSubnetShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			VPC:                vpcEntry.ID.String(),
			Subnet:             op.Reference.ID.String(),
			Zone:               c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(vpcSubnetCmd, &vpcSubnetCreateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
