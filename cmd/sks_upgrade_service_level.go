package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

type sksUpgradeServiceLevelCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"upgrade-service-level"`

	Cluster string `cli-arg:"#" cli-usage:"NAME|ID"`

	Force bool        `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  v3.ZoneName `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksUpgradeServiceLevelCmd) cmdAliases() []string { return nil }

func (c *sksUpgradeServiceLevelCmd) cmdShort() string {
	return "Upgrade an SKS cluster service level"
}

func (c *sksUpgradeServiceLevelCmd) cmdLong() string {
	return `This command upgrades an SKS cluster's service level to "pro".

Note: once upgraded to pro, an SKS cluster service level cannot be downgraded
to a lower level.`
}

func (c *sksUpgradeServiceLevelCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksUpgradeServiceLevelCmd) confirmUpgrade(cluster string) bool {
	return askQuestion(fmt.Sprintf(
		"Are you sure you want to upgrade the cluster %q to service level pro?",
		c.Cluster,
	))
}

func (c *sksUpgradeServiceLevelCmd) cmdRun(_ *cobra.Command, _ []string) error {
	if !c.Force {
		if !c.confirmUpgrade(c.Cluster) {
			return nil
		}
	}

	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
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

	decorateAsyncOperation(fmt.Sprintf("Upgrading SKS cluster %q service level...", c.Cluster), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&sksShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			Cluster:            cluster.ID.String(),
			Zone:               c.Zone,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(sksCmd, &sksUpgradeServiceLevelCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
