package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type sksNodepoolScaleCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"scale"`

	Cluster  string `cli-arg:"#" cli-usage:"CLUSTER-NAME|ID"`
	Nodepool string `cli-arg:"#" cli-usage:"NODEPOOL-NAME|ID"`
	Size     int64  `cli-arg:"#"`

	Force bool        `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  v3.ZoneName `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksNodepoolScaleCmd) cmdAliases() []string { return nil }

func (c *sksNodepoolScaleCmd) cmdShort() string { return "Scale an SKS cluster Nodepool size" }

func (c *sksNodepoolScaleCmd) cmdLong() string {
	return fmt.Sprintf(`This command scales an SKS cluster Nodepool size up (growing) or down
(shrinking).

In case of a scale-down, operators should use the
"exo compute sks nodepool evict" command, allowing them to specify which
specific Nodes should be evicted from the pool rather than leaving the
decision to the SKS manager.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&sksNodepoolShowOutput{}), ", "))
}

func (c *sksNodepoolScaleCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksNodepoolScaleCmd) cmdRun(_ *cobra.Command, _ []string) error {
	if c.Size <= 0 {
		return errors.New("minimum Nodepool size is 1")
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to scale Nodepool %q to %d?", c.Nodepool, c.Size)) {
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

	decorateAsyncOperation(fmt.Sprintf("Scaling Nodepool %q...", c.Nodepool), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&sksNodepoolShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			Cluster:            cluster.ID.String(),
			Nodepool:           nodepool.ID.String(),
			Zone:               c.Zone,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(sksNodepoolCmd, &sksNodepoolScaleCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
