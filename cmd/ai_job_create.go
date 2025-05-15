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

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	exossh "github.com/exoscale/cli/pkg/ssh"
	"github.com/exoscale/cli/pkg/userdata"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

const (
	// AI job default template name
	aiJobDefaultTemplateID = "linux-debian-12-gpu"
)

var (
	// Creates a PVC to store the fine-tuning check points
	aiJobPVCTemplate = `
#cloud-config
write_files:
 - path: /var/lib/rancher/k3s/server/manifests/ai-job.yaml
   content: |
     apiVersion: ai.re-cinq.com/v1
     kind: Job
     metadata:
       name: ai-job
     spec:
       image: %s
       model: %s
       diskSize: %d
       huggingFaceSecret: %s
   owner: 'root:root'
   permissions: '0640'
`
)

type aiJobCreateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#" cli-usage:"NAME"`

	AntiAffinityGroups []string          `cli-flag:"anti-affinity-group" cli-usage:"instance Anti-Affinity Group NAME|ID (can be specified multiple times)"`
	DeployTarget       string            `cli-usage:"instance Deploy Target NAME|ID"`
	DiskSize           int64             `cli-usage:"instance disk size"`
	IPv6               bool              `cli-flag:"ipv6" cli-usage:"enable IPv6 on instance"`
	InstanceType       string            `cli-usage:"instance type (format: [FAMILY.]SIZE)"`
	Labels             map[string]string `cli-flag:"label" cli-usage:"instance label (format: key=value)"`
	PrivateNetworks    []string          `cli-flag:"private-network" cli-usage:"instance Private Network NAME|ID (can be specified multiple times)"`
	PrivateInstance    bool              `cli-flag:"private-instance" cli-usage:"enable private instance to be created"`
	SSHKeys            []string          `cli-flag:"ssh-key" cli-usage:"SSH key to deploy on the instance (can be specified multiple times)"`
	Protection         bool              `cli-flag:"protection" cli-usage:"enable delete protection"`
	SecurityGroups     []string          `cli-flag:"security-group" cli-usage:"instance Security Group NAME|ID (can be specified multiple times)"`
	Template           string            `cli-usage:"instance template NAME|ID"`
	TemplateVisibility string            `cli-usage:"instance template visibility (public|private)"`
	Zone               v3.ZoneName       `cli-short:"z" cli-usage:"instance zone"`

	// AI job options
	ContainerImage    string `cli-short:"i" cli-flag:"container-image" cli-usage:"container image to use for the AI job"`
	HuggingFaceSecret string `cli-short:"s" cli-flag:"hf-secret" cli-usage:"HuggingFace secret to use for the AI job"`
	Model             string `cli-flag:"m" cli-usage:"model to use for the AI job"`
	JobName           string `cli-short:"j" cli-flag:"job-name" cli-usage:"name of the AI job to create"`
}

func (c *aiJobCreateCmd) cmdAliases() []string { return gCreateAlias }

func (c *aiJobCreateCmd) cmdShort() string { return "Create an AI job" }

func (c *aiJobCreateCmd) cmdLong() string {
	return fmt.Sprintf(`This command creates an AI Job.

Supported Compute instance type families: %s

Supported Compute instance type sizes: %s

Supported output template annotations: %s`,
		strings.Join(gpuInstanceTypeFamilies, ", "),
		strings.Join(gpuInstanceTypeSizes, ", "),
		strings.Join(output.TemplateAnnotations(&instanceShowOutput{}), ", "))
}

