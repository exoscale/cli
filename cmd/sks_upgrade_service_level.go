package cmd

import (
	"fmt"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type sksUpgradeServiceLevelCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"upgrade-service-level"`

	Cluster string `cli-arg:"#" cli-usage:"NAME|ID"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"SKS cluster zone"`
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

func (c *sksUpgradeServiceLevelCmd) cmdRun(_ *cobra.Command, _ []string) error {
	if !c.Force {
		if !askQuestion(fmt.Sprintf(
			"Are you sure you want to upgrade the cluster %q to service level pro?",
			c.Cluster,
		)) {
			return nil
		}
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	cluster, err := cs.FindSKSCluster(ctx, c.Zone, c.Cluster)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Upgrading SKS cluster %q service level...", c.Cluster), func() {
		err = cs.UpgradeSKSClusterServiceLevel(ctx, c.Zone, cluster)
	})
	if err != nil {
		return err
	}

	if !gQuiet {
		return (&sksShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			Cluster:            *cluster.ID,
			Zone:               c.Zone,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(sksCmd, &sksUpgradeServiceLevelCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))

	// FIXME: remove this someday.
	cobra.CheckErr(registerCLICommand(deprecatedSKSCmd, &sksUpgradeServiceLevelCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
