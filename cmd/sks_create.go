package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

var (
	defaultSKSClusterCNI          = "calico"
	defaultSKSClusterServiceLevel = "pro"
	sksClusterAddonExoscaleCCM    = "exoscale-cloud-controller"
	sksClusterAddonExoscaleCSI    = "exoscale-container-storage-interface"
	sksClusterAddonMetricsServer  = "metrics-server"
)

type sksCreateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#" cli-usage:"NAME"`

	AutoUpgrade                  bool              `cli-usage:"enable automatic upgrading of the SKS cluster control plane Kubernetes version"`
	CNI                          string            `cli-usage:"CNI plugin to deploy. e.g. 'calico', or 'cilium'"`
	Description                  string            `cli-usage:"SKS cluster description"`
	EnableKubeProxy              bool              `cli-flag:"enable-kube-proxy" cli-usage:"deploy the Kubernetes network proxy"`
	KubernetesVersion            string            `cli-usage:"SKS cluster control plane Kubernetes version"`
	Labels                       map[string]string `cli-flag:"label" cli-usage:"SKS cluster label (format: key=value)"`
	NoCNI                        bool              `cli-usage:"do not deploy a default Container Network Interface plugin in the cluster control plane"`
	NoExoscaleCCM                bool              `cli-usage:"do not deploy the Exoscale Cloud Controller Manager in the cluster control plane"`
	NoMetricsServer              bool              `cli-usage:"do not deploy the Kubernetes Metrics Server in the cluster control plane"`
	ExoscaleCSI                  bool              `cli-usage:"deploy the Exoscale Container Storage Interface on worker nodes"`
	NodepoolAntiAffinityGroups   []string          `cli-flag:"nodepool-anti-affinity-group" cli-usage:"default Nodepool Anti-Affinity Group NAME|ID (can be specified multiple times)"`
	NodepoolDeployTarget         string            `cli-usage:"default Nodepool Deploy Target NAME|ID"`
	NodepoolDescription          string            `cli-usage:"default Nodepool description"`
	NodepoolDiskSize             int64             `cli-usage:"default Nodepool Compute instances disk size"`
	NodepoolImageGcLowThreshold  int64             `cli-flag:"nodepool-image-gc-low-threshold" cli-usage:"default Nodepool the percent of disk usage after which image garbage collection is never run"`
	NodepoolImageGcHighThreshold int64             `cli-flag:"nodepool-image-gc-high-threshold" cli-usage:"default Nodepool the percent of disk usage after which image garbage collection is always run"`
	NodepoolImageGcMinAge        string            `cli-flag:"nodepool-image-gc-min-age" cli-usage:"default Nodepool maximum age an image can be unused before it is garbage collected"`
	NodepoolInstancePrefix       string            `cli-usage:"string to prefix default Nodepool member names with"`
	NodepoolInstanceType         string            `cli-usage:"default Nodepool Compute instances type"`
	NodepoolLabels               map[string]string `cli-flag:"nodepool-label" cli-usage:"default Nodepool label (format: key=value)"`
	NodepoolName                 string            `cli-usage:"default Nodepool name"`
	NodepoolPrivateNetworks      []string          `cli-flag:"nodepool-private-network" cli-usage:"default Nodepool Private Network NAME|ID (can be specified multiple times)"`
	NodepoolSecurityGroups       []string          `cli-flag:"nodepool-security-group" cli-usage:"default Nodepool Security Group NAME|ID (can be specified multiple times)"`
	NodepoolSize                 int64             `cli-usage:"default Nodepool size. If 0, no default Nodepool will be added to the cluster."`
	NodepoolTaints               []string          `cli-flag:"nodepool-taint" cli-usage:"Kubernetes taint to apply to default Nodepool Nodes (format: KEY=VALUE:EFFECT, can be specified multiple times)"`
	OIDCClientID                 string            `cli-flag:"oidc-client-id" cli-usage:"OpenID client ID"`
	OIDCGroupsClaim              string            `cli-flag:"oidc-groups-claim" cli-usage:"OpenID JWT claim to use as the user's group"`
	OIDCGroupsPrefix             string            `cli-flag:"oidc-groups-prefix" cli-usage:"OpenID prefix prepended to group claims"`
	OIDCIssuerURL                string            `cli-flag:"oidc-issuer-url" cli-usage:"OpenID provider URL"`
	OIDCRequiredClaim            map[string]string `cli-flag:"oidc-required-claim" cli-usage:"OpenID token required claim (format: key=value)"`
	OIDCUsernameClaim            string            `cli-flag:"oidc-username-claim" cli-usage:"OpenID JWT claim to use as the user name"`
	OIDCUsernamePrefix           string            `cli-flag:"oidc-username-prefix" cli-usage:"OpenID prefix prepended to username claims"`
	ServiceLevel                 string            `cli-usage:"SKS cluster control plane service level (starter|pro)"`
	Zone                         string            `cli-short:"z" cli-usage:"SKS cluster zone"`
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
		strings.Join(output.TemplateAnnotations(&sksShowOutput{}), ", "))
}

