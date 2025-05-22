package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type instancePoolEvictCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"evict"`

	InstancePool string   `cli-arg:"#" cli-usage:"INSTANCE-POOL-NAME|ID"`
	Instances    []string `cli-arg:"*" cli-usage:"INSTANCE-NAME|ID"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"Instance Pool zone"`
}

func (c *instancePoolEvictCmd) CmdAliases() []string { return nil }

func (c *instancePoolEvictCmd) CmdShort() string { return "Evict Instance Pool members" }

func (c *instancePoolEvictCmd) CmdLong() string {
	return `This command evicts specific members from an Instance Pool, effectively
scaling down the Instance Pool similar to the "exo compute instance-pool scale"
command.`
}

func (c *instancePoolEvictCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instancePoolEvictCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	if len(c.Instances) == 0 {
		CmdExitOnUsageError(cmd, "no instances specified")
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf(
			"Are you sure you want to evict %v from Instance Pool %q?",
			c.Instances,
			c.InstancePool,
		)) {
			return nil
		}
	}

	ctx := exoapi.WithEndpoint(GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	instancePool, err := globalstate.EgoscaleClient.FindInstancePool(ctx, c.Zone, c.InstancePool)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	instances := make([]string, len(c.Instances))
	for i, n := range c.Instances {
		instance, err := globalstate.EgoscaleClient.FindInstance(ctx, c.Zone, n)
		if err != nil {
			return fmt.Errorf("invalid instance %q: %w", n, err)
		}
		instances[i] = *instance.ID
	}

	decorateAsyncOperation(
		fmt.Sprintf("Evicting instances from Instance Pool %q...", c.InstancePool),
		func() {
			err = globalstate.EgoscaleClient.EvictInstancePoolMembers(ctx, c.Zone, instancePool, instances)
		},
	)
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
	cobra.CheckErr(RegisterCLICommand(instancePoolCmd, &instancePoolEvictCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
