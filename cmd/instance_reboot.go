package cmd

import (
	"fmt"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instanceRebootCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"reboot"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceRebootCmd) cmdAliases() []string { return nil }

func (c *instanceRebootCmd) cmdShort() string { return "Reboot a Compute instance" }

func (c *instanceRebootCmd) cmdLong() string { return "" }

func (c *instanceRebootCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceRebootCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	instance, err := cs.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		return err
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to reboot instance %q?", c.Instance)) {
			return nil
		}
	}

	decorateAsyncOperation(fmt.Sprintf("Rebooting instance %q...", c.Instance), func() {
		err = instance.Reboot(ctx)
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(computeInstanceCmd, &instanceRebootCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
