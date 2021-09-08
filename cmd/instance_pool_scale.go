package cmd

import (
	"errors"
	"fmt"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instancePoolScaleCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"scale"`

	InstancePool string `cli-arg:"#" cli-usage:"INSTANCE-POOL-NAME|ID"`
	Size         int64  `cli-arg:"#"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"Instance Pool zone"`
}

func (c *instancePoolScaleCmd) cmdAliases() []string { return nil }

func (c *instancePoolScaleCmd) cmdShort() string { return "Scale an Instance Pool size" }

func (c *instancePoolScaleCmd) cmdLong() string {
	return `This command scales an Instance Pool size up (growing) or down
(shrinking).

In case of a scale-down, operators should use the "exo instancepool evict"
variant, allowing them to specify which specific instance should be evicted
from the Instance Pool rather than leaving the decision to the orchestrator.`
}

func (c *instancePoolScaleCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instancePoolScaleCmd) cmdRun(_ *cobra.Command, _ []string) error {
	if c.Size <= 0 {
		return errors.New("minimum Instance Pool size is 1")
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf(
			"Are you sure you want to scale Instance Pool %q to %d?",
			c.InstancePool,
			c.Size,
		)) {
			return nil
		}
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	instancePool, err := cs.FindInstancePool(ctx, c.Zone, c.InstancePool)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Scaling Instance Pool %q...", c.InstancePool), func() {
		err = cs.ScaleInstancePool(ctx, c.Zone, instancePool, c.Size)
	})
	if err != nil {
		return err
	}

	if !gQuiet {
		return output(showInstancePool(c.Zone, *instancePool.ID))
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(instancePoolCmd, &instancePoolScaleCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))

	// FIXME: remove this someday.
	cobra.CheckErr(registerCLICommand(deprecatedInstancePoolCmd, &instancePoolScaleCmd{}))
}
