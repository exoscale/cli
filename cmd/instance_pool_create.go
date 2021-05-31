package cmd

import (
	"fmt"
	"strings"

	exov2 "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instancePoolCreateCmd struct {
	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#" cli-usage:"NAME"`

	AntiAffinityGroups []string `cli-flag:"anti-affinity-group" cli-short:"a" cli-usage:"managed Compute instances Anti-Affinity Group NAME|ID (can be specified multiple times)"`
	CloudInitFile      string   `cli-flag:"cloud-init" cli-short:"c" cli-usage:"cloud-init user data configuration file path"`
	DeployTarget       string   `cli-usage:"managed Compute instances Deploy Target NAME|ID"`
	Description        string   `cli-usage:"Instance Pool description"`
	DiskSize           int64    `cli-flag:"disk" cli-short:"d" cli-usage:"managed Compute instances disk size"`
	ElasticIPs         []string `cli-flag:"elastic-ip" cli-short:"e" cli-usage:"managed Compute instances Elastic IP ADDRESS|ID (can be specified multiple times)"`
	IPv6               bool     `cli-flag:"ipv6" cli-short:"6" cli-usage:"enable IPv6 on managed Compute instances"`
	InstancePrefix     string   `cli-usage:"string to prefix managed Compute instances names with"`
	InstanceType       string   `cli-flag:"service-offering" cli-short:"o" cli-usage:"managed Compute instances type"`
	PrivateNetworks    []string `cli-flag:"privnet" cli-short:"p" cli-usage:"managed Compute instances Private Network NAME|ID (can be specified multiple times)"`
	SSHKey             string   `cli-short:"k" cli-flag:"keypair" cli-usage:"SSH key to deploy on managed Compute instances"`
	SecurityGroups     []string `cli-flag:"security-group" cli-short:"s" cli-usage:"managed Compute instances Security Group NAME|ID (can be specified multiple times)"`
	Size               int64    `cli-usage:"Instance Pool size"`
	Template           string   `cli-short:"t" cli-usage:"managed Compute instances template NAME|ID"`
	TemplateFilter     string   `cli-usage:"managed Compute instances template filter"`
	Zone               string   `cli-short:"z" cli-usage:"Instance Pool zone"`
}

func (c *instancePoolCreateCmd) cmdAliases() []string { return gCreateAlias }

func (c *instancePoolCreateCmd) cmdShort() string { return "Create an Instance Pool" }

func (c *instancePoolCreateCmd) cmdLong() string {
	return fmt.Sprintf(`This command creates an Instance Pool.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&instancePoolShowOutput{}), ", "))
}

func (c *instancePoolCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instancePoolCreateCmd) cmdRun(_ *cobra.Command, _ []string) error {
	instancePool := &exov2.InstancePool{
		DeployTargetID: c.DeployTarget,
		Description:    c.Description,
		DiskSize:       c.DiskSize,
		IPv6Enabled:    c.IPv6,
		InstancePrefix: c.InstancePrefix,
		Name:           c.Name,
		SSHKey:         c.SSHKey,
		Size:           c.Size,
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	zoneV1, err := getZoneByNameOrID(c.Zone)
	if err != nil {
		return err
	}

	if l := len(c.AntiAffinityGroups); l > 0 {
		instancePool.AntiAffinityGroupIDs = make([]string, l)
		for i := range c.AntiAffinityGroups {
			antiAffinityGroup, err := cs.FindAntiAffinityGroup(ctx, c.Zone, c.AntiAffinityGroups[i])
			if err != nil {
				return fmt.Errorf("error retrieving Anti-Affinity Group: %s", err)
			}
			instancePool.AntiAffinityGroupIDs[i] = antiAffinityGroup.ID
		}
	}

	if c.DeployTarget != "" {
		deployTarget, err := cs.FindDeployTarget(ctx, c.Zone, c.DeployTarget)
		if err != nil {
			return fmt.Errorf("error retrieving Deploy Target: %s", err)
		}
		instancePool.DeployTargetID = deployTarget.ID
	}

	if l := len(c.ElasticIPs); l > 0 {
		instancePool.ElasticIPIDs = make([]string, l)
		for i := range c.ElasticIPs {
			elasticIP, err := cs.FindElasticIP(ctx, c.Zone, c.ElasticIPs[i])
			if err != nil {
				return fmt.Errorf("error retrieving Elastic IP: %s", err)
			}
			instancePool.ElasticIPIDs[i] = elasticIP.ID
		}
	}

	instanceType, err := cs.FindInstanceType(ctx, c.Zone, c.InstanceType)
	if err != nil {
		return fmt.Errorf("error retrieving instance type: %s", err)
	}
	instancePool.InstanceTypeID = instanceType.ID

	if l := len(c.PrivateNetworks); l > 0 {
		instancePool.PrivateNetworkIDs = make([]string, l)
		for i := range c.PrivateNetworks {
			privateNetwork, err := cs.FindPrivateNetwork(ctx, c.Zone, c.PrivateNetworks[i])
			if err != nil {
				return fmt.Errorf("error retrieving Private Network: %s", err)
			}
			instancePool.PrivateNetworkIDs[i] = privateNetwork.ID
		}
	}

	if l := len(c.SecurityGroups); l > 0 {
		instancePool.SecurityGroupIDs = make([]string, l)
		for i := range c.SecurityGroups {
			securityGroup, err := cs.FindSecurityGroup(ctx, c.Zone, c.SecurityGroups[i])
			if err != nil {
				return fmt.Errorf("error retrieving Security Group: %s", err)
			}
			instancePool.SecurityGroupIDs[i] = securityGroup.ID
		}
	}

	if instancePool.SSHKey == "" {
		instancePool.SSHKey = gCurrentAccount.DefaultSSHKey
	}

	templateFilter, err := validateTemplateFilter(c.TemplateFilter)
	if err != nil {
		return err
	}

	template, err := getTemplateByNameOrID(zoneV1.ID, c.Template, templateFilter)
	if err != nil {
		return fmt.Errorf("error retrieving template: %s", err)
	}
	instancePool.TemplateID = template.ID.String()

	if c.CloudInitFile != "" {
		if instancePool.UserData, err = getUserDataFromFile(c.CloudInitFile); err != nil {
			return fmt.Errorf("error parsing cloud-init user data: %s", err)
		}
	}

	decorateAsyncOperation(fmt.Sprintf("Creating Instance Pool %q...", instancePool.Name), func() {
		instancePool, err = cs.CreateInstancePool(ctx, c.Zone, instancePool)
	})
	if err != nil {
		return err
	}

	if !gQuiet {
		return output(showInstancePool(c.Zone, instancePool.ID))
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(instancePoolCmd, &instancePoolCreateCmd{
		DiskSize:       50,
		InstanceType:   defaultServiceOffering,
		Size:           1,
		Template:       defaultTemplate,
		TemplateFilter: defaultTemplateFilter,
	}))
}
