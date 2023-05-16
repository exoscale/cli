package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	exoapi "github.com/exoscale/egoscale/v2/api"
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
		strings.Join(output.TemplateAnnotations(&instanceShowOutput{}), ", "),
	)
}

func (c *instancePrivnetDetachCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instancePrivnetDetachCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

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

	decorateAsyncOperation(fmt.Sprintf(
		"Detaching instance %q from Private Network %q...",
		c.Instance,
		c.PrivateNetwork,
	), func() {
		if err = globalstate.EgoscaleClient.DetachInstanceFromPrivateNetwork(ctx, c.Zone, instance, privateNetwork); err != nil {
			return
		}
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&instanceShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			Instance:           *instance.ID,
			Zone:               c.Zone,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(instancePrivnetCmd, &instancePrivnetDetachCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
