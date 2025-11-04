package sks

import (
	"errors"
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

type sksUpdateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	Cluster string `cli-arg:"#" cli-usage:"NAME|ID"`

	AutoUpgrade         bool              `cli-usage:"enable automatic upgrading of the SKS cluster control plane Kubernetes version(--auto-upgrade=false to disable again)"`
	Description         string            `cli-usage:"SKS cluster description"`
	FeatureGates        []string          `cli-flag:"feature-gates" cli-usage:"SKS cluster feature gates to enable"`
	Labels              map[string]string `cli-flag:"label" cli-usage:"SKS cluster label (format: key=value)"`
	Name                string            `cli-usage:"SKS cluster name"`
	EnableCSIAddon      bool              `cli-usage:"enable the Exoscale CSI driver"`
	AuditEnabled        bool              `cli-flag:"audit-enabled" cli-usage:"enable or disable Kubernetes Audit logging"`
	AuditEndpoint       string            `cli-flag:"audit-endpoint" cli-usage:"Kubernetes Audit endpoint URL"`
	AuditBearerToken    string            `cli-flag:"audit-bearer-token" cli-usage:"Bearer token for Kubernetes Audit endpoint authentication"`
	AuditInitialBackoff string            `cli-flag:"audit-initial-backoff" cli-usage:"Initial backoff for Kubernetes Audit endpoint retry (default: 10s)"`
	OIDCClientID        string            `cli-flag:"oidc-client-id" cli-usage:"OpenID client ID"`
	OIDCGroupsClaim     string            `cli-flag:"oidc-groups-claim" cli-usage:"OpenID JWT claim to use as the user's group"`
	OIDCGroupsPrefix    string            `cli-flag:"oidc-groups-prefix" cli-usage:"OpenID prefix prepended to group claims"`
	OIDCIssuerURL       string            `cli-flag:"oidc-issuer-url" cli-usage:"OpenID provider URL"`
	OIDCRequiredClaim   map[string]string `cli-flag:"oidc-required-claim" cli-usage:"OpenID token required claim (format: key=value)"`
	OIDCUsernameClaim   string            `cli-flag:"oidc-username-claim" cli-usage:"OpenID JWT claim to use as the user name"`
	OIDCUsernamePrefix  string            `cli-flag:"oidc-username-prefix" cli-usage:"OpenID prefix prepended to username claims"`
	Zone                v3.ZoneName       `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksUpdateCmd) CmdAliases() []string { return nil }

func (c *sksUpdateCmd) CmdShort() string { return "Update an SKS cluster" }

func (c *sksUpdateCmd) CmdLong() string {
	return fmt.Sprintf(`This command updates an SKS cluster.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&sksShowOutput{}), ", "),
	)
}

func (c *sksUpdateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksUpdateCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	clusters, err := client.ListSKSClusters(ctx)
	if err != nil {
		return err
	}

	cluster, err := clusters.FindSKSCluster(c.Cluster)
	if err != nil {
		return err
	}

	updateReq := v3.UpdateSKSClusterRequest{}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.AutoUpgrade)) {
		updateReq.AutoUpgrade = &c.AutoUpgrade
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.FeatureGates)) {
		updateReq.FeatureGates = c.FeatureGates
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Labels)) {
		updateReq.Labels = c.Labels
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Name)) {
		updateReq.Name = c.Name
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Description)) {
		updateReq.Description = &c.Description
		updated = true
	}

	// Always ensure we have CSI addon enabled when updating them
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.EnableCSIAddon)) && (cluster.Addons == nil || !slices.Contains(cluster.Addons, sksClusterAddonExoscaleCSI)) {
		if cluster.Addons == nil {
			updateReq.Addons = []string{sksClusterAddonExoscaleCSI}
		} else {
			updateReq.Addons = append(cluster.Addons, sksClusterAddonExoscaleCSI) //nolint:gocritic
		}

		updated = true
	}

	// Configure Kubernetes Audit update if any audit flag is changed
	auditChanged := cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.AuditEnabled)) ||
		cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.AuditEndpoint)) ||
		cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.AuditBearerToken)) ||
		cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.AuditInitialBackoff))

	if auditChanged {
		if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.AuditEnabled)) &&
			cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.AuditEndpoint)) && c.AuditEnabled && c.AuditEndpoint == "" {
			return errors.New("audit endpoint is required when enabling audit")
		}

		updateReq.Audit = &v3.SKSAuditUpdate{}

		if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.AuditEnabled)) {
			updateReq.Audit.Enabled = &c.AuditEnabled
		}
		if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.AuditEndpoint)) {
			updateReq.Audit.Endpoint = v3.SKSAuditEndpoint(c.AuditEndpoint)
		}
		if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.AuditBearerToken)) {
			updateReq.Audit.BearerToken = v3.SKSAuditBearerToken(c.AuditBearerToken)
		}
		if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.AuditInitialBackoff)) {
			updateReq.Audit.InitialBackoff = v3.SKSAuditInitialBackoff(c.AuditInitialBackoff)
		}

		updated = true
	}

	// Configure OIDC update if any audit flag is changed
	OIDCChanged := cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.OIDCClientID)) ||
		cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.OIDCGroupsClaim)) ||
		cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.OIDCGroupsPrefix)) ||
		cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.OIDCIssuerURL)) ||
		cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.OIDCRequiredClaim)) ||
		cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.OIDCUsernameClaim)) ||
		cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.OIDCUsernamePrefix))

	if OIDCChanged {
		updateReq.Oidc = &v3.SKSOidc{}

		if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.OIDCClientID)) {
			updateReq.Oidc.ClientID = c.OIDCClientID
		}

		if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.OIDCGroupsClaim)) {
			updateReq.Oidc.GroupsClaim = c.OIDCGroupsClaim
		}

		if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.OIDCGroupsPrefix)) {
			updateReq.Oidc.GroupsPrefix = c.OIDCGroupsPrefix
		}

		if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.OIDCIssuerURL)) {
			updateReq.Oidc.IssuerURL = c.OIDCIssuerURL
		}

		if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.OIDCRequiredClaim)) {
			updateReq.Oidc.RequiredClaim = c.OIDCRequiredClaim
		}

		if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.OIDCUsernameClaim)) {
			updateReq.Oidc.UsernameClaim = c.OIDCUsernameClaim
		}

		if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.OIDCUsernamePrefix)) {
			updateReq.Oidc.UsernamePrefix = c.OIDCUsernamePrefix
		}

		updated = true
	}

	if updated {
		op, err := client.UpdateSKSCluster(ctx, cluster.ID, updateReq)
		if err != nil {
			return err
		}

		utils.DecorateAsyncOperation(fmt.Sprintf("Updating SKS cluster %q...", c.Cluster), func() {
			_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		})

		if err != nil {
			return err
		}
	}

	if !globalstate.Quiet {
		return (&sksShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Cluster:            string(cluster.ID),
			Zone:               c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(sksCmd, &sksUpdateCmd{
		CliCommandSettings:  exocmd.DefaultCLICmdSettings(),
		AuditInitialBackoff: "10s",
	}))
}
