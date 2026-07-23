package vpc

import (
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

type vpcRouteCreateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	VPC string `cli-arg:"#" cli-usage:"VPC-NAME|ID"`

	Subnet      string      `cli-usage:"Subnet to create the route in (NAME|ID)"`
	Destination string      `cli-usage:"route destination CIDR (e.g. 10.9.0.0/24)"`
	Target      string      `cli-usage:"route target, as ip=<IP address> (e.g. ip=10.0.0.5)"`
	Description string      `cli-usage:"route description"`
	Zone        v3.ZoneName `cli-short:"z" cli-usage:"VPC zone"`
}

func (c *vpcRouteCreateCmd) CmdAliases() []string { return exocmd.GCreateAlias }

func (c *vpcRouteCreateCmd) CmdShort() string { return "Create a VPC route" }

func (c *vpcRouteCreateCmd) CmdLong() string {
	return `This command creates a route on a VPC Subnet.

Routes are scoped to a Subnet, so --subnet is required.`
}

func (c *vpcRouteCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	if err := exocmd.CliCommandDefaultPreRun(c, cmd, args); err != nil {
		return err
	}

	return exocmd.CmdCheckRequiredFlags(cmd, []string{"subnet", "destination", "target"})
}

func (c *vpcRouteCreateCmd) CmdRun(_ *cobra.Command, _ []string) error {
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

	if _, err := client.CreateRoute(ctx, vpcEntry.ID, subnetEntry.ID, v3.CreateRouteRequest{
		Destination: c.Destination,
		Target:      c.Target,
		Description: c.Description,
	}); err != nil {
		return err
	}

	// Routes have no show command of their own (they are unnamed and the API
	// exposes no per-route GET), so list the Subnet's routes instead.
	if !globalstate.Quiet {
		return (&vpcRouteListCmd{
			CliCommandSettings: c.CliCommandSettings,
			VPC:                vpcEntry.ID.String(),
			Subnet:             subnetEntry.ID.String(),
			Zone:               c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(vpcRouteCmd, &vpcRouteCreateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
