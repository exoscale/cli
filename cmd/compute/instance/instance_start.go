package instance

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type instanceStartCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

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
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceStartCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	instance, err := globalstate.EgoscaleClient.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	if !c.Force {
		if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to start instance %q?", c.Instance)) {
			return nil
		}
	}

	opts := make([]egoscale.StartInstanceOpt, 0)
	if c.RescueProfile != "" {
		opts = append(opts, egoscale.StartInstanceWithRescueProfile(c.RescueProfile))
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Starting instance %q...", c.Instance), func() {
		err = globalstate.EgoscaleClient.StartInstance(ctx, c.Zone, instance, opts...)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&instanceShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Instance:           *instance.ID,
			Zone:               c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceCmd, &instanceStartCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
