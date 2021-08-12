package cmd

import (
	"fmt"
	"strings"

	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instancePrivnetRemoveCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"remove"`

	Instance        string   `cli-arg:"#" cli-usage:"INSTANCE-NAME|ID"`
	PrivateNetworks []string `cli-arg:"*" cli-usage:"SECURITY-GROUP-NAME|ID"`

	Zone string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instancePrivnetRemoveCmd) cmdAliases() []string { return gRemoveAlias }

func (c *instancePrivnetRemoveCmd) cmdShort() string {
	return "Remove a Compute instance from Private Networks"
}

func (c *instancePrivnetRemoveCmd) cmdLong() string {
	return fmt.Sprintf(`This command removes a Compute instance from Private Networks.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&instanceShowOutput{}), ", "),
	)
}

func (c *instancePrivnetRemoveCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instancePrivnetRemoveCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	if len(c.PrivateNetworks) == 0 {
		cmdExitOnUsageError(cmd, "no Private Networks specified")
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	instance, err := cs.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		return err
	}

	privateNetworks := make([]*egoscale.PrivateNetwork, len(c.PrivateNetworks))
	for i := range c.PrivateNetworks {
		privateNetwork, err := cs.FindPrivateNetwork(ctx, c.Zone, c.PrivateNetworks[i])
		if err != nil {
			return fmt.Errorf("error retrieving Private Network: %s", err)
		}
		privateNetworks[i] = privateNetwork
	}

	decorateAsyncOperation(fmt.Sprintf("Updating instance %q Private Networks...", c.Instance), func() {
		for _, privateNetwork := range privateNetworks {
			if err = cs.DetachInstanceFromPrivateNetwork(ctx, c.Zone, instance, privateNetwork); err != nil {
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
	cobra.CheckErr(registerCLICommand(computeInstancePrivnetCmd, &instancePrivnetRemoveCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
