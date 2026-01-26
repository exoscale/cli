package sks

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type sksNodepoolScaleCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"scale"`

	Cluster  string `cli-arg:"#" cli-usage:"CLUSTER-NAME|ID"`
	Nodepool string `cli-arg:"#" cli-usage:"NODEPOOL-NAME|ID"`
	Size     int64  `cli-arg:"#"`

	Force bool        `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  v3.ZoneName `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksNodepoolScaleCmd) CmdAliases() []string { return nil }

func (c *sksNodepoolScaleCmd) CmdShort() string { return "Scale an SKS cluster Nodepool size" }

func (c *sksNodepoolScaleCmd) CmdLong() string {
	return fmt.Sprintf(`This command scales an SKS cluster Nodepool size up (growing) or down
(shrinking).

In case of a scale-down, operators should use the
"exo compute sks nodepool evict" command, allowing them to specify which
specific Nodes should be evicted from the pool rather than leaving the
decision to the SKS manager.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&sksNodepoolShowOutput{}), ", "))
}

func (c *sksNodepoolScaleCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksNodepoolScaleCmd) CmdRun(_ *cobra.Command, _ []string) error {
	if c.Size < 0 {
		return errors.New("minimum Nodepool size is 0")
	}

	ctx := exocmd.GContext

	if !c.Force {
		if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to scale Nodepool %q to %d?", c.Nodepool, c.Size)) {
			return nil
		}
	}

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

	var nodepool *v3.SKSNodepool
	for _, n := range cluster.Nodepools {
		if n.ID.String() == c.Nodepool || n.Name == c.Nodepool {
			nodepool = &n
			break
		}
	}
	if nodepool == nil {
		return errors.New("nodepool not found")
	}

	req := v3.ScaleSKSNodepoolRequest{
		Size: c.Size,
	}

	op, err := client.ScaleSKSNodepool(ctx, cluster.ID, nodepool.ID, req)
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Scaling Nodepool %q...", c.Nodepool), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&sksNodepoolShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Cluster:            cluster.ID.String(),
			Nodepool:           nodepool.ID.String(),
			Zone:               c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(sksNodepoolCmd, &sksNodepoolScaleCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
