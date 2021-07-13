package cmd

import (
	"errors"
	"fmt"
	"strings"

	exov2 "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

var (
	defaultSKSClusterCNI          = "calico"
	defaultSKSClusterServiceLevel = "pro"
	sksClusterAddonExoscaleCCM    = "exoscale-cloud-controller"
	sksClusterAddonMetricsServer  = "metrics-server"
)

type sksCreateCmd struct {
	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#" cli-usage:"NAME"`

	AutoUpgrade                bool              `cli-usage:"enable automatic upgrading of the SKS cluster control plane Kubernetes version"`
	Description                string            `cli-usage:"SKS cluster description"`
	KubernetesVersion          string            `cli-usage:"SKS cluster control plane Kubernetes version"`
	Labels                     map[string]string `cli-flag:"label" cli-usage:"SKS cluster label (format: key=value)"`
	NoCNI                      bool              `cli-usage:"do not deploy a default Container Network Interface plugin in the cluster control plane"`
	NoExoscaleCCM              bool              `cli-usage:"do not deploy the Exoscale Cloud Controller Manager in the cluster control plane"`
	NoMetricsServer            bool              `cli-usage:"do not deploy the Kubernetes Metrics Server in the cluster control plane"`
	NodepoolAntiAffinityGroups []string          `cli-flag:"nodepool-anti-affinity-group" cli-usage:"default Nodepool Anti-Affinity Group NAME|ID (can be specified multiple times)"`
	NodepoolDeployTarget       string            `cli-usage:"default Nodepool Deploy Target NAME|ID"`
	NodepoolDescription        string            `cli-usage:"default Nodepool description"`
	NodepoolDiskSize           int64             `cli-usage:"default Nodepool Compute instances disk size"`
	NodepoolInstancePrefix     string            `cli-usage:"string to prefix default Nodepool member names with"`
	NodepoolInstanceType       string            `cli-usage:"default Nodepool Compute instances type"`
	NodepoolLabels             map[string]string `cli-flag:"nodepool-label" cli-usage:"default Nodepool label (format: key=value)"`
	NodepoolName               string            `cli-usage:"default Nodepool name"`
	NodepoolSecurityGroups     []string          `cli-flag:"nodepool-security-group" cli-usage:"default Nodepool Security Group NAME|ID (can be specified multiple times)"`
	NodepoolSize               int64             `cli-usage:"default Nodepool size. If 0, no default Nodepool will be added to the cluster."`
	ServiceLevel               string            `cli-usage:"SKS cluster control plane service level (starter|pro)"`
	Zone                       string            `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksCreateCmd) cmdAliases() []string { return gCreateAlias }

func (c *sksCreateCmd) cmdShort() string { return "Create an SKS cluster" }

func (c *sksCreateCmd) cmdLong() string {
	return fmt.Sprintf(`This command creates an SKS cluster.

Note: SKS cluster Nodes' kubelet configuration is set to use the Exoscale
Cloud Controller Manager (CCM) as Cloud Provider by default. Cluster Nodes
will remain in the "NotReady" status until the Exoscale CCM is deployed by
cluster operators. Please refer to the Exoscale CCM documentation for more
information:

    https://github.com/exoscale/exoscale-cloud-controller-manager

If you do not want to use a Cloud Controller Manager, add the
"--no-exoscale-ccm" option to the command. This cannot be changed once the
cluster has been created.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&sksShowOutput{}), ", "))
}