func (c *aiJobCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	cmdSetTemplateFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *aiJobCreateCmd) cmdRun(_ *cobra.Command, _ []string) error { //nolint:gocyclo
	var (
		singleUseSSHPrivateKey *rsa.PrivateKey
		singleUseSSHPublicKey  ssh.PublicKey
	)
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	var sshKeys []v3.SSHKey
	for _, sshkeyName := range c.SSHKeys {
		sshKeys = append(sshKeys, v3.SSHKey{Name: sshkeyName})
	}

	// Set the default AI JOB template if not specified
	visibility := "private"
	if c.TemplateVisibility != "" {
		visibility = c.TemplateVisibility
	}

	// Set the default template
	templateID := aiJobDefaultTemplateID

	instanceReq := v3.CreateInstanceRequest{
		DiskSize:    c.DiskSize,
		Ipv6Enabled: &c.IPv6,
		Labels:      c.Labels,
		Name:        c.Name,
		SSHKeys:     sshKeys,
	}

	if c.PrivateInstance {
		instanceReq.PublicIPAssignment = v3.PublicIPAssignmentNone
	}

	if l := len(c.AntiAffinityGroups); l > 0 {
		instanceReq.AntiAffinityGroups = make([]v3.AntiAffinityGroup, l)
		af, err := client.ListAntiAffinityGroups(ctx)
		if err != nil {
			return fmt.Errorf("error listing Anti-Affinity Group: %w", err)
		}
		for i := range c.AntiAffinityGroups {
			antiAffinityGroup, err := af.FindAntiAffinityGroup(c.AntiAffinityGroups[i])
			if err != nil {
				return fmt.Errorf("error retrieving Anti-Affinity Group: %w", err)
			}
			instanceReq.AntiAffinityGroups[i] = v3.AntiAffinityGroup{ID: antiAffinityGroup.ID}
		}
	}

	if c.DeployTarget != "" {
		targets, err := client.ListDeployTargets(ctx)
		if err != nil {
			return fmt.Errorf("error listing Deploy Target: %w", err)
		}
		deployTarget, err := targets.FindDeployTarget(c.DeployTarget)
		if err != nil {
			return fmt.Errorf("error retrieving Deploy Target: %w", err)
		}
		instanceReq.DeployTarget = &v3.DeployTarget{ID: deployTarget.ID}
	}

	instanceTypes, err := client.ListInstanceTypes(ctx)
	if err != nil {
		return fmt.Errorf("error listing instance type: %w", err)
	}

	// c.InstanceType is never empty
	instanceType := utils.ParseInstanceType(c.InstanceType)
	for i, it := range instanceTypes.InstanceTypes {
		if it.Family == instanceType.Family && it.Size == instanceType.Size {
			instanceReq.InstanceType = &instanceTypes.InstanceTypes[i]
			break
		}
	}
	if instanceReq.InstanceType == nil {
		return fmt.Errorf("error retrieving instance type %s: not found", c.InstanceType)
	}

	privateNetworks := make([]v3.PrivateNetwork, len(c.PrivateNetworks))
	if l := len(c.PrivateNetworks); l > 0 {
		pNetworks, err := client.ListPrivateNetworks(ctx)
		if err != nil {
			return fmt.Errorf("error listing Private Network: %w", err)
		}

		for i := range c.PrivateNetworks {
			privateNetwork, err := pNetworks.FindPrivateNetwork(c.PrivateNetworks[i])
			if err != nil {
				return fmt.Errorf("error retrieving Private Network: %w", err)
			}
			privateNetworks[i] = privateNetwork
		}
	}

	if l := len(c.SecurityGroups); l > 0 {
		sgs, err := client.ListSecurityGroups(ctx)
		if err != nil {
			return fmt.Errorf("error listing Security Group: %w", err)
		}
		instanceReq.SecurityGroups = make([]v3.SecurityGroup, l)
		for i := range c.SecurityGroups {
			securityGroup, err := sgs.FindSecurityGroup(c.SecurityGroups[i])
			if err != nil {
				return fmt.Errorf("error retrieving Security Group: %w", err)
			}
			instanceReq.SecurityGroups[i] = v3.SecurityGroup{ID: securityGroup.ID}
		}
	}

	if instanceReq.SSHKeys == nil && account.CurrentAccount.DefaultSSHKey != "" {
		instanceReq.SSHKeys = []v3.SSHKey{{Name: account.CurrentAccount.DefaultSSHKey}}
	}

	// Generating a single-use SSH key pair for this instance.
	if instanceReq.SSHKeys == nil {
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

		sshKeyName := fmt.Sprintf("%s-%d", c.Name, time.Now().Unix())
		op, err := client.RegisterSSHKey(
			ctx,
			v3.RegisterSSHKeyRequest{
				Name:      sshKeyName,
				PublicKey: string(ssh.MarshalAuthorizedKey(singleUseSSHPublicKey)),
			},
		)
		if err != nil {
			return fmt.Errorf("error registering SSH key: %w", err)
		}
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		if err != nil {
			return fmt.Errorf("error wait registering SSH key: %w", err)
		}

		instanceReq.SSHKeys = []v3.SSHKey{{Name: sshKeyName}}
	}

	templates, err := client.ListTemplates(ctx, v3.ListTemplatesWithVisibility(v3.ListTemplatesVisibility(visibility)))
	if err != nil {
		return fmt.Errorf("error listing template with visibility %q: %w", visibility, err)
	}
	template, err := templates.FindTemplate(templateID)
	if err != nil {
		return fmt.Errorf(
			"no template %q found with visibility %s in zone %s",
			templateID,
			visibility,
			c.Zone,
		)
	}
	instanceReq.Template = &v3.Template{ID: template.ID}

	// Build the cloud-init user data
	// Make sure we have valid input parameters
	if c.HuggingFaceSecret == "" {
		return fmt.Errorf("HuggingFace secret cannot be empty")
	}

	// Set the volume size to 25GB if not specified
	// or set it to half of the disk size if specified
	// (minimum 25GB)
	var volumeSize int64
	if c.DiskSize == 0 {
		volumeSize = 25
	} else {
		volumeSize = c.DiskSize / 2
	}

	aiJobConfig := fmt.Sprintf(aiJobPVCTemplate,
		c.ContainerImage,
		c.Model,
		volumeSize,
		c.HuggingFaceSecret,
	)

	userData, err := userdata.EncodeUserData([]byte(aiJobConfig), false)
	if err != nil {
		return fmt.Errorf("error encoding cloud-init user data: %w", err)

	}
	instanceReq.UserData = userData

	var instanceID v3.UUID
	decorateAsyncOperation(fmt.Sprintf("Creating instance %q...", c.Name), func() {
		var op *v3.Operation
		op, err = client.CreateInstance(ctx, instanceReq)
		if err != nil {
			return
		}

		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		if err != nil {
			return
		}
		if op.Reference != nil {
			instanceID = op.Reference.ID
		}

		for _, p := range privateNetworks {
			op, err = client.AttachInstanceToPrivateNetwork(ctx, p.ID, v3.AttachInstanceToPrivateNetworkRequest{
				Instance: &v3.AttachInstanceToPrivateNetworkRequestInstance{ID: instanceID},
			})
			if err != nil {
				return
			}
			_, err = client.Wait(ctx, op)
			if err != nil {
				return
			}
		}

		if c.Protection {
			var value v3.UUID
			var op *v3.Operation
			value, err = v3.ParseUUID(instanceID.String())
			if err != nil {
				return
			}
			op, err = globalstate.EgoscaleV3Client.AddInstanceProtection(ctx, value)
			if err != nil {
				return
			}
			_, err = globalstate.EgoscaleV3Client.Wait(ctx, op, v3.OperationStateSuccess)

		}
	})
	if err != nil {
		return err
	}

	if singleUseSSHPrivateKey != nil {
		privateKeyFilePath := exossh.GetInstanceSSHKeyPath(instanceID.String())

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

		op, err := client.DeleteSSHKey(ctx, instanceReq.SSHKeys[0].Name)
		if err != nil {
			return fmt.Errorf("error deleting SSH key: %w", err)
		}
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		if err != nil {
			return fmt.Errorf("error wait deleting SSH key: %w", err)
		}
	}

	if !globalstate.Quiet {
		return (&instanceShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			Instance:           instanceID.String(),
			// TODO migrate instanceShow to v3 to pass v3.ZoneName
			Zone: string(c.Zone),
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(aiJobCmd, &aiJobCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
