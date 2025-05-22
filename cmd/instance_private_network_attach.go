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
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	v3 "github.com/exoscale/egoscale/v3"
)

type instancePrivnetAttachCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"attach"`

	Instance       string `cli-arg:"#" cli-usage:"INSTANCE-NAME|ID"`
	PrivateNetwork string `cli-arg:"#" cli-usage:"PRIVATE-NETWORK-NAME|ID"`

	IPAddress string `cli-flag:"ip" cli-usage:"network IP address to assign to the Compute instance (managed Private Networks only)"`
	Zone      string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instancePrivnetAttachCmd) CmdAliases() []string { return nil }

func (c *instancePrivnetAttachCmd) CmdShort() string {
	return "Attach a Compute instance to a Private Network"
}

func (c *instancePrivnetAttachCmd) CmdLong() string {
	return fmt.Sprintf(`This command attaches a Compute instance to a Private Network.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&InstanceShowOutput{}), ", "),
	)
}

func (c *instancePrivnetAttachCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instancePrivnetAttachCmd) CmdRun(_ *cobra.Command, _ []string) error {
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

	opts := make([]egoscale.AttachInstanceToPrivateNetworkOpt, 0)
	if c.IPAddress != "" {
		opts = append(opts, egoscale.AttachInstanceToPrivateNetworkWithIPAddress(net.ParseIP(c.IPAddress)))
	}

	decorateAsyncOperation(fmt.Sprintf(
		"Attaching instance %q to Private Network %q...",
		c.Instance,
		c.PrivateNetwork,
	), func() {
		if err = globalstate.EgoscaleClient.AttachInstanceToPrivateNetwork(ctx, c.Zone, instance, privateNetwork, opts...); err != nil {
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
	cobra.CheckErr(RegisterCLICommand(instancePrivnetCmd, &instancePrivnetAttachCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
