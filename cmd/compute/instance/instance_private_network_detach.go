package instance

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
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
	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

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

	utils.DecorateAsyncOperation(fmt.Sprintf(
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
			CliCommandSettings: c.CliCommandSettings,
			Instance:           *instance.ID,
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
