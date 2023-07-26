package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/userdata"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type instancePoolUpdateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	InstancePool string `cli-arg:"#" cli-usage:"NAME|ID"`

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
	Name               string            `cli-short:"n" cli-usage:"Instance Pool name"`
	PrivateNetworks    []string          `cli-flag:"private-network" cli-usage:"managed Compute instances Private Network NAME|ID (can be specified multiple times)"`
	SSHKey             string            `cli-flag:"ssh-key" cli-usage:"SSH key to deploy on managed Compute instances"`
	SecurityGroups     []string          `cli-flag:"security-group" cli-short:"s" cli-usage:"managed Compute instances Security Group NAME|ID (can be specified multiple times)"`
	Template           string            `cli-short:"t" cli-usage:"managed Compute instances template NAME|ID"`
	TemplateVisibility string            `cli-usage:"instance template visibility (public|private)"`
	Zone               string            `cli-short:"z" cli-usage:"Instance Pool zone"`
}

func (c *instancePoolUpdateCmd) cmdAliases() []string { return nil }

func (c *instancePoolUpdateCmd) cmdShort() string { return "Update an Instance Pool" }

func (c *instancePoolUpdateCmd) cmdLong() string {
	return fmt.Sprintf(`This command updates an Instance Pool.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instancePoolShowOutput{}), ", "),
	)
}

func (c *instancePoolUpdateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instancePoolUpdateCmd) cmdRun(cmd *cobra.Command, _ []string) error { //nolint:gocyclo
	var updated bool

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	instancePool, err := globalstate.EgoscaleClient.FindInstancePool(ctx, c.Zone, c.InstancePool)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.AntiAffinityGroups)) {
		antiAffinityGroupIDs := make([]string, len(c.AntiAffinityGroups))
		for i, v := range c.AntiAffinityGroups {
			antiAffinityGroup, err := globalstate.EgoscaleClient.FindAntiAffinityGroup(ctx, c.Zone, v)
			if err != nil {
				return fmt.Errorf("error retrieving Anti-Affinity Group: %w", err)
			}
			antiAffinityGroupIDs[i] = *antiAffinityGroup.ID
		}
		instancePool.AntiAffinityGroupIDs = &antiAffinityGroupIDs
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.DeployTarget)) {
		deployTarget, err := globalstate.EgoscaleClient.FindDeployTarget(ctx, c.Zone, c.DeployTarget)
		if err != nil {
			return fmt.Errorf("error retrieving Deploy Target: %w", err)
		}
		instancePool.DeployTargetID = deployTarget.ID
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Description)) {
		instancePool.Description = &c.Description
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.DiskSize)) {
		instancePool.DiskSize = &c.DiskSize
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.ElasticIPs)) {
		elasticIPIDs := make([]string, len(c.ElasticIPs))
		for i, v := range c.ElasticIPs {
			elasticIP, err := globalstate.EgoscaleClient.FindElasticIP(ctx, c.Zone, v)
			if err != nil {
				return fmt.Errorf("error retrieving Elastic IP: %w", err)
			}
			elasticIPIDs[i] = *elasticIP.ID
		}
		instancePool.ElasticIPIDs = &elasticIPIDs
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.InstancePrefix)) {
		instancePool.InstancePrefix = &c.InstancePrefix
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.IPv6)) {
		instancePool.IPv6Enabled = &c.IPv6
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Labels)) {
		instancePool.Labels = &c.Labels
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Name)) {
		instancePool.Name = &c.Name
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.PrivateNetworks)) {
		privateNetworkIDs := make([]string, len(c.PrivateNetworks))
		for i, v := range c.PrivateNetworks {
			privateNetwork, err := globalstate.EgoscaleClient.FindPrivateNetwork(ctx, c.Zone, v)
			if err != nil {
				return fmt.Errorf("error retrieving Private Network: %w", err)
			}
			privateNetworkIDs[i] = *privateNetwork.ID
		}
		instancePool.PrivateNetworkIDs = &privateNetworkIDs
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.SecurityGroups)) {
		securityGroupIDs := make([]string, len(c.SecurityGroups))
		for i, v := range c.SecurityGroups {
			securityGroup, err := globalstate.EgoscaleClient.FindSecurityGroup(ctx, c.Zone, v)
			if err != nil {
				return fmt.Errorf("error retrieving Security Group: %w", err)
			}
			securityGroupIDs[i] = *securityGroup.ID
		}
		instancePool.SecurityGroupIDs = &securityGroupIDs
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.InstanceType)) {
		instanceType, err := globalstate.EgoscaleClient.FindInstanceType(ctx, c.Zone, c.InstanceType)
		if err != nil {
			return fmt.Errorf("error retrieving instance type: %w", err)
		}
		instancePool.InstanceTypeID = instanceType.ID
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.SSHKey)) {
		instancePool.SSHKey = &c.SSHKey
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Template)) {
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
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.CloudInitFile)) {
		userData, err := userdata.GetUserDataFromFile(c.CloudInitFile, c.CloudInitCompress)
		if err != nil {
			return fmt.Errorf("error parsing cloud-init user data: %w", err)
		}
		instancePool.UserData = &userData
		updated = true
	}

	if updated {
		decorateAsyncOperation(fmt.Sprintf("Updating Instance Pool %q...", c.InstancePool), func() {
			if err = globalstate.EgoscaleClient.UpdateInstancePool(ctx, c.Zone, instancePool); err != nil {
				return
			}
		})
		if err != nil {
			return err
		}
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
	cobra.CheckErr(registerCLICommand(instancePoolCmd, &instancePoolUpdateCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		TemplateVisibility: defaultTemplateVisibility,
	}))
}
