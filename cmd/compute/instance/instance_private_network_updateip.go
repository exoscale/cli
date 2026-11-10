package instance

import (
	"fmt"
	"net"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type instancePrivnetUpdateIPCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update-ip"`

	Instance       string `cli-arg:"#" cli-usage:"INSTANCE-NAME|ID"`
	PrivateNetwork string `cli-arg:"#" cli-usage:"PRIVATE-NETWORK-NAME|ID"`
	IPAddress      string `cli-flag:"ip" cli-usage:"network IP address to assign to the Compute instance"`

	Zone string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instancePrivnetUpdateIPCmd) CmdAliases() []string { return nil }

func (c *instancePrivnetUpdateIPCmd) CmdShort() string {
	return "Update a Compute instance Private Network IP address"
}

func (c *instancePrivnetUpdateIPCmd) CmdLong() string {
	return fmt.Sprintf(`This command updates the IP address assigned to a Compute instance in a
managed Private Network.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&InstanceShowOutput{}), ", "),
	)
}

func (c *instancePrivnetUpdateIPCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instancePrivnetUpdateIPCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	instances, err := client.ListInstances(ctx)
	if err != nil {
		return err
	}
	instance, err := findInstance(instances, c.Instance, c.Zone)
	if err != nil {
		return err
	}

	privateNetworks, err := client.ListPrivateNetworks(ctx)
	if err != nil {
		return err
	}
	privateNetwork, err := privateNetworks.FindPrivateNetwork(c.PrivateNetwork)
	if err != nil {
		return err
	}

	op, err := client.UpdatePrivateNetworkInstanceIP(ctx, privateNetwork.ID, v3.UpdatePrivateNetworkInstanceIPRequest{
		Instance: &v3.UpdatePrivateNetworkInstanceIPRequestInstance{
			ID: instance.ID,
		},
		IP: net.ParseIP(c.IPAddress),
	})
	utils.DecorateAsyncOperation(fmt.Sprintf("Updating instance %q Private Network IP address...", c.Instance), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&instanceShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Instance:           instance.ID.String(),
			Zone:               v3.ZoneName(c.Zone),
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instancePrivnetCmd, &instancePrivnetUpdateIPCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
