package cmd

import (
	"fmt"
	"net"
	"strings"

	exov2 "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instancePrivnetAddCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"add"`

	Instance        string   `cli-arg:"#" cli-usage:"INSTANCE-NAME|ID"`
	PrivateNetworks []string `cli-arg:"*" cli-usage:"PRIVATE-NETWORK-NAME|ID"`

	IPAddress string `cli-flag:"ip" cli-usage:"network IP address to assign to the Compute instance (managed Private Networks only)"`
	Zone      string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instancePrivnetAddCmd) cmdAliases() []string { return nil }

func (c *instancePrivnetAddCmd) cmdShort() string {
	return "Add a Compute instance to Private Networks"
}

func (c *instancePrivnetAddCmd) cmdLong() string {
	return fmt.Sprintf(`This command adds a Compute instance to Private Networks.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&instanceShowOutput{}), ", "),
	)
}

func (c *instancePrivnetAddCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instancePrivnetAddCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	if len(c.PrivateNetworks) == 0 {
		cmdExitOnUsageError(cmd, "no Private Networks specified")
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	instance, err := cs.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		return err
	}

	privateNetworks := make([]*exov2.PrivateNetwork, len(c.PrivateNetworks))
	for i := range c.PrivateNetworks {
		privateNetwork, err := cs.FindPrivateNetwork(ctx, c.Zone, c.PrivateNetworks[i])
		if err != nil {
			return fmt.Errorf("error retrieving Private Network: %s", err)
		}
		privateNetworks[i] = privateNetwork
	}

	decorateAsyncOperation(fmt.Sprintf("Updating instance %q Private Networks...", c.Instance), func() {
		for _, privateNetwork := range privateNetworks {
			if err = instance.AttachPrivateNetwork(ctx, privateNetwork, net.ParseIP(c.IPAddress)); err != nil {
				return
			}
		}
	})
	if err != nil {
		return err
	}

	if !gQuiet {
		return output(showInstance(c.Zone, *instance.ID))
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(computeInstancePrivnetCmd, &instancePrivnetAddCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
