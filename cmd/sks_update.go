package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type sksUpdateCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	Cluster string `cli-arg:"#" cli-usage:"NAME|ID"`

	AutoUpgrade    bool              `cli-usage:"enable automatic upgrading of the SKS cluster control plane Kubernetes version(--auto-upgrade=false to disable again)"`
	Description    string            `cli-usage:"SKS cluster description"`
	FeatureGates   []string          `cli-flag:"feature-gates" cli-usage:"SKS cluster feature gates to enable"`
	Labels         map[string]string `cli-flag:"label" cli-usage:"SKS cluster label (format: key=value)"`
	Name           string            `cli-usage:"SKS cluster name"`
	EnableCSIAddon bool              `cli-usage:"enable the Exoscale CSI driver"`
	Zone           v3.ZoneName       `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksUpdateCmd) CmdAliases() []string { return nil }

func (c *sksUpdateCmd) CmdShort() string { return "Update an SKS cluster" }

func (c *sksUpdateCmd) CmdLong() string {
	return fmt.Sprintf(`This command updates an SKS cluster.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&sksShowOutput{}), ", "),
	)
}

func (c *sksUpdateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksUpdateCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := GContext
	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
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

	updateReq := v3.UpdateSKSClusterRequest{}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.AutoUpgrade)) {
		updateReq.AutoUpgrade = &c.AutoUpgrade
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.FeatureGates)) {
		updateReq.FeatureGates = c.FeatureGates
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.Labels)) {
		updateReq.Labels = c.Labels
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.Name)) {
		updateReq.Name = c.Name
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.Description)) {
		updateReq.Description = &c.Description
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.EnableCSIAddon)) && !slices.Contains(cluster.Addons, sksClusterAddonExoscaleCSI) {
		updateReq.Addons = append(cluster.Addons, sksClusterAddonExoscaleCSI) //nolint:gocritic
		updated = true
	}

	if updated {
		op, err := client.UpdateSKSCluster(ctx, cluster.ID, updateReq)
		if err != nil {
			return err
		}

		decorateAsyncOperation(fmt.Sprintf("Updating SKS cluster %q...", c.Cluster), func() {
			_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		})

		if err != nil {
			return err
		}
	}

	if !globalstate.Quiet {
		return (&sksShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Cluster:            string(cluster.ID),
			Zone:               c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(RegisterCLICommand(sksCmd, &sksUpdateCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
