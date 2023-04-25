package cmd

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/exoscale/cli/pkg/output"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instancePrivnetAttachCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"attach"`

	Instance       string `cli-arg:"#" cli-usage:"INSTANCE-NAME|ID"`
	PrivateNetwork string `cli-arg:"#" cli-usage:"PRIVATE-NETWORK-NAME|ID"`

	IPAddress string `cli-flag:"ip" cli-usage:"network IP address to assign to the Compute instance (managed Private Networks only)"`
	Zone      string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instancePrivnetAttachCmd) cmdAliases() []string { return nil }

func (c *instancePrivnetAttachCmd) cmdShort() string {
	return "Attach a Compute instance to a Private Network"
}

func (c *instancePrivnetAttachCmd) cmdLong() string {
	return fmt.Sprintf(`This command attaches a Compute instance to a Private Network.

Supported output template annotations: %s`,
		strings.Join(output.OutputterTemplateAnnotations(&instanceShowOutput{}), ", "),
	)
}

func (c *instancePrivnetAttachCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instancePrivnetAttachCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	instance, err := cs.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	privateNetwork, err := cs.FindPrivateNetwork(ctx, c.Zone, c.PrivateNetwork)
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
		if err = cs.AttachInstanceToPrivateNetwork(ctx, c.Zone, instance, privateNetwork, opts...); err != nil {
			return
		}
	})
	if err != nil {
		return err
	}

	if !gQuiet {
		return (&instanceShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			Instance:           *instance.ID,
			Zone:               c.Zone,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(instancePrivnetCmd, &instancePrivnetAttachCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
