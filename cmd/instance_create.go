package cmd

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	exossh "github.com/exoscale/cli/pkg/ssh"
	"github.com/exoscale/cli/pkg/userdata"
	"github.com/exoscale/cli/utils"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceCreateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#" cli-usage:"NAME"`

	AntiAffinityGroups []string          `cli-flag:"anti-affinity-group" cli-usage:"instance Anti-Affinity Group NAME|ID (can be specified multiple times)"`
	BlockStorageVolume string            `cli-flag:"block-storage-volume" cli-usage:"Block Storage Volume NAME|ID to attach on the instance"`
	CloudInitFile      string            `cli-flag:"cloud-init" cli-usage:"instance cloud-init user data configuration file path"`
	CloudInitCompress  bool              `cli-flag:"cloud-init-compress" cli-usage:"compress instance cloud-init user data"`
	DeployTarget       string            `cli-usage:"instance Deploy Target NAME|ID"`
	DiskSize           int64             `cli-usage:"instance disk size"`
	IPv6               bool              `cli-flag:"ipv6" cli-usage:"enable IPv6 on instance"`
	InstanceType       string            `cli-usage:"instance type (format: [FAMILY.]SIZE)"`
	Labels             map[string]string `cli-flag:"label" cli-usage:"instance label (format: key=value)"`
	PrivateNetworks    []string          `cli-flag:"private-network" cli-usage:"instance Private Network NAME|ID (can be specified multiple times)"`
	PrivateInstance    bool              `cli-flag:"private-instance" cli-usage:"enable private instance to be created"`
	SSHKey             string            `cli-flag:"ssh-key" cli-usage:"SSH key to deploy on the instance"`
	SecurityGroups     []string          `cli-flag:"security-group" cli-usage:"instance Security Group NAME|ID (can be specified multiple times)"`
	Template           string            `cli-usage:"instance template NAME|ID"`
	TemplateVisibility string            `cli-usage:"instance template visibility (public|private)"`
	Zone               string            `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceCreateCmd) cmdAliases() []string { return gCreateAlias }

func (c *instanceCreateCmd) cmdShort() string { return "Create a Compute instance" }

func (c *instanceCreateCmd) cmdLong() string {
	return fmt.Sprintf(`This command creates a Compute instance.

Supported Compute instance type families: %s

Supported Compute instance type sizes: %s

