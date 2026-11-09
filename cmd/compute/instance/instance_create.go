package instance

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/cmd/compute"
	"github.com/exoscale/cli/cmd/networking/vpc"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	exossh "github.com/exoscale/cli/pkg/ssh"
	"github.com/exoscale/cli/pkg/userdata"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceCreateCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#" cli-usage:"NAME"`

	AppConsistentSnapshot bool              `cli-flag:"application-consistent-snapshot-enabled" cli-usage:"enable application-consistent snapshots when supported; false disables; omit for template default"`
	AntiAffinityGroups    []string          `cli-flag:"anti-affinity-group" cli-usage:"instance Anti-Affinity Group NAME|ID (can be specified multiple times)"`
	CloudInitFile         string            `cli-flag:"cloud-init" cli-usage:"instance cloud-init user data configuration file path"`
	CloudInitCompress     bool              `cli-flag:"cloud-init-compress" cli-usage:"compress instance cloud-init user data"`
	DeployTarget          string            `cli-usage:"instance Deploy Target NAME|ID"`
	DiskSize              int64             `cli-usage:"instance disk size"`
	TPM                   bool              `cli-flag:"tpm" cli-usage:"enable TPM on instance"`
	IpForwarding          bool              `cli-flag:"ip-forwarding" cli-usage:"enable ip forwarding on instance"`
	SecureBoot            bool              `cli-flag:"secureboot" cli-usage:"enable Secure boot on instance"`
	InstanceType          string            `cli-usage:"instance type (format: [FAMILY.]SIZE)"`
	Labels                map[string]string `cli-flag:"label" cli-usage:"instance label (format: key=value)"`
	PrivateNetworks       []string          `cli-flag:"private-network" cli-usage:"instance Private Network NAME|ID (can be specified multiple times)"`
	PublicIPAssignment    string            `cli-flag:"public-ip" cli-usage:"Configures public IP assignment of the Instances (none|inet4|dual). (default: inet4)"`
	ReverseDNS            string            `cli-usage:"Reverse DNS Domain"`
	SSHKeys               []string          `cli-flag:"ssh-key" cli-usage:"SSH key to deploy on the instance (can be specified multiple times)"`
	Protection            bool              `cli-flag:"protection" cli-usage:"enable delete protection"`
	SecurityGroups        []string          `cli-flag:"security-group" cli-usage:"instance Security Group NAME|ID (can be specified multiple times)"`
	Template              string            `cli-usage:"instance template NAME|ID"`
	TemplateVisibility    string            `cli-usage:"instance template visibility (public|private)"`
	VPC                   string            `cli-flag:"vpc" cli-usage:"VPC NAME|ID to attach the instance to (requires --vpc-subnet)"`
	VPCSubnets            vpcSubnetSpecs    `cli-flag:"vpc-subnet" cli-usage:"VPC Subnet to attach the instance to (format: NAME|ID[,ipv4=IP]; can be specified multiple times)"`
	Zone                  v3.ZoneName       `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceCreateCmd) CmdAliases() []string { return clicmd.GCreateAlias }

func (c *instanceCreateCmd) CmdShort() string { return "Create a Compute instance" }

func (c *instanceCreateCmd) CmdLong() string {
	return fmt.Sprintf(`This command creates a Compute instance.

Supported Compute instance type families: %s

Supported Compute instance type sizes: %s

