package cmd

import (
	"errors"
	"fmt"
	"strings"

	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type sksNodepoolScaleCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"scale"`

	Cluster  string `cli-arg:"#" cli-usage:"CLUSTER-NAME|ID"`
	Nodepool string `cli-arg:"#" cli-usage:"NODEPOOL-NAME|ID"`
	Size     int64  `cli-arg:"#"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"SKS cluster zone"`
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
		strings.Join(outputterTemplateAnnotations(&sksNodepoolShowOutput{}), ", "))
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

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	cluster, err := cs.FindSKSCluster(ctx, c.Zone, c.Cluster)
	if err != nil {
		return err
	}

	var nodepool *egoscale.SKSNodepool
	for _, n := range cluster.Nodepools {
		if *n.ID == c.Nodepool || *n.Name == c.Nodepool {
			nodepool = n
			break
		}
	}
	if nodepool == nil {
		return errors.New("Nodepool not found") // nolint:golint
	}

	decorateAsyncOperation(fmt.Sprintf("Scaling Nodepool %q...", c.Nodepool), func() {
		err = cs.ScaleSKSNodepool(ctx, c.Zone, cluster, nodepool, c.Size)
	})
	if err != nil {
		return err
	}

	if !gQuiet {
		return (&sksNodepoolShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			Cluster:            *cluster.ID,
			Nodepool:           *nodepool.ID,
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