Supported output template annotations: %s`,
		strings.Join(instanceTypeFamilies, ", "),
		strings.Join(instanceTypeSizes, ", "),
		strings.Join(output.TemplateAnnotations(&instanceShowOutput{}), ", "))
}

func (c *instanceCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	cmdSetTemplateFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceCreateCmd) cmdRun(_ *cobra.Command, _ []string) error { //nolint:gocyclo
	var (
		singleUseSSHPrivateKey *rsa.PrivateKey
		singleUseSSHPublicKey  ssh.PublicKey
		sshKey                 *egoscale.SSHKey
	)

	instance := &egoscale.Instance{
		DiskSize:    &c.DiskSize,
		IPv6Enabled: &c.IPv6,
		Labels: func() (v *map[string]string) {
			if len(c.Labels) > 0 {
				return &c.Labels
			}
			return
		}(),
		Name:   &c.Name,
		SSHKey: utils.NonEmptyStringPtr(c.SSHKey),
	}

	if c.PrivateInstance {
		t := "none"
		instance.PublicIPAssignment = &t
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	if l := len(c.AntiAffinityGroups); l > 0 {
		antiAffinityGroupIDs := make([]string, l)
		for i := range c.AntiAffinityGroups {
			antiAffinityGroup, err := globalstate.EgoscaleClient.FindAntiAffinityGroup(ctx, c.Zone, c.AntiAffinityGroups[i])
			if err != nil {
				return fmt.Errorf("error retrieving Anti-Affinity Group: %w", err)
			}
			antiAffinityGroupIDs[i] = *antiAffinityGroup.ID
		}
		instance.AntiAffinityGroupIDs = &antiAffinityGroupIDs
	}

	if c.DeployTarget != "" {
		deployTarget, err := globalstate.EgoscaleClient.FindDeployTarget(ctx, c.Zone, c.DeployTarget)
		if err != nil {
			return fmt.Errorf("error retrieving Deploy Target: %w", err)
		}
		instance.DeployTargetID = deployTarget.ID
	}

	instanceType, err := globalstate.EgoscaleClient.FindInstanceType(ctx, c.Zone, c.InstanceType)
	if err != nil {
		return fmt.Errorf("error retrieving instance type: %w", err)
	}
	instance.InstanceTypeID = instanceType.ID

	privateNetworks := make([]*egoscale.PrivateNetwork, len(c.PrivateNetworks))
	if l := len(c.PrivateNetworks); l > 0 {
		for i := range c.PrivateNetworks {
			privateNetwork, err := globalstate.EgoscaleClient.FindPrivateNetwork(ctx, c.Zone, c.PrivateNetworks[i])
			if err != nil {
				return fmt.Errorf("error retrieving Private Network: %w", err)
			}
			privateNetworks[i] = privateNetwork
		}
	}

	if l := len(c.SecurityGroups); l > 0 {
		securityGroupIDs := make([]string, l)
		for i := range c.SecurityGroups {
			securityGroup, err := globalstate.EgoscaleClient.FindSecurityGroup(ctx, c.Zone, c.SecurityGroups[i])
			if err != nil {
				return fmt.Errorf("error retrieving Security Group: %w", err)
			}
			securityGroupIDs[i] = *securityGroup.ID
		}
		instance.SecurityGroupIDs = &securityGroupIDs
	}

	if instance.SSHKey == nil && account.CurrentAccount.DefaultSSHKey != "" {
		instance.SSHKey = &account.CurrentAccount.DefaultSSHKey
	}

	// Generating a single-use SSH key pair for this instance.
	if instance.SSHKey == nil {
		singleUseSSHPrivateKey, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return fmt.Errorf("error generating SSH private key: %w", err)
		}
		if err = singleUseSSHPrivateKey.Validate(); err != nil {
			return fmt.Errorf("error generating SSH private key: %w", err)
		}

		singleUseSSHPublicKey, err = ssh.NewPublicKey(&singleUseSSHPrivateKey.PublicKey)
		if err != nil {
			return fmt.Errorf("error generating SSH public key: %w", err)
		}

		sshKey, err = globalstate.EgoscaleClient.RegisterSSHKey(
			ctx,
			c.Zone,
			fmt.Sprintf("%s-%d", c.Name, time.Now().Unix()),
			string(ssh.MarshalAuthorizedKey(singleUseSSHPublicKey)),
		)
		if err != nil {
			return fmt.Errorf("error registering SSH key: %w", err)
		}

		instance.SSHKey = sshKey.Name
	}

	template, err := globalstate.EgoscaleClient.FindTemplate(ctx, c.Zone, c.Template, c.TemplateVisibility)
	if err != nil {
		return fmt.Errorf(
			"no template %q found with visibility %s in zone %s",
			c.Template,
			c.TemplateVisibility,
			c.Zone,
		)
	}
	instance.TemplateID = template.ID

	if c.CloudInitFile != "" {
		userData, err := userdata.GetUserDataFromFile(c.CloudInitFile, c.CloudInitCompress)
		if err != nil {
			return fmt.Errorf("error parsing cloud-init user data: %w", err)
		}
		instance.UserData = &userData
	}

	clientv3, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Creating instance %q...", c.Name), func() {
		instance, err = globalstate.EgoscaleClient.CreateInstance(ctx, c.Zone, instance)
		if err != nil {
			return
		}

		for _, p := range privateNetworks {
			if err = globalstate.EgoscaleClient.AttachInstanceToPrivateNetwork(ctx, c.Zone, instance, p); err != nil {
				return
			}
		}

		if c.BlockStorageVolume != "" {
			var volumes *v3.ListBlockStorageVolumesResponse
			volumes, err = clientv3.ListBlockStorageVolumes(ctx)
			if err != nil {
				return
			}

			var volume v3.BlockStorageVolume
			volume, err = volumes.FindBlockStorageVolume(c.BlockStorageVolume)
			if err != nil {
				return
			}

			var op *v3.Operation
			op, err = clientv3.AttachBlockStorageVolumeToInstance(ctx, volume.ID, v3.AttachBlockStorageVolumeToInstanceRequest{
				Instance: &v3.InstanceTarget{
					ID: v3.UUID(*instance.ID),
				},
			})
			if err != nil {
				return
			}
			_, err = clientv3.Wait(ctx, op, v3.OperationStateSuccess)
		}
	})
	if err != nil {
		return err
	}

	if singleUseSSHPrivateKey != nil {
		privateKeyFilePath := exossh.GetInstanceSSHKeyPath(*instance.ID)

		if err = os.MkdirAll(path.Dir(privateKeyFilePath), 0o700); err != nil {
			return fmt.Errorf("error writing SSH private key file: %w", err)
		}

		if err = os.WriteFile(
			privateKeyFilePath,
			pem.EncodeToMemory(&pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(singleUseSSHPrivateKey),
			}),
			0o600,
		); err != nil {
			return fmt.Errorf("error writing SSH private key file: %w", err)
		}

		if err = globalstate.EgoscaleClient.DeleteSSHKey(ctx, c.Zone, sshKey); err != nil {
			return fmt.Errorf("error deleting SSH key: %w", err)
		}
	}

	if !globalstate.Quiet {
		return (&instanceShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			Instance:           *instance.ID,
			Zone:               c.Zone,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceCmd, &instanceCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		DiskSize:           50,
		InstanceType:       fmt.Sprintf("%s.%s", defaultInstanceTypeFamily, defaultInstanceType),
		TemplateVisibility: defaultTemplateVisibility,
	}))
}
