package cmd

import (
	"fmt"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instanceStopCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"stop"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceStopCmd) cmdAliases() []string { return nil }

func (c *instanceStopCmd) cmdShort() string { return "Stop a Compute instance" }

func (c *instanceStopCmd) cmdLong() string { return "" }

func (c *instanceStopCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceStopCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	instance, err := cs.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		return err
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to stop instance %q?", c.Instance)) {
			return nil
		}
	}

	decorateAsyncOperation(fmt.Sprintf("Stopping instance %q...", c.Instance), func() {
		err = cs.StopInstance(ctx, c.Zone, instance)
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(computeInstanceCmd, &instanceStopCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
