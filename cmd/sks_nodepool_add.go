package cmd

import (
	"fmt"
	"strings"

	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type sksNodepoolAddCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"add"`

	Cluster string `cli-arg:"#" cli-usage:"CLUSTER-NAME|ID"`
	Name    string `cli-arg:"#" cli-usage:"NODEPOOL-NAME"`

	AntiAffinityGroups []string `cli-flag:"anti-affinity-group" cli-usage:"Nodepool Anti-Affinity Group NAME|ID (can be specified multiple times)"`
	DeployTarget       string   `cli-usage:"Nodepool Deploy Target NAME|ID"`
	Description        string   `cli-usage:"Nodepool description"`
	DiskSize           int64    `cli-usage:"Nodepool Compute instances disk size"`
	InstancePrefix     string   `cli-usage:"string to prefix Nodepool member names with"`
	InstanceType       string   `cli-usage:"Nodepool Compute instances type"`
	Labels             []string `cli-flag:"label" cli-usage:"Nodepool label (format: key=value)"`
	PrivateNetworks    []string `cli-flag:"private-network" cli-usage:"Nodepool Private Network NAME|ID (can be specified multiple times)"`
	SecurityGroups     []string `cli-flag:"security-group" cli-usage:"Nodepool Security Group NAME|ID (can be specified multiple times)"`
	Size               int64    `cli-usage:"Nodepool size"`
	Taints             []string `cli-flag:"taint" cli-usage:"Kubernetes taint to apply to Nodepool Nodes (format: KEY=VALUE:EFFECT, can be specified multiple times)"`
	Zone               string   `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksNodepoolAddCmd) cmdAliases() []string { return nil }

func (c *sksNodepoolAddCmd) cmdShort() string { return "Add a Nodepool to an SKS cluster" }

func (c *sksNodepoolAddCmd) cmdLong() string {
	return fmt.Sprintf(`This command adds a Nodepool to an SKS cluster.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&sksNodepoolShowOutput{}), ", "))
}

func (c *sksNodepoolAddCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksNodepoolAddCmd) cmdRun(_ *cobra.Command, _ []string) error {
	nodepool := &egoscale.SKSNodepool{
		Description:    nonEmptyStringPtr(c.Description),
		DiskSize:       &c.DiskSize,
		InstancePrefix: nonEmptyStringPtr(c.InstancePrefix),
		Name:           &c.Name,
		Size:           &c.Size,
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	cluster, err := cs.FindSKSCluster(ctx, c.Zone, c.Cluster)
	if err != nil {
		return fmt.Errorf("error retrieving cluster: %w", err)
	}

	if l := len(c.AntiAffinityGroups); l > 0 {
		nodepoolAntiAffinityGroupIDs := make([]string, l)
		for i := range c.AntiAffinityGroups {
			antiAffinityGroup, err := cs.FindAntiAffinityGroup(ctx, c.Zone, c.AntiAffinityGroups[i])
			if err != nil {
				return fmt.Errorf("error retrieving Anti-Affinity Group: %w", err)
			}
			nodepoolAntiAffinityGroupIDs[i] = *antiAffinityGroup.ID
		}
		nodepool.AntiAffinityGroupIDs = &nodepoolAntiAffinityGroupIDs
	}

	if c.DeployTarget != "" {
		deployTarget, err := cs.FindDeployTarget(ctx, c.Zone, c.DeployTarget)
		if err != nil {
			return fmt.Errorf("error retrieving Deploy Target: %w", err)
		}
		nodepool.DeployTargetID = deployTarget.ID
	}

	nodepoolInstanceType, err := cs.FindInstanceType(ctx, c.Zone, c.InstanceType)
	if err != nil {
		return fmt.Errorf("error retrieving instance type: %w", err)
	}
	nodepool.InstanceTypeID = nodepoolInstanceType.ID

	if len(c.Labels) > 0 {
		labels := make(map[string]string)
		if len(c.Labels) > 0 {
			labels, err = sliceToMap(c.Labels)
			if err != nil {
				return fmt.Errorf("label: %w", err)
			}
		}
		nodepool.Labels = &labels
	}

	if l := len(c.PrivateNetworks); l > 0 {
		nodepoolPrivateNetworkIDs := make([]string, l)
		for i := range c.PrivateNetworks {
			privateNetwork, err := cs.FindPrivateNetwork(ctx, c.Zone, c.PrivateNetworks[i])
			if err != nil {
				return fmt.Errorf("error retrieving Private Network: %w", err)
			}
			nodepoolPrivateNetworkIDs[i] = *privateNetwork.ID
		}
		nodepool.PrivateNetworkIDs = &nodepoolPrivateNetworkIDs
	}

	if l := len(c.SecurityGroups); l > 0 {
		nodepoolSecurityGroupIDs := make([]string, l)
		for i := range c.SecurityGroups {
			securityGroup, err := cs.FindSecurityGroup(ctx, c.Zone, c.SecurityGroups[i])
			if err != nil {
				return fmt.Errorf("error retrieving Security Group: %w", err)
			}
			nodepoolSecurityGroupIDs[i] = *securityGroup.ID
		}
		nodepool.SecurityGroupIDs = &nodepoolSecurityGroupIDs
	}

	if len(c.Taints) > 0 {
		taints := make(map[string]*egoscale.SKSNodepoolTaint)
		for _, t := range c.Taints {
			key, taint, err := parseSKSNodepoolTaint(t)
			if err != nil {
				return fmt.Errorf("invalid taint value %q: %w", t, err)
			}
			taints[key] = taint
		}
		nodepool.Taints = &taints
	}

	decorateAsyncOperation(fmt.Sprintf("Adding Nodepool %q...", *nodepool.Name), func() {
		nodepool, err = cs.CreateSKSNodepool(ctx, c.Zone, cluster, nodepool)
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
	cobra.CheckErr(registerCLICommand(sksNodepoolCmd, &sksNodepoolAddCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		Size:         2,
		InstanceType: defaultServiceOffering,
		DiskSize:     50,
	}))
}
