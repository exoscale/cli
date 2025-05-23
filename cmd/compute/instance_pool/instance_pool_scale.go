package instance_pool

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type instancePoolScaleCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"scale"`

	InstancePool string `cli-arg:"#" cli-usage:"INSTANCE-POOL-NAME|ID"`
	Size         int64  `cli-arg:"#"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"Instance Pool zone"`
}

func (c *instancePoolScaleCmd) CmdAliases() []string { return nil }

func (c *instancePoolScaleCmd) CmdShort() string { return "Scale an Instance Pool size" }

func (c *instancePoolScaleCmd) CmdLong() string {
	return `This command scales an Instance Pool size up (growing) or down
(shrinking).

In case of a scale-down, operators should use the
"exo compute instance-pool evict" command, allowing them to specify which
specific instance should be evicted from the Instance Pool rather than leaving
the decision to the orchestrator.`
}

func (c *instancePoolScaleCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instancePoolScaleCmd) CmdRun(_ *cobra.Command, _ []string) error {
	if c.Size <= 0 {
		return errors.New("minimum Instance Pool size is 1")
	}

	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	if !c.Force {
		if !utils.AskQuestion(
			ctx,
			fmt.Sprintf(
				"Are you sure you want to scale Instance Pool %q to %d?",
				c.InstancePool,
				c.Size,
			)) {
			return nil
		}
	}

	instancePool, err := globalstate.EgoscaleClient.FindInstancePool(ctx, c.Zone, c.InstancePool)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Scaling Instance Pool %q...", c.InstancePool), func() {
		err = globalstate.EgoscaleClient.ScaleInstancePool(ctx, c.Zone, instancePool, c.Size)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&instancePoolShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Zone:               c.Zone,
			InstancePool:       *instancePool.ID,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instancePoolCmd, &instancePoolScaleCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
