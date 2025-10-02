package sks

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

var (
	defaultSKSClusterCNI          = "calico"
	defaultSKSClusterServiceLevel = "pro"
	defaultSKSAuditInitialBackoff = "10s"
	sksClusterAddonExoscaleCCM    = "exoscale-cloud-controller"
	sksClusterAddonExoscaleCSI    = "exoscale-container-storage-interface"
	sksClusterAddonMetricsServer  = "metrics-server"
)

type sksCreateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

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
	FeatureGates                 []string          `cli-flag:"feature-gates" cli-usage:"SKS cluster feature gates to enable"`
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
	AuditEndpoint                string            `cli-flag:"audit-endpoint" cli-usage:"Kubernetes Audit endpoint URL (enables audit logging when set)"`
	AuditBearerToken             string            `cli-flag:"audit-bearer-token" cli-usage:"Bearer token for Kubernetes Audit endpoint authentication"`
	AuditInitialBackoff          string            `cli-flag:"audit-initial-backoff" cli-usage:"Initial backoff for Kubernetes Audit endpoint retry (default: 10s)"`
	ServiceLevel                 string            `cli-usage:"SKS cluster control plane service level (starter|pro)"`
	Zone                         string            `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksCreateCmd) CmdAliases() []string { return exocmd.GCreateAlias }

func (c *sksCreateCmd) CmdShort() string { return "Create an SKS cluster" }

func (c *sksCreateCmd) CmdLong() string {
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

func (c *sksCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksCreateCmd) CmdRun(cmd *cobra.Command, _ []string) error { //nolint:gocyclo

	clusterReq := v3.CreateSKSClusterRequest{
		AutoUpgrade: &c.AutoUpgrade,
		Cni:         v3.CreateSKSClusterRequestCni(c.CNI),
		Description: utils.NonEmptyStringPtr(c.Description),
		Labels: func() v3.SKSClusterLabels {
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
		FeatureGates: c.FeatureGates,
	}

	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}
	if c.NoCNI {
		clusterReq.Cni = ""
	}

	clusterReq.Addons = func() (v *v3.SKSClusterAddons) {
		addOns := make(v3.SKSClusterAddons, 0)

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
			v = &addOns
		}
		return
	}()

	// Configure Kubernetes Audit if endpoint is provided
	if c.AuditEndpoint != "" {
		if c.AuditBearerToken == "" {
			return errors.New("audit bearer token is required when audit endpoint is specified")
		}
		clusterReq.Audit = &v3.SKSAuditCreate{
			Endpoint:       v3.SKSAuditEndpoint(c.AuditEndpoint),
			BearerToken:    v3.SKSAuditBearerToken(c.AuditBearerToken),
			InitialBackoff: v3.SKSAuditInitialBackoff(c.AuditInitialBackoff),
		}
	}

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
	utils.DecorateAsyncOperation(fmt.Sprintf("Creating SKS cluster %q...", clusterReq.Name), func() {
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
			CreateNodepoolOpts{
				Name:               nodepoolName,
				Description:        c.NodepoolDescription,
				DiskSize:           c.NodepoolDiskSize,
				InstancePrefix:     c.NodepoolInstancePrefix,
				Size:               c.NodepoolSize,
				InstanceType:       c.NodepoolInstanceType,
				Labels:             c.NodepoolLabels,
				AntiAffinityGroups: c.NodepoolAntiAffinityGroups,
				DeployTarget:       c.NodepoolDeployTarget,
				PrivateNetworks:    c.NodepoolPrivateNetworks,
				SecurityGroups:     c.NodepoolSecurityGroups,
				Taints:             c.NodepoolTaints,
				KubeletImageGC: &v3.KubeletImageGC{
					MinAge:        c.NodepoolImageGcMinAge,
					LowThreshold:  c.NodepoolImageGcLowThreshold,
					HighThreshold: c.NodepoolImageGcHighThreshold,
				},
			},
		)
		if err != nil {
			return err
		}

		op, err := client.CreateSKSNodepool(ctx, clusterId, nodepoolReq)
		if err != nil {
			return err
		}
		utils.DecorateAsyncOperation(fmt.Sprintf("Adding Nodepool %q...", nodepoolReq.Name), func() {
			op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		})
		if err != nil {
			return err
		}
	}

	if !globalstate.Quiet {
		return (&sksShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Cluster:            clusterId.String(),
			Zone:               v3.ZoneName(c.Zone),
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(sksCmd, &sksCreateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),

		CNI:                          defaultSKSClusterCNI,
		KubernetesVersion:            "latest",
		NodepoolDiskSize:             50,
		NodepoolInstanceType:         fmt.Sprintf("%s.%s", exocmd.DefaultInstanceTypeFamily, exocmd.DefaultInstanceType),
		NodepoolImageGcLowThreshold:  kubeletImageGcLowThreshold,
		NodepoolImageGcHighThreshold: kubeletImageGcHighThreshold,
		NodepoolImageGcMinAge:        kubeletImageGcMinAge,
		AuditInitialBackoff:          defaultSKSAuditInitialBackoff,
		ServiceLevel:                 defaultSKSClusterServiceLevel,
	}))

}
