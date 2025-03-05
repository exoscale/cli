package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	v3 "github.com/exoscale/egoscale/v3"
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

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	cluster, err := globalstate.EgoscaleClient.FindSKSCluster(ctx, c.Zone, c.Cluster)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
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
		return errors.New("nodepool not found")
	}

	decorateAsyncOperation(fmt.Sprintf("Scaling Nodepool %q...", c.Nodepool), func() {
		err = globalstate.EgoscaleClient.ScaleSKSNodepool(ctx, c.Zone, cluster, nodepool, c.Size)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&sksNodepoolShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			Cluster:            *cluster.ID,
			Nodepool:           *nodepool.ID,
			Zone:               v3.ZoneName(c.Zone),
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(sksNodepoolCmd, &sksNodepoolScaleCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
