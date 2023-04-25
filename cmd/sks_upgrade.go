package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v2 "github.com/exoscale/egoscale/v2"
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
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	cluster, err := cs.FindSKSCluster(ctx, c.Zone, c.Cluster)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	if !c.Force {
		if utils.VersionIsNewer(c.Version, *cluster.Version) {
			deprecatedResources, err := cs.ListSKSClusterDeprecatedResources(
				ctx,
				c.Zone,
				cluster,
			)
			if err != nil {
				return fmt.Errorf("error retrieving deprecated resources: %w", err)
			}

			removedDeprecatedResources := []*v2.SKSClusterDeprecatedResource{}
			for _, resource := range deprecatedResources {
				if utils.VersionsAreEquivalent(*resource.RemovedRelease, *cluster.Version) {
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

		if !askQuestion(fmt.Sprintf(
			"Are you sure you want to upgrade the cluster %q to version %s?",
			c.Cluster,
			c.Version,
		)) {
			return nil
		}
	}

	decorateAsyncOperation(fmt.Sprintf("Upgrading SKS cluster %q...", c.Cluster), func() {
		err = cs.UpgradeSKSCluster(ctx, c.Zone, cluster, c.Version)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&sksShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			Cluster:            *cluster.ID,
			Zone:               c.Zone,
		}).cmdRun(nil, nil)
	}

	return nil
}

func formatDeprecatedResource(deprecatedResource *v2.SKSClusterDeprecatedResource) string {
	var version string
	var resource string

	if !utils.IsEmptyStringPtr(deprecatedResource.Group) && !utils.IsEmptyStringPtr(deprecatedResource.Version) {
		version = *deprecatedResource.Group + "/" + *deprecatedResource.Version
	}

	if !utils.IsEmptyStringPtr(deprecatedResource.Resource) {
		resource = *deprecatedResource.Resource

		if !utils.IsEmptyStringPtr(deprecatedResource.SubResource) {
			resource += " (" + *deprecatedResource.SubResource + " subresource)"
		}
	}

	deprecationNotice := strings.Join([]string{version, resource}, " ")

	if !utils.IsEmptyStringPtr(deprecatedResource.RemovedRelease) {
		return "Removed in Kubernetes v" + *deprecatedResource.RemovedRelease + ": " + deprecationNotice
	}

	return deprecationNotice
}

func init() {
	cobra.CheckErr(registerCLICommand(sksCmd, &sksUpgradeCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))

	// FIXME: remove this someday.
	cobra.CheckErr(registerCLICommand(deprecatedSKSCmd, &sksUpgradeCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
