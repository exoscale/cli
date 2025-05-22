package sks

import (
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type sksUpgradeServiceLevelCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"upgrade-service-level"`

	Cluster string `cli-arg:"#" cli-usage:"NAME|ID"`

	Force bool        `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  v3.ZoneName `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksUpgradeServiceLevelCmd) CmdAliases() []string { return nil }

func (c *sksUpgradeServiceLevelCmd) CmdShort() string {
	return "Upgrade an SKS cluster service level"
}

func (c *sksUpgradeServiceLevelCmd) CmdLong() string {
	return `This command upgrades an SKS cluster's service level to "pro".

Note: once upgraded to pro, an SKS cluster service level cannot be downgraded
to a lower level.`
}

func (c *sksUpgradeServiceLevelCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksUpgradeServiceLevelCmd) confirmUpgrade(cluster string) bool {
	return utils.AskQuestion(
		exocmd.GContext,
		fmt.Sprintf(
			"Are you sure you want to upgrade the cluster %q to service level pro?",
			c.Cluster,
		))
}

func (c *sksUpgradeServiceLevelCmd) CmdRun(_ *cobra.Command, _ []string) error {
	if !c.Force {
		if !c.confirmUpgrade(c.Cluster) {
			return nil
		}
	}

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	resp, err := client.ListSKSClusters(ctx)
	if err != nil {
		return err
	}

	cluster, err := resp.FindSKSCluster(c.Cluster)
	if err != nil {
		return err
	}

	op, err := client.UpgradeSKSClusterServiceLevel(ctx, cluster.ID)

	utils.DecorateAsyncOperation(fmt.Sprintf("Upgrading SKS cluster %q service level...", c.Cluster), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&sksShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Cluster:            cluster.ID.String(),
			Zone:               c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(sksCmd, &sksUpgradeServiceLevelCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
