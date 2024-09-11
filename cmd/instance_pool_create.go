package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/userdata"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type instancePoolCreateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#" cli-usage:"NAME"`

	AntiAffinityGroups []string          `cli-flag:"anti-affinity-group" cli-short:"a" cli-usage:"managed Compute instances Anti-Affinity Group NAME|ID (can be specified multiple times)"`
	CloudInitFile      string            `cli-flag:"cloud-init" cli-short:"c" cli-usage:"cloud-init user data configuration file path"`
	CloudInitCompress  bool              `cli-flag:"cloud-init-compress" cli-usage:"compress instance cloud-init user data"`
	DeployTarget       string            `cli-usage:"managed Compute instances Deploy Target NAME|ID"`
	Description        string            `cli-usage:"Instance Pool description"`
	DiskSize           int64             `cli-usage:"managed Compute instances disk size"`
	ElasticIPs         []string          `cli-flag:"elastic-ip" cli-short:"e" cli-usage:"managed Compute instances Elastic IP ADDRESS|ID (can be specified multiple times)"`
	IPv6               bool              `cli-flag:"ipv6" cli-short:"6" cli-usage:"enable IPv6 on managed Compute instances"`
	InstancePrefix     string            `cli-usage:"string to prefix managed Compute instances names with"`
	InstanceType       string            `cli-usage:"managed Compute instances type (format: [FAMILY.]SIZE)"`
	Labels             map[string]string `cli-flag:"label" cli-usage:"Instance Pool label (format: key=value)"`
	MinAvailable       int64             `cli-usage:"Minimum number of running Instances"`
	PrivateNetworks    []string          `cli-flag:"private-network" cli-usage:"managed Compute instances Private Network NAME|ID (can be specified multiple times)"`
	SSHKey             string            `cli-flag:"ssh-key" cli-usage:"SSH key to deploy on managed Compute instances"`
	SecurityGroups     []string          `cli-flag:"security-group" cli-short:"s" cli-usage:"managed Compute instances Security Group NAME|ID (can be specified multiple times)"`
	Size               int64             `cli-usage:"Instance Pool size"`
	Template           string            `cli-short:"t" cli-usage:"managed Compute instances template NAME|ID"`
	TemplateVisibility string            `cli-usage:"instance template visibility (public|private)"`
	Zone               v3.ZoneName       `cli-short:"z" cli-usage:"Instance Pool zone"`
}

func (c *instancePoolCreateCmd) cmdAliases() []string { return gCreateAlias }

func (c *instancePoolCreateCmd) cmdShort() string { return "Create an Instance Pool" }

func (c *instancePoolCreateCmd) cmdLong() string {
	return fmt.Sprintf(`This command creates an Instance Pool.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instancePoolShowOutput{}), ", "))
}

