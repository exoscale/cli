package cmd

import (
	"fmt"

	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instanceStartCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"start"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`

	Force         bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	RescueProfile string `cli-usage:"rescue profile to start the instance with"`
	Zone          string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceStartCmd) cmdAliases() []string { return nil }

func (c *instanceStartCmd) cmdShort() string { return "Start a Compute instance" }

func (c *instanceStartCmd) cmdLong() string { return "" }

func (c *instanceStartCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceStartCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	instance, err := cs.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
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
		err = cs.StartInstance(ctx, c.Zone, instance, opts...)
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceCmd, &instanceStartCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
