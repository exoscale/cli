package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceStartCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"start"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`

	Force         bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	RescueProfile string `cli-usage:"rescue profile to start the instance with"`
	Zone          string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceStartCmd) CmdAliases() []string { return nil }

func (c *instanceStartCmd) CmdShort() string { return "Start a Compute instance" }

func (c *instanceStartCmd) CmdLong() string { return "" }

func (c *instanceStartCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceStartCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	instance, err := globalstate.EgoscaleClient.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to start instance %q?", c.Instance)) {
			return nil
		}
	}

	opts := make([]egoscale.StartInstanceOpt, 0)
	if c.RescueProfile != "" {
		opts = append(opts, egoscale.StartInstanceWithRescueProfile(c.RescueProfile))
	}

	decorateAsyncOperation(fmt.Sprintf("Starting instance %q...", c.Instance), func() {
		err = globalstate.EgoscaleClient.StartInstance(ctx, c.Zone, instance, opts...)
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
	cobra.CheckErr(RegisterCLICommand(instanceCmd, &instanceStartCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
