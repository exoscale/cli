package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/exoscale/cli/utils"
	egoscale "github.com/exoscale/egoscale/v2"
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
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#" cli-usage:"NAME"`

	AutoUpgrade                bool              `cli-usage:"enable automatic upgrading of the SKS cluster control plane Kubernetes version"`
	CNI                        string            `cli-usage:"CNI plugin to deploy. e.g. 'calico', or 'cilium'"`
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
	NodepoolPrivateNetworks    []string          `cli-flag:"nodepool-private-network" cli-usage:"default Nodepool Private Network NAME|ID (can be specified multiple times)"`
	NodepoolSecurityGroups     []string          `cli-flag:"nodepool-security-group" cli-usage:"default Nodepool Security Group NAME|ID (can be specified multiple times)"`
	NodepoolSize               int64             `cli-usage:"default Nodepool size. If 0, no default Nodepool will be added to the cluster."`
	NodepoolTaints             []string          `cli-flag:"nodepool-taint" cli-usage:"Kubernetes taint to apply to default Nodepool Nodes (format: KEY=VALUE:EFFECT, can be specified multiple times)"`
	OIDCClientID               string            `cli-flag:"oidc-client-id" cli-usage:"OpenID client ID"`
	OIDCGroupsClaim            string            `cli-flag:"oidc-groups-claim" cli-usage:"OpenID JWT claim to use as the user's group"`
	OIDCGroupsPrefix           string            `cli-flag:"oidc-groups-prefix" cli-usage:"OpenID prefix prepended to group claims"`
	OIDCIssuerURL              string            `cli-flag:"oidc-issuer-url" cli-usage:"OpenID provider URL"`
	OIDCRequiredClaim          map[string]string `cli-flag:"oidc-required-claim" cli-usage:"OpenID token required claim (format: key=value)"`
	OIDCUsernameClaim          string            `cli-flag:"oidc-username-claim" cli-usage:"OpenID JWT claim to use as the user name"`
	OIDCUsernamePrefix         string            `cli-flag:"oidc-username-prefix" cli-usage:"OpenID prefix prepended to username claims"`
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
	cluster := &egoscale.SKSCluster{
		AutoUpgrade: &c.AutoUpgrade,
		CNI:         &c.CNI,
		Description: utils.NonEmptyStringPtr(c.Description),
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
			return fmt.Errorf("unable to retrieve SKS versions: %w", err)
		}
		cluster.Version = &versions[0]
	}

	var opts []egoscale.CreateSKSClusterOpt
	if c.OIDCClientID != "" {
		opts = append(opts, egoscale.CreateSKSClusterWithOIDC(&egoscale.SKSClusterOIDCConfig{
			ClientID:     &c.OIDCClientID,
			GroupsClaim:  utils.NonEmptyStringPtr(c.OIDCGroupsClaim),
			GroupsPrefix: utils.NonEmptyStringPtr(c.OIDCGroupsPrefix),
			IssuerURL:    &c.OIDCIssuerURL,
			RequiredClaim: func() (v *map[string]string) {
				if len(c.OIDCRequiredClaim) > 0 {
					v = &c.OIDCRequiredClaim
				}
				return
			}(),
			UsernameClaim:  utils.NonEmptyStringPtr(c.OIDCUsernameClaim),
			UsernamePrefix: utils.NonEmptyStringPtr(c.OIDCUsernamePrefix),
		}))
	}

	var err error
	decorateAsyncOperation(fmt.Sprintf("Creating SKS cluster %q...", *cluster.Name), func() {
		cluster, err = cs.CreateSKSCluster(ctx, c.Zone, cluster, opts...)
	})
	if err != nil {
		return err
	}

	if c.NodepoolSize > 0 {
		nodepool := &egoscale.SKSNodepool{
			Description:    utils.NonEmptyStringPtr(c.NodepoolDescription),
			DiskSize:       &c.NodepoolDiskSize,
			InstancePrefix: utils.NonEmptyStringPtr(c.NodepoolInstancePrefix),
			Labels: func() (v *map[string]string) {
				if len(c.NodepoolLabels) > 0 {
					return &c.NodepoolLabels
				}
				return
			}(),
			Name: utils.NonEmptyStringPtr(func() string {
				if c.NodepoolName != "" {
					return c.NodepoolName
				}
				return c.Name
			}()),
			Size: &c.NodepoolSize,
		}

		if l := len(c.NodepoolAntiAffinityGroups); l > 0 {
			nodepoolAntiAffinityGroupIDs := make([]string, l)
			for i, v := range c.NodepoolAntiAffinityGroups {
				antiAffinityGroup, err := cs.FindAntiAffinityGroup(ctx, c.Zone, v)
				if err != nil {
					return fmt.Errorf("error retrieving Anti-Affinity Group: %w", err)
				}
				nodepoolAntiAffinityGroupIDs[i] = *antiAffinityGroup.ID
			}
			nodepool.AntiAffinityGroupIDs = &nodepoolAntiAffinityGroupIDs
		}

		if c.NodepoolDeployTarget != "" {
			deployTarget, err := cs.FindDeployTarget(ctx, c.Zone, c.NodepoolDeployTarget)
			if err != nil {
				return fmt.Errorf("error retrieving Deploy Target: %w", err)
			}
			nodepool.DeployTargetID = deployTarget.ID
		}

		nodepoolInstanceType, err := cs.FindInstanceType(ctx, c.Zone, c.NodepoolInstanceType)
		if err != nil {
			return fmt.Errorf("error retrieving instance type: %w", err)
		}
		nodepool.InstanceTypeID = nodepoolInstanceType.ID

		if l := len(c.NodepoolPrivateNetworks); l > 0 {
			nodepoolPrivateNetworkIDs := make([]string, l)
			for i, v := range c.NodepoolPrivateNetworks {
				privateNetwork, err := cs.FindPrivateNetwork(ctx, c.Zone, v)
				if err != nil {
					return fmt.Errorf("error retrieving Private Network: %w", err)
				}
				nodepoolPrivateNetworkIDs[i] = *privateNetwork.ID
			}
			nodepool.PrivateNetworkIDs = &nodepoolPrivateNetworkIDs
		}

		if l := len(c.NodepoolSecurityGroups); l > 0 {
			nodepoolSecurityGroupIDs := make([]string, l)
			for i, v := range c.NodepoolSecurityGroups {
				securityGroup, err := cs.FindSecurityGroup(ctx, c.Zone, v)
				if err != nil {
					return fmt.Errorf("error retrieving Security Group: %w", err)
				}
				nodepoolSecurityGroupIDs[i] = *securityGroup.ID
			}
			nodepool.SecurityGroupIDs = &nodepoolSecurityGroupIDs
		}

		if len(c.NodepoolTaints) > 0 {
			taints := make(map[string]*egoscale.SKSNodepoolTaint)
			for _, t := range c.NodepoolTaints {
				key, taint, err := parseSKSNodepoolTaint(t)
				if err != nil {
					return fmt.Errorf("invalid taint value %q: %w", t, err)
				}
				taints[key] = taint
			}
			nodepool.Taints = &taints
		}

		decorateAsyncOperation(fmt.Sprintf("Adding Nodepool %q...", *nodepool.Name), func() {
			_, err = cs.CreateSKSNodepool(ctx, c.Zone, cluster, nodepool)
		})
		if err != nil {
			return err
		}
	}

	if !gQuiet {
		return (&sksShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			Cluster:            *cluster.ID,
			Zone:               c.Zone,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(sksCmd, &sksCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		CNI:                  defaultSKSClusterCNI,
		KubernetesVersion:    "latest",
		NodepoolDiskSize:     50,
		NodepoolInstanceType: defaultServiceOffering,
		ServiceLevel:         defaultSKSClusterServiceLevel,
	}))

	// FIXME: remove this someday.
	cobra.CheckErr(registerCLICommand(deprecatedSKSCmd, &sksCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		CNI:                  defaultSKSClusterCNI,
		KubernetesVersion:    "latest",
		NodepoolDiskSize:     50,
		NodepoolInstanceType: defaultServiceOffering,
		ServiceLevel:         defaultSKSClusterServiceLevel,
	}))
}