func (c *instancePoolCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {

	cmdSetZoneFlagFromDefault(cmd)
	cmdSetTemplateFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instancePoolCreateCmd) cmdRun(_ *cobra.Command, _ []string) error {

	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	sshKey := &v3.SSHKey{Name: c.SSHKey}

	instancePoolReq := v3.CreateInstancePoolRequest{
		Description:    c.Description,
		DiskSize:       c.DiskSize,
		Ipv6Enabled:    &c.IPv6,
		InstancePrefix: c.InstancePrefix,
		Labels:         c.Labels,
		MinAvailable:   c.MinAvailable,
		Name:           c.Name,
		SSHKey:         sshKey,
		Size:           c.Size,
	}

	if l := len(c.AntiAffinityGroups); l > 0 {
		instancePoolReq.AntiAffinityGroups = make([]v3.AntiAffinityGroup, l)
		af, err := client.ListAntiAffinityGroups(ctx)
		if err != nil {
			return fmt.Errorf("error listing Anti-Affinity Group: %w", err)
		}
		for i := range c.AntiAffinityGroups {
			antiAffinityGroup, err := af.FindAntiAffinityGroup(c.AntiAffinityGroups[i])
			if err != nil {
				return fmt.Errorf("error retrieving Anti-Affinity Group: %w", err)
			}
			instancePoolReq.AntiAffinityGroups[i] = v3.AntiAffinityGroup{ID: antiAffinityGroup.ID}
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
		instancePoolReq.DeployTarget = &v3.DeployTarget{ID: deployTarget.ID}
	}

	if l := len(c.ElasticIPs); l > 0 {
		result := []v3.ElasticIP{}
		eipList, err := client.ListElasticIPS(ctx)
		if err != nil {
			return fmt.Errorf("error listing Elastic IP: %w", err)
		}
		for _, input := range c.ElasticIPs {
			eip, err := eipList.FindElasticIP(input)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: Elastic IP %s not found.\n", input)
				continue
			}

			result = append(result, v3.ElasticIP{ID: eip.ID})
		}

		if len(result) != 0 {
			instancePoolReq.ElasticIPS = result
		}
	}

	instanceTypes, err := client.ListInstanceTypes(ctx)
	if err != nil {
		return fmt.Errorf("error listing instance type: %w", err)
	}

	// c.InstanceType is never empty
	instanceType := utils.ParseInstanceType(c.InstanceType)
	for i, it := range instanceTypes.InstanceTypes {
		if it.Family == instanceType.Family && it.Size == instanceType.Size {
			instancePoolReq.InstanceType = &instanceTypes.InstanceTypes[i]
			break
		}
	}
	if instancePoolReq.InstanceType == nil {
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
		instancePoolReq.SecurityGroups = make([]v3.SecurityGroup, l)
		for i := range c.SecurityGroups {
			securityGroup, err := sgs.FindSecurityGroup(c.SecurityGroups[i])
			if err != nil {
				return fmt.Errorf("error retrieving Security Group: %w", err)
			}
			instancePoolReq.SecurityGroups[i] = v3.SecurityGroup{ID: securityGroup.ID}
		}
	}

	if instancePoolReq.SSHKey == nil && account.CurrentAccount.DefaultSSHKey != "" {
		instancePoolReq.SSHKey = &v3.SSHKey{Name: account.CurrentAccount.DefaultSSHKey}
	}

	templates, err := client.ListTemplates(ctx, v3.ListTemplatesWithVisibility(v3.ListTemplatesVisibility(c.TemplateVisibility)))
	if err != nil {
		return fmt.Errorf("error listing template with visibility %q: %w", c.TemplateVisibility, err)
	}
	template, err := templates.FindTemplate(c.Template)
	if err != nil {
		return fmt.Errorf(
			"no template %q found with visibility %s in zone %s",
			c.Template,
			c.TemplateVisibility,
			c.Zone,
		)
	}
	instancePoolReq.Template = &v3.Template{ID: template.ID}

	if c.CloudInitFile != "" {
		userData, err := userdata.GetUserDataFromFile(c.CloudInitFile, c.CloudInitCompress)
		if err != nil {
			return fmt.Errorf("error parsing cloud-init user data: %w", err)
		}
		instancePoolReq.UserData = userData
	}

	var instancePoolID v3.UUID

	decorateAsyncOperation(fmt.Sprintf("Creating Instance Pool %q...", c.Name), func() {
		var op *v3.Operation
		op, err = client.CreateInstancePool(ctx, instancePoolReq)
		if err != nil {
			return
		}

		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		if err != nil {
			return
		}
		if op.Reference != nil {
			instancePoolID = op.Reference.ID
		}

	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&instancePoolShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			InstancePool:       instancePoolID.String(),
			// TODO migrate instanceShow to v3 to pass v3.ZoneName
			Zone: string(c.Zone),
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(instancePoolCmd, &instancePoolCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		DiskSize:           50,
		InstanceType:       fmt.Sprintf("%s.%s", defaultInstanceTypeFamily, defaultInstanceType),
		Size:               1,
		MinAvailable:       0,
		TemplateVisibility: defaultTemplateVisibility,
	}))
}