Supported output template annotations: %s`,
		strings.Join(instanceTypeFamilies, ", "),
		strings.Join(instanceTypeSizes, ", "),
		strings.Join(output.TemplateAnnotations(&InstanceShowOutput{}), ", "))
}

func (c *instanceCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	clicmd.CmdSetZoneFlagFromDefault(cmd)
	clicmd.CmdSetTemplateFlagFromDefault(cmd)
	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceCreateCmd) CmdRun(cmd *cobra.Command, _ []string) error { //nolint:gocyclo
	var (
		singleUseSSHPrivateKey *rsa.PrivateKey
		singleUseSSHPublicKey  ssh.PublicKey
	)

	if err := validateVPCFlags(c.VPC, c.VPCSubnets); err != nil {
		return err
	}

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	template, err := compute.ResolveTemplate(ctx, client, c.Template, c.TemplateVisibility, c.Zone)
	if err != nil {
		return err
	}
	diskSize, err := compute.ResolveTemplateDiskSize(
		c.DiskSize,
		template.Size,
		cmd.Flags().Changed(clicmd.MustCLICommandFlagName(c, &c.DiskSize)),
	)
	if err != nil {
		return err
	}
	if diskSize != c.DiskSize {
		fmt.Fprintf(
			os.Stderr,
			"warning: template %q requires at least %d GiB; increasing the default disk size from %d GiB to %d GiB\n",
			template.Name,
			diskSize,
			c.DiskSize,
			diskSize,
		)
	}

	var sshKeys []v3.SSHKey
	for _, sshkeyName := range c.SSHKeys {
		sshKeys = append(sshKeys, v3.SSHKey{Name: sshkeyName})
	}

	publicIPAssignment := v3.PublicIPAssignmentInet4
	if c.PublicIPAssignment != "" {
		if !slices.Contains([]v3.PublicIPAssignment{
			v3.PublicIPAssignmentDual, v3.PublicIPAssignmentInet4, v3.PublicIPAssignmentNone,
		}, v3.PublicIPAssignment(c.PublicIPAssignment)) {
			return fmt.Errorf("error invalid public-ip: %s", c.PublicIPAssignment)
		}
		publicIPAssignment = v3.PublicIPAssignment(c.PublicIPAssignment)
	}

	instanceReq := v3.CreateInstanceRequest{
		DiskSize:           diskSize,
		PublicIPAssignment: publicIPAssignment,
		TpmEnabled:         &c.TPM,
		IPForwarding:       &c.IpForwarding,
		SecurebootEnabled:  &c.SecureBoot,
		Labels:             c.Labels,
		Name:               c.Name,
		SSHKeys:            sshKeys,
		Template:           &v3.Template{ID: template.ID},
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

	if c.VPC != "" {
		instanceReq.Vpc, err = resolveCreateInstanceVPC(ctx, client, c.VPC, c.VPCSubnets)
		if err != nil {
			return err
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

	if c.CloudInitFile != "" {
		userData, err := userdata.GetUserDataFromFile(c.CloudInitFile, c.CloudInitCompress)
		if err != nil {
			return fmt.Errorf("error parsing cloud-init user data: %w", err)
		}
		instanceReq.UserData = userData
	}

	if cmd.Flags().Changed(clicmd.MustCLICommandFlagName(c, &c.AppConsistentSnapshot)) {
		instanceReq.ApplicationConsistentSnapshotEnabled = &c.AppConsistentSnapshot
	}

	updateRDNS := cmd.Flags().Changed(clicmd.MustCLICommandFlagName(c, &c.ReverseDNS))

	var instanceID v3.UUID
	utils.DecorateAsyncOperation(fmt.Sprintf("Creating instance %q...", c.Name), func() {
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

		if updateRDNS {
			op, err = client.UpdateReverseDNSInstance(ctx, instanceID, v3.UpdateReverseDNSInstanceRequest{DomainName: c.ReverseDNS})
			if err != nil {
				return
			}
			_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
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
			op, err = client.AddInstanceProtection(ctx, value)
			if err != nil {
				return
			}
			_, err = client.Wait(ctx, op, v3.OperationStateSuccess)

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
			CliCommandSettings: c.CliCommandSettings,
			Instance:           instanceID.String(),
			Zone:               c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

// resolveCreateInstanceVPC builds the VPC attachment of an instance creation
// request from a VPC NAME|ID and a list of subnet specs.
func resolveCreateInstanceVPC(
	ctx context.Context,
	client *v3.Client,
	vpcNameOrID string,
	subnetSpecs []vpcSubnetSpec,
) (*v3.CreateInstanceRequestVpc, error) {
	vpcEntry, err := vpc.FindVPC(ctx, client, vpcNameOrID)
	if err != nil {
		return nil, err
	}

	subnets, err := client.ListSubnets(ctx, vpcEntry.ID)
	if err != nil {
		return nil, fmt.Errorf("error listing VPC Subnets: %w", err)
	}

	req := &v3.CreateInstanceRequestVpc{ID: vpcEntry.ID}
	for _, spec := range subnetSpecs {
		subnet, err := subnets.FindListSubnetEntry(spec.NameOrID)
		if err != nil {
			if errors.Is(err, v3.ErrNotFound) {
				return nil, fmt.Errorf("subnet %q not found in VPC %q", spec.NameOrID, vpcNameOrID)
			}
			return nil, fmt.Errorf("error retrieving VPC Subnet: %w", err)
		}

		req.Subnets = append(req.Subnets, v3.CreateInstanceRequestVpcSubnets{ID: subnet.ID, Ipv4: spec.IPv4})
	}

	return req, nil
}

func validateVPCFlags(vpcNameOrID string, subnetSpecs []vpcSubnetSpec) error {
	if vpcNameOrID == "" && len(subnetSpecs) > 0 {
		return fmt.Errorf("--vpc is required when --vpc-subnet is specified")
	}
	if vpcNameOrID != "" && len(subnetSpecs) == 0 {
		return fmt.Errorf("at least one --vpc-subnet is required when --vpc is specified")
	}
	return nil
}

// vpcSubnetSpec is one --vpc-subnet value: NAME|ID[,ipv4=IP].
type vpcSubnetSpec struct {
	NameOrID string
	IPv4     net.IP
}

// vpcSubnetSpecs implements pflag.Value. Unlike a []string flag, it does not
// split on commas, so each --vpc-subnet occurrence holds one subnet with its
// key=value options.
type vpcSubnetSpecs []vpcSubnetSpec

func (s *vpcSubnetSpecs) Set(v string) error {
	spec, err := parseVPCSubnetSpec(v)
	if err != nil {
		return err
	}
	*s = append(*s, spec)
	return nil
}

func (s *vpcSubnetSpecs) String() string {
	out := make([]string, len(*s))
	for i, spec := range *s {
		out[i] = spec.NameOrID
		if spec.IPv4 != nil {
			out[i] += ",ipv4=" + spec.IPv4.String()
		}
	}
	return strings.Join(out, " ")
}

func (s *vpcSubnetSpecs) Type() string { return "stringArray" }

// parseVPCSubnetSpec parses a NAME|ID[,ipv4=IP] subnet spec.
func parseVPCSubnetSpec(v string) (vpcSubnetSpec, error) {
	parts := strings.Split(v, ",")
	spec := vpcSubnetSpec{NameOrID: parts[0]}
	if spec.NameOrID == "" || strings.Contains(spec.NameOrID, "=") {
		return vpcSubnetSpec{}, fmt.Errorf("missing subnet NAME|ID (format: NAME|ID[,ipv4=IP])")
	}

	for _, opt := range parts[1:] {
		key, value, ok := strings.Cut(opt, "=")
		if !ok {
			return vpcSubnetSpec{}, fmt.Errorf("invalid option %q (expected key=value)", opt)
		}

		switch key {
		case "ipv4":
			if spec.IPv4 != nil {
				return vpcSubnetSpec{}, fmt.Errorf("ipv4 specified more than once")
			}
			ip := net.ParseIP(value)
			if ip == nil || ip.To4() == nil {
				return vpcSubnetSpec{}, fmt.Errorf("invalid IPv4 address %q", value)
			}
			spec.IPv4 = ip.To4()
		default:
			return vpcSubnetSpec{}, fmt.Errorf("unsupported option %q (supported: ipv4)", key)
		}
	}

	return spec, nil
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(instanceCmd, &instanceCreateCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),

		DiskSize:           50,
		InstanceType:       fmt.Sprintf("%s.%s", clicmd.DefaultInstanceTypeFamily, clicmd.DefaultInstanceType),
		TemplateVisibility: clicmd.DefaultTemplateVisibility,
	}))
}
