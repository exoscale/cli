package cmd

import (
	"fmt"
	"net"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instancePrivnetUpdateIPCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update-ip"`

	Instance       string `cli-arg:"#" cli-usage:"INSTANCE-NAME|ID"`
	PrivateNetwork string `cli-arg:"#" cli-usage:"PRIVATE-NETWORK-NAME|ID"`

	Zone string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instancePrivnetUpdateIPCmd) cmdAliases() []string { return nil }

func (c *instancePrivnetUpdateIPCmd) cmdShort() string {
	return "Update Private Network IP address of a Compute instance"
}

func (c *instancePrivnetUpdateIPCmd) cmdLong() string {
	return fmt.Sprintf(`This command updates the IP address assigned to a Compute instance in a
managed Private Network.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&instanceShowOutput{}), ", "),
	)
}

func (c *instancePrivnetUpdateIPCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instancePrivnetUpdateIPCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	instance, err := cs.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		return err
	}

	privateNetwork, err := cs.FindPrivateNetwork(ctx, c.Zone, c.PrivateNetwork)
	if err != nil {
		return fmt.Errorf("error retrieving Private Network: %s", err)
	}

	var instanceIPAddress net.IP
	for _, lease := range privateNetwork.Leases {
		if *lease.InstanceID == *instance.ID {
			instanceIPAddress = *lease.IPAddress
			break
		}
	}
	if instanceIPAddress == nil {
		return fmt.Errorf("instance %q has no IP address in Private Network %q", c.Instance, c.PrivateNetwork)
	}

	decorateAsyncOperation(fmt.Sprintf("Updating instance %q Private Network IP address...", c.Instance), func() {
		if err = privateNetwork.UpdateInstanceIPAddress(ctx, instance, instanceIPAddress); err != nil {
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
	cobra.CheckErr(registerCLICommand(computeInstancePrivnetCmd, &instancePrivnetUpdateIPCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
