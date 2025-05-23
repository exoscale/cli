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

type sksNodepoolEvictCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"evict"`

	Cluster  string   `cli-arg:"#" cli-usage:"CLUSTER-NAME|ID"`
	Nodepool string   `cli-arg:"#" cli-usage:"NODEPOOL-NAME|ID"`
	Nodes    []string `cli-arg:"*" cli-usage:"NODE-NAME|ID"`

	Force bool        `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  v3.ZoneName `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksNodepoolEvictCmd) CmdAliases() []string { return nil }

func (c *sksNodepoolEvictCmd) CmdShort() string { return "Evict SKS cluster Nodepool members" }

func (c *sksNodepoolEvictCmd) CmdLong() string {
	return fmt.Sprintf(`This command evicts specific members from an SKS cluster Nodepool, effectively
scaling down the Nodepool similar to the "exo compute sks nodepool scale"
command.

Note: Kubernetes Nodes should be drained from their workload prior to being
evicted from their Nodepool, e.g. using "kubectl drain".

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&sksNodepoolShowOutput{}), ", "))
}

func (c *sksNodepoolEvictCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksNodepoolEvictCmd) CmdRun(cmd *cobra.Command, _ []string) error {
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

	ctx := GContext
	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
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

	nodes := make([]v3.UUID, len(c.Nodes))
	for i, n := range c.Nodes {
		resp, err := client.ListInstances(ctx)
		if err != nil {
			return err
		}

		instance, err := resp.FindListInstancesResponseInstances(n)
		if err != nil {
			return fmt.Errorf("invalid Node %q: %w", n, err)
		}
		nodes[i] = instance.ID
	}

	req := v3.EvictSKSNodepoolMembersRequest{
		Instances: nodes,
	}

	op, err := client.EvictSKSNodepoolMembers(ctx, cluster.ID, nodepool.ID, req)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Evicting Nodes from Nodepool %q...", c.Nodepool), func() {
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
	cobra.CheckErr(RegisterCLICommand(sksNodepoolCmd, &sksNodepoolEvictCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
