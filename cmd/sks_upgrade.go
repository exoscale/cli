package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type sksUpgradeCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"upgrade"`

	Cluster string `cli-arg:"#" cli-usage:"NAME|ID"`
	Version string `cli-arg:"#"`

	Force bool        `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  v3.ZoneName `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksUpgradeCmd) cmdAliases() []string { return nil }

func (c *sksUpgradeCmd) cmdShort() string { return "Upgrade an SKS cluster Kubernetes version" }

func (c *sksUpgradeCmd) cmdLong() string { return "" }

func (c *sksUpgradeCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksUpgradeCmd) cmdRun(_ *cobra.Command, _ []string) error {
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

	if !c.Force {
		if err := c.checkDeprecatedResources(ctx, client, cluster); err != nil {
			return err
		}

		if !c.confirmUpgrade(cluster) {
			return nil
		}
	}

	upgradeReq := v3.UpgradeSKSClusterRequest{
		Version: c.Version,
	}

	op, err := client.UpgradeSKSCluster(ctx, cluster.ID, upgradeReq)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Upgrading SKS cluster %q...", c.Cluster), func() {
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
func (c *sksUpgradeCmd) checkDeprecatedResources(ctx context.Context, client *v3.Client, cluster v3.SKSCluster) error {
	if !utils.VersionIsNewer(c.Version, cluster.Version) {
		return nil
	}

	deprecatedResources, err := client.ListSKSClusterDeprecatedResources(ctx, cluster.ID)
	if err != nil {
		return fmt.Errorf("error retrieving deprecated resources: %w", err)
	}

	removedDeprecatedResources := []*v3.SKSClusterDeprecatedResource{}
	for _, resource := range deprecatedResources {
		removed_release, exists := resource["removed_release"]
		if exists && utils.VersionsAreEquivalent(removed_release, cluster.Version) {
			newResource := v3.SKSClusterDeprecatedResource{"removed_release": removed_release}
			removedDeprecatedResources = append(removedDeprecatedResources, &newResource)
		}
	}

	if len(removedDeprecatedResources) > 0 {
		fmt.Println("Some resources in your cluster are using deprecated APIs:")
		for _, t := range removedDeprecatedResources {
			fmt.Println("- " + c.formatDeprecatedResource(t))
		}
	}

	return nil
}

func (c *sksUpgradeCmd) confirmUpgrade(cluster v3.SKSCluster) bool {
	return askQuestion(fmt.Sprintf(
		"Are you sure you want to upgrade the cluster %q to version %s?",
		c.Cluster,
		c.Version,
	))
}

func (c *sksUpgradeCmd) formatDeprecatedResource(deprecatedResource *v3.SKSClusterDeprecatedResource) string {
	var versionStr string
	var resourceStr string
	data := *deprecatedResource

	// Extract values from the map, with empty string as fallback
	group := data["group"]
	version := data["version"]
	resource := data["resource"]
	subResource := data["subresource"]
	removedRelease := data["removed_release"]

	if group != "" && version != "" {
		versionStr = group + "/" + version
	}

	if resource != "" {
		resourceStr = resource

		if subResource != "" {
			resourceStr += " (" + subResource + " subresource)"
		}
	}

	deprecationNotice := strings.Join([]string{versionStr, resourceStr}, " ")

	if removedRelease != "" {
		return "Removed in Kubernetes v" + removedRelease + ": " + deprecationNotice
	}

	return deprecationNotice
}

func init() {
	cobra.CheckErr(registerCLICommand(sksCmd, &sksUpgradeCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
