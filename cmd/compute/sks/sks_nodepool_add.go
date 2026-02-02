package sks

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
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
	exocmd.CliCommandSettings `cli-cmd:"-"`

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
	PublicIPAssignment   string   `cli-flag:"public-ip" cli-usage:"Configures public IP assignment of the Instances (inet4|dual). (default: inet4)"`
}

func (c *sksNodepoolAddCmd) CmdAliases() []string { return nil }

func (c *sksNodepoolAddCmd) CmdShort() string { return "Add a Nodepool to an SKS cluster" }

func (c *sksNodepoolAddCmd) CmdLong() string {
	return fmt.Sprintf(`This command adds a Nodepool to an SKS cluster.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&sksNodepoolShowOutput{}), ", "))
}

func (c *sksNodepoolAddCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksNodepoolAddCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
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

	labels := make(map[string]string)
	if len(c.Labels) > 0 {
		labels, err = utils.SliceToMap(c.Labels)
		if err != nil {
			return fmt.Errorf("label: %w", err)
		}
	}
	publicIPAssignment := v3.CreateSKSNodepoolRequestPublicIPAssignmentInet4
	if c.PublicIPAssignment != "" {
		if !slices.Contains([]v3.CreateSKSNodepoolRequestPublicIPAssignment{
			v3.CreateSKSNodepoolRequestPublicIPAssignmentInet4, v3.CreateSKSNodepoolRequestPublicIPAssignmentDual,
		}, v3.CreateSKSNodepoolRequestPublicIPAssignment(c.PublicIPAssignment)) {
			return fmt.Errorf("error invalid public-ip: %s", c.PublicIPAssignment)
		}
		publicIPAssignment = v3.CreateSKSNodepoolRequestPublicIPAssignment(c.PublicIPAssignment)
	}

	opts := CreateNodepoolOpts{
		Name:               c.Name,
		Description:        c.Description,
		DiskSize:           c.DiskSize,
		InstancePrefix:     c.InstancePrefix,
		Size:               c.Size,
		InstanceType:       c.InstanceType,
		Labels:             labels,
		AntiAffinityGroups: c.AntiAffinityGroups,
		DeployTarget:       c.DeployTarget,
		PrivateNetworks:    c.PrivateNetworks,
		SecurityGroups:     c.SecurityGroups,
		Taints:             c.Taints,
		KubeletImageGC: &v3.KubeletImageGC{
			MinAge:        c.ImageGcMinAge,
			LowThreshold:  c.ImageGcLowThreshold,
			HighThreshold: c.ImageGcHighThreshold,
		},
		PublicIPAssignment: publicIPAssignment,
	}

	nodepoolReq, err := createNodepoolRequest(ctx, client, opts)
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

	op, err := client.CreateSKSNodepool(ctx, cluster.ID, nodepoolReq)
	if err != nil {
		return err
	}
	utils.DecorateAsyncOperation(fmt.Sprintf("Adding Nodepool %q...", nodepoolReq.Name), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&sksNodepoolShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Cluster:            cluster.ID.String(),
			Nodepool:           op.Reference.ID.String(),
			Zone:               v3.ZoneName(c.Zone),
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(sksNodepoolCmd, &sksNodepoolAddCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),

		Size:                 2,
		InstanceType:         fmt.Sprintf("%s.%s", exocmd.DefaultInstanceTypeFamily, exocmd.DefaultInstanceType),
		DiskSize:             50,
		ImageGcLowThreshold:  kubeletImageGcLowThreshold,
		ImageGcHighThreshold: kubeletImageGcHighThreshold,
		ImageGcMinAge:        kubeletImageGcMinAge,
	}))
}
