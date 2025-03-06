package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

const (
	kubeletImageGcLowThreshold  = 80
	kubeletImageGcHighThreshold = 85
	kubeletImageGcMinAge        = "2m"
)

type sksNodepoolAddCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"add"`

	Cluster string `cli-arg:"#" cli-usage:"CLUSTER-NAME|ID"`
	Name    string `cli-arg:"#" cli-usage:"NODEPOOL-NAME"`

	AntiAffinityGroups   []string `cli-flag:"anti-affinity-group" cli-usage:"Nodepool Anti-Affinity Group NAME|ID (can be specified multiple times)"`
	DeployTarget         string   `cli-usage:"Nodepool Deploy Target NAME|ID"`
	Description          string   `cli-usage:"Nodepool description"`
	DiskSize             int64    `cli-usage:"Nodepool Compute instances disk size"`
	ImageGcLowThreshold  int64    `cli-flag:"image-gc-low-threshold" cli-usage:"the percent of disk usage after which image garbage collection is never run"`
	ImageGcHighThreshold int64    `cli-flag:"image-gc-high-threshold" cli-usage:"the percent of disk usage after which image garbage collection is always run"`
	ImageGcMinAge        string   `cli-flag:"image-gc-min-age" cli-usage:"maximum age an image can be unused before it is garbage collected"`
	InstancePrefix       string   `cli-usage:"string to prefix Nodepool member names with"`
	InstanceType         string   `cli-usage:"Nodepool Compute instances type"`
	Labels               []string `cli-flag:"label" cli-usage:"Nodepool label (format: key=value)"`
	PrivateNetworks      []string `cli-flag:"private-network" cli-usage:"Nodepool Private Network NAME|ID (can be specified multiple times)"`
	SecurityGroups       []string `cli-flag:"security-group" cli-usage:"Nodepool Security Group NAME|ID (can be specified multiple times)"`
	Size                 int64    `cli-usage:"Nodepool size"`
	StorageLvm           bool     `cli-usage:"Create nodes with non-standard partitioning for persistent storage"`
	Taints               []string `cli-flag:"taint" cli-usage:"Kubernetes taint to apply to Nodepool Nodes (format: KEY=VALUE:EFFECT, can be specified multiple times)"`
	Zone                 string   `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksNodepoolAddCmd) cmdAliases() []string { return nil }

func (c *sksNodepoolAddCmd) cmdShort() string { return "Add a Nodepool to an SKS cluster" }

