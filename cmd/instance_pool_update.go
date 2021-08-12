package cmd

import (
	"fmt"
	"os"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instancePoolUpdateCmd struct {
	_ bool `cli-cmd:"update"`

	InstancePool string `cli-arg:"#" cli-usage:"NAME|ID"`

	AntiAffinityGroups []string          `cli-flag:"anti-affinity-group" cli-short:"a" cli-usage:"managed Compute instances Anti-Affinity Group NAME|ID (can be specified multiple times)"`
	CloudInitFile      string            `cli-flag:"cloud-init" cli-short:"c" cli-usage:"cloud-init user data configuration file path"`
	DeployTarget       string            `cli-usage:"managed Compute instances Deploy Target NAME|ID"`
	Description        string            `cli-usage:"Instance Pool description"`
	DiskSize           int64             `cli-flag:"disk" cli-short:"d" cli-usage:"managed Compute instances disk size"`
	ElasticIPs         []string          `cli-flag:"elastic-ip" cli-short:"e" cli-usage:"managed Compute instances Elastic IP ADDRESS|ID (can be specified multiple times)"`
	IPv6               bool              `cli-flag:"ipv6" cli-short:"6" cli-usage:"enable IPv6 on managed Compute instances"`
	InstancePrefix     string            `cli-usage:"string to prefix managed Compute instances names with"`
	InstanceType       string            `cli-flag:"service-offering" cli-short:"o" cli-usage:"managed Compute instances type"`
	Labels             map[string]string `cli-flag:"label" cli-usage:"Instance Pool label (format: key=value)"`
	Name               string            `cli-short:"n" cli-usage:"Instance Pool name"`
	PrivateNetworks    []string          `cli-flag:"privnet" cli-short:"p" cli-usage:"managed Compute instances Private Network NAME|ID (can be specified multiple times)"`
	SSHKey             string            `cli-short:"k" cli-flag:"keypair" cli-usage:"SSH key to deploy on managed Compute instances"`
	SecurityGroups     []string          `cli-flag:"security-group" cli-short:"s" cli-usage:"managed Compute instances Security Group NAME|ID (can be specified multiple times)"`
	Size               int64             `cli-usage:"Instance Pool size. Deprecated, replaced by the 'exo instancepool scale' command."`
	Template           string            `cli-short:"t" cli-usage:"managed Compute instances template NAME|ID"`
	TemplateFilter     string            `cli-usage:"managed Compute instances template filter"`
	Zone               string            `cli-short:"z" cli-usage:"Instance Pool zone"`
}

func (c *instancePoolUpdateCmd) cmdAliases() []string { return nil }

func (c *instancePoolUpdateCmd) cmdShort() string { return "Update an Instance Pool" }

func (c *instancePoolUpdateCmd) cmdLong() string {
	return fmt.Sprintf(`This command updates an Instance Pool.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&instancePoolShowOutput{}), ", "),
	)
}

func (c *instancePoolUpdateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instancePoolUpdateCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	zoneV1, err := getZoneByNameOrID(c.Zone)
	if err != nil {
		return err
	}

	instancePool, err := cs.FindInstancePool(ctx, c.Zone, c.InstancePool)
	if err != nil {
		return err
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.AntiAffinityGroups)) {
		antiAffinityGroupIDs := make([]string, len(c.AntiAffinityGroups))
		for i, v := range c.AntiAffinityGroups {
			antiAffinityGroup, err := cs.FindAntiAffinityGroup(ctx, c.Zone, v)
			if err != nil {
				return fmt.Errorf("error retrieving Anti-Affinity Group: %s", err)
			}
			antiAffinityGroupIDs[i] = *antiAffinityGroup.ID
		}
		instancePool.AntiAffinityGroupIDs = &antiAffinityGroupIDs
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.DeployTarget)) {
		deployTarget, err := cs.FindDeployTarget(ctx, c.Zone, c.DeployTarget)
		if err != nil {
			return fmt.Errorf("error retrieving Deploy Target: %s", err)
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
			elasticIP, err := cs.FindElasticIP(ctx, c.Zone, v)
			if err != nil {
				return fmt.Errorf("error retrieving Elastic IP: %s", err)
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
			privateNetwork, err := cs.FindPrivateNetwork(ctx, c.Zone, v)
			if err != nil {
				return fmt.Errorf("error retrieving Private Network: %s", err)
			}
			privateNetworkIDs[i] = *privateNetwork.ID
		}
		instancePool.PrivateNetworkIDs = &privateNetworkIDs
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.SecurityGroups)) {
		securityGroupIDs := make([]string, len(c.SecurityGroups))
		for i, v := range c.SecurityGroups {
			securityGroup, err := cs.FindSecurityGroup(ctx, c.Zone, v)
			if err != nil {
				return fmt.Errorf("error retrieving Security Group: %s", err)
			}
			securityGroupIDs[i] = *securityGroup.ID
		}
		instancePool.SecurityGroupIDs = &securityGroupIDs
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.InstanceType)) {
		instanceType, err := cs.FindInstanceType(ctx, c.Zone, c.InstanceType)
		if err != nil {
			return fmt.Errorf("error retrieving instance type: %s", err)
		}
		instancePool.InstanceTypeID = instanceType.ID
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.SSHKey)) {
		instancePool.SSHKey = &c.SSHKey
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Template)) {
		templateFilter, err := validateTemplateFilter(c.TemplateFilter)
		if err != nil {
			return err
		}

		templateFlagVal, err := cmd.Flags().GetString("template")
		if err != nil {
			return err
		}
		template, err := getTemplateByNameOrID(zoneV1.ID, templateFlagVal, templateFilter)
		if err != nil {
			return fmt.Errorf("error retrieving template: %s", err)
		}
		templateID := template.ID.String()
		instancePool.TemplateID = &templateID
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.CloudInitFile)) {
		userData, err := getUserDataFromFile(c.CloudInitFile)
		if err != nil {
			return fmt.Errorf("error parsing cloud-init user data: %s", err)
		}
		instancePool.UserData = &userData
		updated = true
	}

	if updated {
		decorateAsyncOperation(fmt.Sprintf("Updating Instance Pool %q...", c.InstancePool), func() {
			if err = cs.UpdateInstancePool(ctx, c.Zone, instancePool); err != nil {
				return
			}
		})
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Size)) {
		_, _ = fmt.Fprintln(
			os.Stderr,
			`WARNING: the "--size" flag is deprecated and replaced by the `+
				`"exo instancepool scale" command, it will be removed in a future version.`,
		)

		decorateAsyncOperation(fmt.Sprintf("Scaling Instance Pool %q...", c.InstancePool), func() {
			err = cs.ScaleInstancePool(ctx, c.Zone, instancePool, c.Size)
		})
	}

	if !gQuiet {
		return output(showInstancePool(c.Zone, *instancePool.ID))
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(instancePoolCmd, &instancePoolUpdateCmd{
		TemplateFilter: defaultTemplateFilter,
	}))
}
