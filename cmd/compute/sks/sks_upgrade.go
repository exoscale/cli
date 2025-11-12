package sks

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

// TODO: full v3 migration is blocked by
// https://app.shortcut.com/exoscale/story/122943/bug-in-egoscale-v3-listsksclusterdeprecatedresources

type sksUpgradeCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"upgrade"`

	Cluster string `cli-arg:"#" cli-usage:"NAME|ID"`
	Version string `cli-arg:"#"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksUpgradeCmd) CmdAliases() []string { return nil }

func (c *sksUpgradeCmd) CmdShort() string { return "Upgrade an SKS cluster Kubernetes version" }

func (c *sksUpgradeCmd) CmdLong() string { return "" }

func (c *sksUpgradeCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksUpgradeCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	clusters, err := client.ListSKSClusters(ctx)
	if err != nil {
		return err
	}

	cluster, err := clusters.FindSKSCluster(c.Cluster)
	if err != nil {
		return err
	}

	if !c.Force {
		if utils.VersionIsNewer(c.Version, cluster.Version) {
			deprecatedResources, err := client.ListSKSClusterDeprecatedResources(ctx, cluster.ID)
			if err != nil {
				return fmt.Errorf("error retrieving deprecated resources: %w", err)
			}

			removedDeprecatedResources := []v3.SKSClusterDeprecatedResource{}
			for _, resource := range deprecatedResources {
				if utils.VersionsAreEquivalent(resource.RemovedRelease, cluster.Version) {
					removedDeprecatedResources = append(removedDeprecatedResources, resource)
				}
			}

			if len(removedDeprecatedResources) > 0 {
				fmt.Println("Some resources in your cluster are using deprecated APIs:")

				for _, t := range removedDeprecatedResources {
					fmt.Println("- " + formatDeprecatedResource(t))
				}
			}
		}

		if !utils.AskQuestion(
			ctx,
			fmt.Sprintf(
				"Are you sure you want to upgrade the cluster %q to version %s?",
				c.Cluster,
				c.Version,
			)) {
			return nil
		}
	}

	op, err := client.UpgradeSKSCluster(ctx, cluster.ID, v3.UpgradeSKSClusterRequest{
		Version: c.Version,
	})
	utils.DecorateAsyncOperation(fmt.Sprintf("Upgrading SKS cluster %q...", c.Cluster), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&sksShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Cluster:            cluster.ID.String(),
			Zone:               v3.ZoneName(c.Zone),
		}).CmdRun(nil, nil)
	}

	return nil
}

func formatDeprecatedResource(deprecatedResource v3.SKSClusterDeprecatedResource) string {
	var version string
	var resource string

	if deprecatedResource.Group != "" && deprecatedResource.Version != "" {
		version = deprecatedResource.Group + "/" + deprecatedResource.Version
	}

	if deprecatedResource.Resource != "" {
		resource = deprecatedResource.Resource

		if deprecatedResource.Subresource != "" {
			resource += " (" + deprecatedResource.Subresource + " subresource)"
		}
	}

	deprecationNotice := strings.Join([]string{version, resource}, " ")

	if deprecatedResource.RemovedRelease != "" {
		return "Removed in Kubernetes v" + deprecatedResource.RemovedRelease + ": " + deprecationNotice
	}

	return deprecationNotice
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(sksCmd, &sksUpgradeCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
