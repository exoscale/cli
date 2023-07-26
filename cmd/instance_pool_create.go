package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/userdata"
	"github.com/exoscale/cli/utils"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
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
	PrivateNetworks    []string          `cli-flag:"private-network" cli-usage:"managed Compute instances Private Network NAME|ID (can be specified multiple times)"`
	SSHKey             string            `cli-flag:"ssh-key" cli-usage:"SSH key to deploy on managed Compute instances"`
	SecurityGroups     []string          `cli-flag:"security-group" cli-short:"s" cli-usage:"managed Compute instances Security Group NAME|ID (can be specified multiple times)"`
	Size               int64             `cli-usage:"Instance Pool size"`
	Template           string            `cli-short:"t" cli-usage:"managed Compute instances template NAME|ID"`
	TemplateVisibility string            `cli-usage:"instance template visibility (public|private)"`
	Zone               string            `cli-short:"z" cli-usage:"Instance Pool zone"`
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
	instancePool := &egoscale.InstancePool{
		Description:    utils.NonEmptyStringPtr(c.Description),
		DiskSize:       &c.DiskSize,
		IPv6Enabled:    &c.IPv6,
		InstancePrefix: utils.NonEmptyStringPtr(c.InstancePrefix),
		Labels: func() (v *map[string]string) {
			if len(c.Labels) > 0 {
				return &c.Labels
			}
			return
		}(),
		Name:   &c.Name,
		SSHKey: utils.NonEmptyStringPtr(c.SSHKey),
		Size:   &c.Size,
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
		instancePool.AntiAffinityGroupIDs = &antiAffinityGroupIDs
	}

	if c.DeployTarget != "" {
		deployTarget, err := globalstate.EgoscaleClient.FindDeployTarget(ctx, c.Zone, c.DeployTarget)
		if err != nil {
			return fmt.Errorf("error retrieving Deploy Target: %w", err)
		}
		instancePool.DeployTargetID = deployTarget.ID
	}

	if l := len(c.ElasticIPs); l > 0 {
		elasticIPIDs := make([]string, l)
		for i := range c.ElasticIPs {
			elasticIP, err := globalstate.EgoscaleClient.FindElasticIP(ctx, c.Zone, c.ElasticIPs[i])
			if err != nil {
				return fmt.Errorf("error retrieving Elastic IP: %w", err)
			}
			elasticIPIDs[i] = *elasticIP.ID
		}
		instancePool.ElasticIPIDs = &elasticIPIDs
	}

	instanceType, err := globalstate.EgoscaleClient.FindInstanceType(ctx, c.Zone, c.InstanceType)
	if err != nil {
		return fmt.Errorf("error retrieving instance type: %w", err)
	}
	instancePool.InstanceTypeID = instanceType.ID

	if l := len(c.PrivateNetworks); l > 0 {
		privateNetworkIDs := make([]string, l)
		for i := range c.PrivateNetworks {
			privateNetwork, err := globalstate.EgoscaleClient.FindPrivateNetwork(ctx, c.Zone, c.PrivateNetworks[i])
			if err != nil {
				return fmt.Errorf("error retrieving Private Network: %w", err)
			}
			privateNetworkIDs[i] = *privateNetwork.ID
		}
		instancePool.PrivateNetworkIDs = &privateNetworkIDs
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
		instancePool.SecurityGroupIDs = &securityGroupIDs
	}

	if instancePool.SSHKey == nil && account.CurrentAccount.DefaultSSHKey != "" {
		instancePool.SSHKey = &account.CurrentAccount.DefaultSSHKey
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
	instancePool.TemplateID = template.ID

	if c.CloudInitFile != "" {
		userData, err := userdata.GetUserDataFromFile(c.CloudInitFile, c.CloudInitCompress)
		if err != nil {
			return fmt.Errorf("error parsing cloud-init user data: %w", err)
		}
		instancePool.UserData = &userData
	}

	decorateAsyncOperation(fmt.Sprintf("Creating Instance Pool %q...", c.Name), func() {
		instancePool, err = globalstate.EgoscaleClient.CreateInstancePool(ctx, c.Zone, instancePool)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&instancePoolShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			Zone:               c.Zone,
			InstancePool:       *instancePool.ID,
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
		TemplateVisibility: defaultTemplateVisibility,
	}))
}