func (c *sksCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksCreateCmd) cmdRun(cmd *cobra.Command, _ []string) error { //nolint:gocyclo

	clusterReq := v3.CreateSKSClusterRequest{
		AutoUpgrade: &c.AutoUpgrade,
		Cni:         v3.CreateSKSClusterRequestCni(c.CNI),
		Description: utils.NonEmptyStringPtr(c.Description),
		Labels: func() v3.Labels {
			if len(c.Labels) > 0 {
				return c.Labels
			}
			return map[string]string{}
		}(),
		Name:    c.Name,
		Level:   v3.CreateSKSClusterRequestLevel(c.ServiceLevel),
		Version: c.KubernetesVersion,
		EnableKubeProxy: func() *bool {
			if cmd.Flags().Changed("enable-kube-proxy") {
				return &c.EnableKubeProxy
			}
			return nil
		}(),
		FeatureGates: []string{},
	}

	ctx := gContext

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}
	if c.NoCNI {
		clusterReq.Cni = ""
	}

	clusterReq.Addons = func() (v []string) {
		addOns := make([]string, 0)

		if !c.NoExoscaleCCM {
			addOns = append(addOns, sksClusterAddonExoscaleCCM)
		}

		if !c.NoMetricsServer {
			addOns = append(addOns, sksClusterAddonMetricsServer)
		}

		if c.ExoscaleCSI {
			addOns = append(addOns, sksClusterAddonExoscaleCSI)
		}

		if len(addOns) > 0 {
			v = addOns
		}
		return
	}()

	if clusterReq.Version == "latest" {
		versions, err := client.ListSKSClusterVersions(ctx)
		if err != nil || len(versions.SKSClusterVersions) == 0 {
			return fmt.Errorf("unable to retrieve SKS versions: %w", err)
		}
		if versions == nil || len(versions.SKSClusterVersions) == 0 {
			return errors.New("no version returned by the API")
		}

		clusterReq.Version = versions.SKSClusterVersions[0]
	}

	if c.OIDCClientID != "" {

		clusterReq.Oidc = &v3.SKSOidc{
			ClientID:       c.OIDCClientID,
			GroupsClaim:    c.OIDCGroupsClaim,
			GroupsPrefix:   c.OIDCGroupsPrefix,
			IssuerURL:      c.OIDCIssuerURL,
			UsernameClaim:  c.OIDCUsernameClaim,
			UsernamePrefix: c.OIDCUsernamePrefix,
			RequiredClaim: func() map[string]string {
				if len(c.OIDCRequiredClaim) > 0 {
					return c.OIDCRequiredClaim
				}
				return map[string]string{}
			}(),
		}
	}

	op, err := client.CreateSKSCluster(ctx, clusterReq)
	if err != nil {
		return err
	}
	decorateAsyncOperation(fmt.Sprintf("Creating SKS cluster %q...", clusterReq.Name), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	clusterId := op.Reference.ID

	if c.NodepoolSize > 0 {
		nodepoolName := c.Name
		if c.NodepoolName != "" {
			nodepoolName = c.NodepoolName
		}

		nodepoolReq, err := createNodepoolRequest(
			ctx,
			client,
			nodepoolName,
			c.NodepoolDescription,
			c.NodepoolDiskSize,
			c.NodepoolInstancePrefix,
			c.NodepoolSize,
			c.NodepoolInstanceType,
			c.NodepoolLabels,
			c.NodepoolAntiAffinityGroups,
			c.NodepoolDeployTarget,
			c.NodepoolPrivateNetworks,
			c.NodepoolSecurityGroups,
			c.NodepoolTaints,
			&v3.KubeletImageGC{
				MinAge:        c.NodepoolImageGcMinAge,
				LowThreshold:  c.NodepoolImageGcLowThreshold,
				HighThreshold: c.NodepoolImageGcHighThreshold,
			},
		)
		if err != nil {
			return err
		}

		op, err := client.CreateSKSNodepool(ctx, clusterId, nodepoolReq)
		if err != nil {
			return err
		}
		decorateAsyncOperation(fmt.Sprintf("Adding Nodepool %q...", nodepoolReq.Name), func() {
			op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		})
		if err != nil {
			return err
		}
	}

	if !globalstate.Quiet {
		return (&sksShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			Cluster:            clusterId.String(),
			Zone:               v3.ZoneName(c.Zone),
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(sksCmd, &sksCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		CNI:                          defaultSKSClusterCNI,
		KubernetesVersion:            "latest",
		NodepoolDiskSize:             50,
		NodepoolInstanceType:         fmt.Sprintf("%s.%s", defaultInstanceTypeFamily, defaultInstanceType),
		NodepoolImageGcLowThreshold:  kubeletImageGcLowThreshold,
		NodepoolImageGcHighThreshold: kubeletImageGcHighThreshold,
		NodepoolImageGcMinAge:        kubeletImageGcMinAge,
		ServiceLevel:                 defaultSKSClusterServiceLevel,
	}))

}
