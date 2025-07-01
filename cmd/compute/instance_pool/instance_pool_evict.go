package instance_pool

import (
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type instancePoolEvictCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

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
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instancePoolEvictCmd) CmdRun(cmd *cobra.Command, _ []string) error {

	if len(c.Instances) == 0 {
		exocmd.CmdExitOnUsageError(cmd, "no instances specified")
	}

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	instancePools, err := client.ListInstancePools(ctx)
	if err != nil {
		return err
	}
	instancePool, err := instancePools.FindInstancePool(c.InstancePool)
	if err != nil {
		return err
	}

	if !c.Force {
		if !utils.AskQuestion(
			ctx,
			fmt.Sprintf(
				"Are you sure you want to evict %v from Instance Pool %q?",
				c.Instances,
				c.InstancePool,
			)) {
			return nil
		}
	}

	instanceIDs := make([]v3.UUID, len(c.Instances))
	instances, err := client.ListInstances(ctx)
	if err != nil {
		return err
	}
	for i, n := range c.Instances {
		instance, err := instances.FindListInstancesResponseInstances(n)
		if err != nil {
			return err
		}
		instanceIDs[i] = instance.ID
	}

	op, err := client.EvictInstancePoolMembers(ctx, instancePool.ID, v3.EvictInstancePoolMembersRequest{
		Instances: instanceIDs,
	})
	if err != nil {
		return err

	}
	utils.DecorateAsyncOperation(
		fmt.Sprintf("Evicting instances from Instance Pool %q...", c.InstancePool),
		func() {
			_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		},
	)
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&instancePoolShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Zone:               c.Zone,
			InstancePool:       instancePool.ID.String(),
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instancePoolCmd, &instancePoolEvictCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
