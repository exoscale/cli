package instance

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

type instancePrivnetDetachCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"detach"`

	Instance       string `cli-arg:"#" cli-usage:"INSTANCE-NAME|ID"`
	PrivateNetwork string `cli-arg:"#" cli-usage:"PRIVATE-NETWORK-NAME|ID"`

	Zone string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instancePrivnetDetachCmd) CmdAliases() []string { return nil }

func (c *instancePrivnetDetachCmd) CmdShort() string {
	return "Detach a Compute instance from a Private Network"
}

func (c *instancePrivnetDetachCmd) CmdLong() string {
	return fmt.Sprintf(`This command detaches a Compute instance from a Private Network.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&InstanceShowOutput{}), ", "),
	)
}

func (c *instancePrivnetDetachCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instancePrivnetDetachCmd) CmdRun(_ *cobra.Command, _ []string) error {
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

	op, err := client.DetachInstanceFromPrivateNetwork(ctx, privateNetwork.ID, v3.DetachInstanceFromPrivateNetworkRequest{
		Instance: &v3.Instance{
			ID: instance.ID,
		},
	})
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf(
		"Detaching instance %q from Private Network %q...",
		c.Instance,
		c.PrivateNetwork,
	), func() {
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
	cobra.CheckErr(exocmd.RegisterCLICommand(instancePrivnetCmd, &instancePrivnetDetachCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
