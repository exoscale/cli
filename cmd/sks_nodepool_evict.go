package cmd

import (
	"errors"
	"fmt"
	"strings"

	exov2 "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type sksNodepoolEvictCmd struct {
	_ bool `cli-cmd:"evict"`

	Cluster  string   `cli-arg:"#" cli-usage:"CLUSTER-NAME|ID"`
	Nodepool string   `cli-arg:"#" cli-usage:"NODEPOOL-NAME|ID"`
	Nodes    []string `cli-arg:"*" cli-usage:"NODE-NAME|ID"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksNodepoolEvictCmd) cmdAliases() []string { return nil }

func (c *sksNodepoolEvictCmd) cmdShort() string { return "Evict SKS cluster Nodepool members" }

func (c *sksNodepoolEvictCmd) cmdLong() string {
	return fmt.Sprintf(`This command evicts specific members from an SKS cluster Nodepool, effectively
scaling down the Nodepool similar to the "exo sks nodepool scale" command.

Note: Kubernetes Nodes should be drained from their workload prior to being
evicted from their Nodepool, e.g. using "kubectl drain".

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&sksNodepoolShowOutput{}), ", "))
}

func (c *sksNodepoolEvictCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksNodepoolEvictCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	if len(c.Nodes) == 0 {
		cmdExitOnUsageError(cmd, "no nodes specified")
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf(
			"Are you sure you want to evict %v from Nodepool %q?",
			c.Nodes,
			c.Nodepool,
		)) {
			return nil
		}
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	cluster, err := cs.FindSKSCluster(ctx, c.Zone, c.Cluster)
	if err != nil {
		return err
	}

	var nodepool *exov2.SKSNodepool
	for _, n := range cluster.Nodepools {
		if *n.ID == c.Nodepool || *n.Name == c.Nodepool {
			nodepool = n
			break
		}
	}
	if nodepool == nil {
		return errors.New("Nodepool not found") // nolint:golint
	}

	nodes := make([]string, len(c.Nodes))
	for i, n := range c.Nodes {
		instance, err := cs.FindInstance(ctx, c.Zone, n)
		if err != nil {
			return fmt.Errorf("invalid Node %q: %s", n, err)
		}
		nodes[i] = *instance.ID
	}

	decorateAsyncOperation(fmt.Sprintf("Evicting Nodes from Nodepool %q...", c.Nodepool), func() {
		err = cluster.EvictNodepoolMembers(ctx, nodepool, nodes)
	})
	if err != nil {
		return err
	}

	if !gQuiet {
		return output(showSKSNodepool(c.Zone, *cluster.ID, *nodepool.ID))
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(sksNodepoolCmd, &sksNodepoolEvictCmd{}))
}
