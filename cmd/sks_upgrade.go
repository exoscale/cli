package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v2 "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	v3 "github.com/exoscale/egoscale/v3"
)

// TODO: full v3 migration is blocked by
// https://app.shortcut.com/exoscale/story/122943/bug-in-egoscale-v3-listsksclusterdeprecatedresources

type sksUpgradeCmd struct {
	CliCommandSettings `cli-cmd:"-"`

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
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksUpgradeCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	cluster, err := globalstate.EgoscaleClient.FindSKSCluster(ctx, c.Zone, c.Cluster)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	if !c.Force {
		if utils.VersionIsNewer(c.Version, *cluster.Version) {
			deprecatedResources, err := globalstate.EgoscaleClient.ListSKSClusterDeprecatedResources(
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
		err = globalstate.EgoscaleClient.UpgradeSKSCluster(ctx, c.Zone, cluster, c.Version)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&sksShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Cluster:            *cluster.ID,
			Zone:               v3.ZoneName(c.Zone),
		}).CmdRun(nil, nil)
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
	cobra.CheckErr(RegisterCLICommand(sksCmd, &sksUpgradeCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