func (c *sksNodepoolAddCmd) cmdLong() string {
	return fmt.Sprintf(`This command adds a Nodepool to an SKS cluster.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&sksNodepoolShowOutput{}), ", "))
}

func (c *sksNodepoolAddCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksNodepoolAddCmd) cmdRun(_ *cobra.Command, _ []string) error {
	nodepoolReq := v3.CreateSKSNodepoolRequest{
		Description:    c.Description,
		DiskSize:       c.DiskSize,
		InstancePrefix: c.InstancePrefix,
		Name:           c.Name,
		Size:           c.Size,
		KubeletImageGC: &v3.KubeletImageGC{
			MinAge:        c.ImageGcMinAge,
			LowThreshold:  c.ImageGcLowThreshold,
			HighThreshold: c.ImageGcHighThreshold,
		},
	}

	ctx := gContext

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
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

	addOns := map[string]bool{
		"storage-lvm": c.StorageLvm,
	}

	nodepoolReq.Addons = []string{}
	for k, v := range addOns {
		if v {
			nodepoolReq.Addons = append(nodepoolReq.Addons, k)
		}
	}

	if l := len(c.AntiAffinityGroups); l > 0 {
		nodepoolReq.AntiAffinityGroups = make([]v3.AntiAffinityGroup, l)
		for i, v := range c.AntiAffinityGroups {
			antiAffinityGroupList, err := client.ListAntiAffinityGroups(ctx)
			if err != nil {
				return err
			}
			aaG, err := antiAffinityGroupList.FindAntiAffinityGroup(v)
			if err != nil {
				return fmt.Errorf("error retrieving Anti-Affinity Group: %w", err)
			}
			nodepoolReq.AntiAffinityGroups[i] = aaG
		}
	}

	if c.DeployTarget != "" {
		deployTargetList, err := client.ListDeployTargets(ctx)
		if err != nil {
			return err
		}
		deployTarget, err := deployTargetList.FindDeployTarget(c.DeployTarget)
		if err != nil {
			return fmt.Errorf("error retrieving Deploy Target: %w", err)
		}
		nodepoolReq.DeployTarget = &deployTarget
	}

	nodepoolInstanceTypeList, err := client.ListInstanceTypes(ctx)
	if err != nil {
		return err
	}
	nodepoolInstanceType, err := nodepoolInstanceTypeList.FindInstanceTypeByIdOrFamilyAndSize(c.InstanceType)
	if err != nil {
		return fmt.Errorf("error retrieving instance type: %w", err)
	}
	nodepoolReq.InstanceType = &nodepoolInstanceType

	if len(c.Labels) > 0 {
		labels := make(map[string]string)
		if len(c.Labels) > 0 {
			labels, err = utils.SliceToMap(c.Labels)
			if err != nil {
				return fmt.Errorf("label: %w", err)
			}
		}
		nodepoolReq.Labels = labels
	}

	if l := len(c.PrivateNetworks); l > 0 {
		nodepoolPrivateNetworks := make([]v3.PrivateNetwork, l)
		for i, v := range c.PrivateNetworks {
			privateNetworksList, err := client.ListPrivateNetworks(ctx)
			if err != nil {
				return err
			}
			privateNetwork, err := privateNetworksList.FindPrivateNetwork(v)
			if err != nil {
				return fmt.Errorf("error retrieving Private Network: %w", err)
			}
			nodepoolPrivateNetworks[i] = privateNetwork
		}
		nodepoolReq.PrivateNetworks = nodepoolPrivateNetworks
	}

	if l := len(c.SecurityGroups); l > 0 {
		nodepoolSecurityGroups := make([]v3.SecurityGroup, l)
		for i, v := range c.SecurityGroups {
			securityGroupList, err := client.ListSecurityGroups(ctx)
			if err != nil {
				return err
			}
			securityGroup, err := securityGroupList.FindSecurityGroup(v)
			if err != nil {
				return fmt.Errorf("error retrieving Security Group: %w", err)
			}
			nodepoolSecurityGroups[i] = securityGroup
		}
		nodepoolReq.SecurityGroups = nodepoolSecurityGroups
	}

	if len(c.Taints) > 0 {
		taints := make(v3.SKSNodepoolTaints)
		for _, t := range c.Taints {
			key, taint, err := parseSKSNodepoolTaintV3(t)
			if err != nil {
				return fmt.Errorf("invalid taint value %q: %w", t, err)
			}
			taints[key] = *taint
		}
		nodepoolReq.Taints = taints
	}

	op, err := client.CreateSKSNodepool(ctx, cluster.ID, nodepoolReq)
	if err != nil {
		return err
	}
	decorateAsyncOperation(fmt.Sprintf("Adding Nodepool %q...", nodepoolReq.Name), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&sksNodepoolShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			Cluster:            cluster.ID.String(),
			Nodepool:           op.Reference.ID.String(),
			Zone:               v3.ZoneName(c.Zone),
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(sksNodepoolCmd, &sksNodepoolAddCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		Size:                 2,
		InstanceType:         fmt.Sprintf("%s.%s", defaultInstanceTypeFamily, defaultInstanceType),
		DiskSize:             50,
		ImageGcLowThreshold:  kubeletImageGcLowThreshold,
		ImageGcHighThreshold: kubeletImageGcHighThreshold,
		ImageGcMinAge:        kubeletImageGcMinAge,
	}))
}
