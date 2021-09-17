package cmd

import (
	"fmt"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instancePrivnetDetachCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"detach"`

	Instance       string `cli-arg:"#" cli-usage:"INSTANCE-NAME|ID"`
	PrivateNetwork string `cli-arg:"#" cli-usage:"PRIVATE-NETWORK-NAME|ID"`

	Zone string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instancePrivnetDetachCmd) cmdAliases() []string { return nil }

func (c *instancePrivnetDetachCmd) cmdShort() string {
	return "Detach a Compute instance from a Private Network"
}

func (c *instancePrivnetDetachCmd) cmdLong() string {
	return fmt.Sprintf(`This command detaches a Compute instance from a Private Network.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&instanceShowOutput{}), ", "),
	)
}

func (c *instancePrivnetDetachCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instancePrivnetDetachCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	instance, err := cs.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		return err
	}

	privateNetwork, err := cs.FindPrivateNetwork(ctx, c.Zone, c.PrivateNetwork)
	if err != nil {
		return fmt.Errorf("error retrieving Private Network: %s", err)
	}

	decorateAsyncOperation(fmt.Sprintf(
		"Detaching instance %q from Private Network %q...",
		c.Instance,
		c.PrivateNetwork,
	), func() {
		if err = cs.DetachInstanceFromPrivateNetwork(ctx, c.Zone, instance, privateNetwork); err != nil {
			return
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
	cobra.CheckErr(registerCLICommand(instancePrivnetCmd, &instancePrivnetDetachCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
