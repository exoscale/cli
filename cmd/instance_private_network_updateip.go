package cmd

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	exoapi "github.com/exoscale/egoscale/v2/api"
	v3 "github.com/exoscale/egoscale/v3"
)

type instancePrivnetUpdateIPCmd struct {
	CliCommandSettings `cli-cmd:"-"`

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
		strings.Join(output.TemplateAnnotations(&instanceShowOutput{}), ", "),
	)
}

func (c *instancePrivnetUpdateIPCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instancePrivnetUpdateIPCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	instance, err := globalstate.EgoscaleClient.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	privateNetwork, err := globalstate.EgoscaleClient.FindPrivateNetwork(ctx, c.Zone, c.PrivateNetwork)
	if err != nil {
		return fmt.Errorf("error retrieving Private Network: %w", err)
	}

	decorateAsyncOperation(fmt.Sprintf("Updating instance %q Private Network IP address...", c.Instance), func() {
		if err = globalstate.EgoscaleClient.UpdatePrivateNetworkInstanceIPAddress(
			ctx,
			c.Zone,
			instance,
			privateNetwork,
			net.ParseIP(c.IPAddress),
		); err != nil {
			return
		}
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&instanceShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Instance:           *instance.ID,
			Zone:               v3.ZoneName(c.Zone),
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(RegisterCLICommand(instancePrivnetCmd, &instancePrivnetUpdateIPCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
