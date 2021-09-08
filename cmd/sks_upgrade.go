package cmd

import (
	"fmt"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type sksUpgradeCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"upgrade"`

	Cluster string `cli-arg:"#" cli-usage:"NAME|ID"`
	Version string `cli-arg:"#"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksUpgradeCmd) cmdAliases() []string { return nil }

func (c *sksUpgradeCmd) cmdShort() string { return "Upgrade an SKS cluster Kubernetes version" }

func (c *sksUpgradeCmd) cmdLong() string { return "" }

func (c *sksUpgradeCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksUpgradeCmd) cmdRun(_ *cobra.Command, _ []string) error {
	if !c.Force {
		if !askQuestion(fmt.Sprintf(
			"Are you sure you want to upgrade the cluster %q to version %s?",
			c.Cluster,
			c.Version,
		)) {
			return nil
		}
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	cluster, err := cs.FindSKSCluster(ctx, c.Zone, c.Cluster)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Upgrading SKS cluster %q...", c.Cluster), func() {
		err = cs.UpgradeSKSCluster(ctx, c.Zone, cluster, c.Version)
	})
	if err != nil {
		return err
	}

	if !gQuiet {
		return output(showSKSCluster(c.Zone, *cluster.ID))
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(sksCmd, &sksUpgradeCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))

	// FIXME: remove this someday.
	cobra.CheckErr(registerCLICommand(deprecatedSKSCmd, &sksUpgradeCmd{}))
}
