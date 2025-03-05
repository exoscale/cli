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

type sksNodepoolEvictCmd struct {
	cliCommandSettings `cli-cmd:"-"`

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
scaling down the Nodepool similar to the "exo compute sks nodepool scale"
command.

Note: Kubernetes Nodes should be drained from their workload prior to being
evicted from their Nodepool, e.g. using "kubectl drain".

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&sksNodepoolShowOutput{}), ", "))
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

	nodes := make([]string, len(c.Nodes))
	for i, n := range c.Nodes {
		instance, err := globalstate.EgoscaleClient.FindInstance(ctx, c.Zone, n)
		if err != nil {
			return fmt.Errorf("invalid Node %q: %w", n, err)
		}
		nodes[i] = *instance.ID
	}

	decorateAsyncOperation(fmt.Sprintf("Evicting Nodes from Nodepool %q...", c.Nodepool), func() {
		err = globalstate.EgoscaleClient.EvictSKSNodepoolMembers(ctx, c.Zone, cluster, nodepool, nodes)
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
	cobra.CheckErr(registerCLICommand(sksNodepoolCmd, &sksNodepoolEvictCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