func (c *sksCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksCreateCmd) cmdRun(_ *cobra.Command, _ []string) error {
	cluster := &exov2.SKSCluster{
		AutoUpgrade: &c.AutoUpgrade,
		CNI:         &defaultSKSClusterCNI,
		Description: func() (v *string) {
			if c.Description != "" {
				v = &c.Description
			}
			return
		}(),
		Labels: func() (v *map[string]string) {
			if len(c.Labels) > 0 {
				return &c.Labels
			}
			return
		}(),
		Name:         &c.Name,
		ServiceLevel: &c.ServiceLevel,
		Version:      &c.KubernetesVersion,
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	if c.NoCNI {
		cluster.CNI = nil
	}

	addOns := map[string]struct{}{
		sksClusterAddonExoscaleCCM:   {},
		sksClusterAddonMetricsServer: {},
	}
	cluster.AddOns = func() (v *[]string) {
		if c.NoExoscaleCCM {
			delete(addOns, sksClusterAddonExoscaleCCM)
		}
		if c.NoMetricsServer {
			delete(addOns, sksClusterAddonMetricsServer)
		}

		if len(addOns) > 0 {
			list := make([]string, 0)
			for k := range addOns {
				list = append(list, k)
			}
			v = &list
		}
		return
	}()

	if *cluster.Version == "latest" {
		versions, err := cs.ListSKSClusterVersions(ctx)
		if err != nil || len(versions) == 0 {
			if len(versions) == 0 {
				err = errors.New("no version returned by the API")
			}
			return fmt.Errorf("unable to retrieve SKS versions: %s", err)
		}
		cluster.Version = &versions[0]
	}

	var err error
	decorateAsyncOperation(fmt.Sprintf("Creating SKS cluster %q...", *cluster.Name), func() {
		cluster, err = cs.CreateSKSCluster(ctx, c.Zone, cluster)
	})
	if err != nil {
		return err
	}

	if c.NodepoolSize > 0 {
		nodepool := &exov2.SKSNodepool{
			Description: func() (v *string) {
				if c.NodepoolDescription != "" {
					v = &c.NodepoolDescription
				}
				return
			}(),
			DiskSize: &c.NodepoolDiskSize,
			InstancePrefix: func() (v *string) {
				if c.NodepoolInstancePrefix != "" {
					v = &c.NodepoolInstancePrefix
				}
				return
			}(),
			Labels: func() (v *map[string]string) {
				if len(c.NodepoolLabels) > 0 {
					return &c.NodepoolLabels
				}
				return
			}(),
			Name: func() *string {
				if c.NodepoolName != "" {
					return &c.NodepoolName
				}
				return &c.Name
			}(),
			Size: &c.NodepoolSize,
		}

		if l := len(c.NodepoolAntiAffinityGroups); l > 0 {
			nodepoolAntiAffinityGroupIDs := make([]string, l)
			for i, v := range c.NodepoolAntiAffinityGroups {
				antiAffinityGroup, err := cs.FindAntiAffinityGroup(ctx, c.Zone, v)
				if err != nil {
					return fmt.Errorf("error retrieving Anti-Affinity Group: %s", err)
				}
				nodepoolAntiAffinityGroupIDs[i] = *antiAffinityGroup.ID
			}
			nodepool.AntiAffinityGroupIDs = &nodepoolAntiAffinityGroupIDs
		}

		if c.NodepoolDeployTarget != "" {
			deployTarget, err := cs.FindDeployTarget(ctx, c.Zone, c.NodepoolDeployTarget)
			if err != nil {
				return fmt.Errorf("error retrieving Deploy Target: %s", err)
			}
			nodepool.DeployTargetID = deployTarget.ID
		}

		nodepoolInstanceType, err := cs.FindInstanceType(ctx, c.Zone, c.NodepoolInstanceType)
		if err != nil {
			return fmt.Errorf("error retrieving instance type: %s", err)
		}
		nodepool.InstanceTypeID = nodepoolInstanceType.ID

		if l := len(c.NodepoolSecurityGroups); l > 0 {
			nodepoolSecurityGroupIDs := make([]string, l)
			for i, v := range c.NodepoolSecurityGroups {
				securityGroup, err := cs.FindSecurityGroup(ctx, c.Zone, v)
				if err != nil {
					return fmt.Errorf("error retrieving Security Group: %s", err)
				}
				nodepoolSecurityGroupIDs[i] = *securityGroup.ID
			}
			nodepool.SecurityGroupIDs = &nodepoolSecurityGroupIDs
		}

		decorateAsyncOperation(fmt.Sprintf("Adding Nodepool %q...", *nodepool.Name), func() {
			_, err = cluster.AddNodepool(ctx, nodepool)
		})
		if err != nil {
			return err
		}
	}

	if !gQuiet {
		return output(showSKSCluster(c.Zone, *cluster.ID))
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(sksCmd, &sksCreateCmd{
		KubernetesVersion:    "latest",
		NodepoolDiskSize:     50,
		NodepoolInstanceType: defaultServiceOffering,
		ServiceLevel:         defaultSKSClusterServiceLevel,
	}))
}
