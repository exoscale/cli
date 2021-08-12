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

	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

type instanceCreateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#" cli-usage:"NAME"`

	AntiAffinityGroups []string          `cli-flag:"anti-affinity-group" cli-usage:"instance Anti-Affinity Group NAME|ID (can be specified multiple times)"`
	CloudInitFile      string            `cli-flag:"cloud-init" cli-usage:"instance cloud-init user data configuration file path"`
	DeployTarget       string            `cli-usage:"instance Deploy Target NAME|ID"`
	DiskSize           int64             `cli-usage:"instance disk size"`
	IPv6               bool              `cli-flag:"ipv6" cli-usage:"enable IPv6 on instance"`
	InstanceType       string            `cli-usage:"instance type (format: [FAMILY.]SIZE)"`
	Labels             map[string]string `cli-flag:"label" cli-usage:"instance label (format: key=value)"`
	PrivateNetworks    []string          `cli-flag:"private-network" cli-usage:"instance Private Network NAME|ID (can be specified multiple times)"`
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
		strings.Join(outputterTemplateAnnotations(&instanceShowOutput{}), ", "))
}

func (c *instanceCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceCreateCmd) cmdRun(_ *cobra.Command, _ []string) error {
	var (
		singleUseSSHPrivateKey *rsa.PrivateKey
		singleUseSSHPublicKey  ssh.PublicKey
		sshKey                 *egoscale.SSHKey
	)

	instance := &egoscale.Instance{
		DeployTargetID: func() (v *string) {
			if c.DeployTarget != "" {
				v = &c.DeployTarget
			}
			return
		}(),
		DiskSize:    &c.DiskSize,
		IPv6Enabled: &c.IPv6,
		Labels: func() (v *map[string]string) {
			if len(c.Labels) > 0 {
				return &c.Labels
			}
			return
		}(),
		Name: &c.Name,
		SSHKey: func() (v *string) {
			if c.SSHKey != "" {
				v = &c.SSHKey
			}
			return
		}(),
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	if l := len(c.AntiAffinityGroups); l > 0 {
		antiAffinityGroupIDs := make([]string, l)
		for i := range c.AntiAffinityGroups {
			antiAffinityGroup, err := cs.FindAntiAffinityGroup(ctx, c.Zone, c.AntiAffinityGroups[i])
			if err != nil {
				return fmt.Errorf("error retrieving Anti-Affinity Group: %s", err)
			}
			antiAffinityGroupIDs[i] = *antiAffinityGroup.ID
		}
		instance.AntiAffinityGroupIDs = &antiAffinityGroupIDs
	}

	if c.DeployTarget != "" {
		deployTarget, err := cs.FindDeployTarget(ctx, c.Zone, c.DeployTarget)
		if err != nil {
			return fmt.Errorf("error retrieving Deploy Target: %s", err)
		}
		instance.DeployTargetID = deployTarget.ID
	}

	instanceType, err := cs.FindInstanceType(ctx, c.Zone, c.InstanceType)
	if err != nil {
		return fmt.Errorf("error retrieving instance type: %s", err)
	}
	instance.InstanceTypeID = instanceType.ID

	privateNetworks := make([]*egoscale.PrivateNetwork, len(c.PrivateNetworks))
	if l := len(c.PrivateNetworks); l > 0 {
		for i := range c.PrivateNetworks {
			privateNetwork, err := cs.FindPrivateNetwork(ctx, c.Zone, c.PrivateNetworks[i])
			if err != nil {
				return fmt.Errorf("error retrieving Private Network: %s", err)
			}
			privateNetworks[i] = privateNetwork
		}
	}

	if l := len(c.SecurityGroups); l > 0 {
		securityGroupIDs := make([]string, l)
		for i := range c.SecurityGroups {
			securityGroup, err := cs.FindSecurityGroup(ctx, c.Zone, c.SecurityGroups[i])
			if err != nil {
				return fmt.Errorf("error retrieving Security Group: %s", err)
			}
			securityGroupIDs[i] = *securityGroup.ID
		}
		instance.SecurityGroupIDs = &securityGroupIDs
	}

	if instance.SSHKey == nil && gCurrentAccount.DefaultSSHKey != "" {
		instance.SSHKey = &gCurrentAccount.DefaultSSHKey
	}

	// Generating a single-use SSH key pair for this instance.
	if instance.SSHKey == nil {
		singleUseSSHPrivateKey, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return fmt.Errorf("error generating SSH private key: %s", err)
		}
		if err = singleUseSSHPrivateKey.Validate(); err != nil {
			return fmt.Errorf("error generating SSH private key: %s", err)
		}

		singleUseSSHPublicKey, err = ssh.NewPublicKey(&singleUseSSHPrivateKey.PublicKey)
		if err != nil {
			return fmt.Errorf("error generating SSH public key: %s", err)
		}

		sshKey, err = cs.RegisterSSHKey(
			ctx,
			c.Zone,
			fmt.Sprintf("%s-%d", c.Name, time.Now().Unix()),
			string(ssh.MarshalAuthorizedKey(singleUseSSHPublicKey)),
		)
		if err != nil {
			return fmt.Errorf("error registering SSH key: %s", err)
		}

		instance.SSHKey = sshKey.Name
	}

	templates, err := cs.ListTemplates(ctx, c.Zone, c.TemplateVisibility, "")
	if err != nil {
		return fmt.Errorf("error retrieving templates: %s", err)
	}
	for _, template := range templates {
		if *template.ID == c.Template || *template.Name == c.Template {
			instance.TemplateID = template.ID
			break
		}
	}
	if instance.TemplateID == nil {
		return fmt.Errorf("no template %q found with visibility %s", c.Template, c.TemplateVisibility)
	}

	if c.CloudInitFile != "" {
		userData, err := getUserDataFromFile(c.CloudInitFile)
		if err != nil {
			return fmt.Errorf("error parsing cloud-init user data: %s", err)
		}
		instance.UserData = &userData
	}

	decorateAsyncOperation(fmt.Sprintf("Creating instance %q...", c.Name), func() {
		instance, err = cs.CreateInstance(ctx, c.Zone, instance)
		if err != nil {
			return
		}

		for _, p := range privateNetworks {
			if err = cs.AttachInstanceToPrivateNetwork(ctx, c.Zone, instance, p, nil); err != nil {
				return
			}
		}
	})
	if err != nil {
		return err
	}

	if singleUseSSHPrivateKey != nil {
		privateKeyFilePath := getInstanceSSHKeyPath(*instance.ID)

		if err = os.MkdirAll(path.Dir(privateKeyFilePath), 0o700); err != nil {
			return fmt.Errorf("error writing SSH private key file: %s", err)
		}

		if err = os.WriteFile(
			privateKeyFilePath,
			pem.EncodeToMemory(&pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(singleUseSSHPrivateKey),
			}),
			0o600,
		); err != nil {
			return fmt.Errorf("error writing SSH private key file: %s", err)
		}

		if err = cs.DeleteSSHKey(ctx, c.Zone, sshKey); err != nil {
			return fmt.Errorf("error deleting SSH key: %s", err)
		}
	}

	if !gQuiet {
		return output(showInstance(c.Zone, *instance.ID))
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(computeInstanceCmd, &instanceCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		DiskSize:           50,
		InstanceType:       fmt.Sprintf("%s.%s", defaultInstanceTypeFamily, defaultInstanceType),
		Template:           defaultTemplate,
		TemplateVisibility: defaultTemplateVisibility,
	}))
}
