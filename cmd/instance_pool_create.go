package cmd

import (
	"fmt"
	"strings"

	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
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
	Disk               int64             `cli-flag:"disk" cli-short:"d" cli-usage:"[DEPRECATED] use --disk-size"`
	DiskSize           int64             `cli-usage:"managed Compute instances disk size"`
	ElasticIPs         []string          `cli-flag:"elastic-ip" cli-short:"e" cli-usage:"managed Compute instances Elastic IP ADDRESS|ID (can be specified multiple times)"`
	IPv6               bool              `cli-flag:"ipv6" cli-short:"6" cli-usage:"enable IPv6 on managed Compute instances"`
	InstancePrefix     string            `cli-usage:"string to prefix managed Compute instances names with"`
	InstanceType       string            `cli-usage:"managed Compute instances type (format: [FAMILY.]SIZE)"`
	Keypair            string            `cli-short:"k" cli-usage:"[DEPRECATED] use --ssh-key"`
	Labels             map[string]string `cli-flag:"label" cli-usage:"Instance Pool label (format: key=value)"`
	PrivateNetworks    []string          `cli-flag:"private-network" cli-usage:"managed Compute instances Private Network NAME|ID (can be specified multiple times)"`
	Privnet            []string          `cli-short:"p" cli-usage:"[DEPRECATED] use --private-network"`
	SSHKey             string            `cli-flag:"ssh-key" cli-usage:"SSH key to deploy on managed Compute instances"`
	SecurityGroups     []string          `cli-flag:"security-group" cli-short:"s" cli-usage:"managed Compute instances Security Group NAME|ID (can be specified multiple times)"`
	ServiceOffering    string            `cli-short:"o" cli-usage:"[DEPRECATED] use --instance-type"`
	Size               int64             `cli-usage:"Instance Pool size"`
	Template           string            `cli-short:"t" cli-usage:"managed Compute instances template NAME|ID"`
	TemplateFilter     string            `cli-usage:"managed Compute instances template filter"`
	Zone               string            `cli-short:"z" cli-usage:"Instance Pool zone"`
}

func (c *instancePoolCreateCmd) cmdAliases() []string { return gCreateAlias }

func (c *instancePoolCreateCmd) cmdShort() string { return "Create an Instance Pool" }

func (c *instancePoolCreateCmd) cmdLong() string {
	return fmt.Sprintf(`This command creates an Instance Pool.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&instancePoolShowOutput{}), ", "))
}

func (c *instancePoolCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	// TODO: remove this once the `--disk` flag is retired.
	if cmd.Flags().Changed("disk") {
		cmd.PrintErr(`**********************************************************************
WARNING: flag "--disk" has been deprecated and will be removed in a
future release, please use "--disk-size" instead.
**********************************************************************
`)
		if !cmd.Flags().Changed("disk-size") {
			diskFlag := cmd.Flags().Lookup("disk")
			if err := cmd.Flags().Set("disk-size", fmt.Sprint(diskFlag.Value.String())); err == nil {
				return err
			}
		}
	}

	// TODO: remove this once the `--keypair` flag is retired.
	if cmd.Flags().Changed("keypair") {
		cmd.PrintErr(`**********************************************************************
WARNING: flag "--keypair" has been deprecated and will be removed in
a future release, please use "--ssh-key" instead.
**********************************************************************
`)
		if !cmd.Flags().Changed("ssh-key") {
			keypairFlag := cmd.Flags().Lookup("keypair")
			if err := cmd.Flags().Set("ssh-key", keypairFlag.Value.String()); err != nil {
				return err
			}
		}
	}

	// TODO: remove this once the `--privnet` flag is retired.
	if cmd.Flags().Changed("privnet") {
		cmd.PrintErr(`**********************************************************************
WARNING: flag "--privnet" has been deprecated and will be removed in
a future release, please use "--private-network" instead.
**********************************************************************
`)
		if !cmd.Flags().Changed("private-network") {
			privnetFlag := cmd.Flags().Lookup("privnet")
			if err := cmd.Flags().Set(
				"private-network",
				strings.Trim(privnetFlag.Value.String(), "[]"),
			); err != nil {
				return err
			}
		}
	}

	// TODO: remove this once the `--service-offering` flag is retired.
	if cmd.Flags().Changed("service-offering") {
		cmd.PrintErr(`**********************************************************************
WARNING: flag "--service-offering" has been deprecated and will be removed
in a future release, please use "--instance-type" instead.
**********************************************************************
`)
		if !cmd.Flags().Changed("instance-type") {
			serviceOfferingFlag := cmd.Flags().Lookup("service-offering")
			if err := cmd.Flags().Set("instance-type", serviceOfferingFlag.Value.String()); err != nil {
				return err
			}
		}
	}

	cmdSetZoneFlagFromDefault(cmd)
	cmdSetTemplateFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instancePoolCreateCmd) cmdRun(_ *cobra.Command, _ []string) error {
	instancePool := &egoscale.InstancePool{
		Description:    nonEmptyStringPtr(c.Description),
		DiskSize:       &c.DiskSize,
		IPv6Enabled:    &c.IPv6,
		InstancePrefix: nonEmptyStringPtr(c.InstancePrefix),
		Labels: func() (v *map[string]string) {
			if len(c.Labels) > 0 {
				return &c.Labels
			}
			return
		}(),
		Name:   &c.Name,
		SSHKey: nonEmptyStringPtr(c.SSHKey),
		Size:   &c.Size,
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	zoneV1, err := getZoneByNameOrID(c.Zone)
	if err != nil {
		return err
	}

	if l := len(c.AntiAffinityGroups); l > 0 {
		antiAffinityGroupIDs := make([]string, l)
		for i := range c.AntiAffinityGroups {
			antiAffinityGroup, err := cs.FindAntiAffinityGroup(ctx, c.Zone, c.AntiAffinityGroups[i])
			if err != nil {
				return fmt.Errorf("error retrieving Anti-Affinity Group: %w", err)
			}
			antiAffinityGroupIDs[i] = *antiAffinityGroup.ID
		}
		instancePool.AntiAffinityGroupIDs = &antiAffinityGroupIDs
	}

	if c.DeployTarget != "" {
		deployTarget, err := cs.FindDeployTarget(ctx, c.Zone, c.DeployTarget)
		if err != nil {
			return fmt.Errorf("error retrieving Deploy Target: %w", err)
		}
		instancePool.DeployTargetID = deployTarget.ID
	}

	if l := len(c.ElasticIPs); l > 0 {
		elasticIPIDs := make([]string, l)
		for i := range c.ElasticIPs {
			elasticIP, err := cs.FindElasticIP(ctx, c.Zone, c.ElasticIPs[i])
			if err != nil {
				return fmt.Errorf("error retrieving Elastic IP: %w", err)
			}
			elasticIPIDs[i] = *elasticIP.ID
		}
		instancePool.ElasticIPIDs = &elasticIPIDs
	}

	instanceType, err := cs.FindInstanceType(ctx, c.Zone, c.InstanceType)
	if err != nil {
		return fmt.Errorf("error retrieving instance type: %w", err)
	}
	instancePool.InstanceTypeID = instanceType.ID

	if l := len(c.PrivateNetworks); l > 0 {
		privateNetworkIDs := make([]string, l)
		for i := range c.PrivateNetworks {
			privateNetwork, err := cs.FindPrivateNetwork(ctx, c.Zone, c.PrivateNetworks[i])
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
			securityGroup, err := cs.FindSecurityGroup(ctx, c.Zone, c.SecurityGroups[i])
			if err != nil {
				return fmt.Errorf("error retrieving Security Group: %w", err)
			}
			securityGroupIDs[i] = *securityGroup.ID
		}
		instancePool.SecurityGroupIDs = &securityGroupIDs
	}

	if instancePool.SSHKey == nil && gCurrentAccount.DefaultSSHKey != "" {
		instancePool.SSHKey = &gCurrentAccount.DefaultSSHKey
	}

	templateFilter, err := validateTemplateFilter(c.TemplateFilter)
	if err != nil {
		return err
	}

	template, err := getTemplateByNameOrID(zoneV1.ID, c.Template, templateFilter)
	if err != nil {
		return fmt.Errorf("error retrieving template: %w", err)
	}
	templateID := template.ID.String()
	instancePool.TemplateID = &templateID

	if c.CloudInitFile != "" {
		userData, err := getUserDataFromFile(c.CloudInitFile, c.CloudInitCompress)
		if err != nil {
			return fmt.Errorf("error parsing cloud-init user data: %w", err)
		}
		instancePool.UserData = &userData
	}

	decorateAsyncOperation(fmt.Sprintf("Creating Instance Pool %q...", c.Name), func() {
		instancePool, err = cs.CreateInstancePool(ctx, c.Zone, instancePool)
	})
	if err != nil {
		return err
	}

	if !gQuiet {
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

		DiskSize:       50,
		InstanceType:   fmt.Sprintf("%s.%s", defaultInstanceTypeFamily, defaultInstanceType),
		Size:           1,
		TemplateFilter: defaultTemplateFilter,
	}))

	// FIXME: remove this someday.
	cobra.CheckErr(registerCLICommand(deprecatedInstancePoolCmd, &instancePoolCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		DiskSize:       50,
		InstanceType:   fmt.Sprintf("%s.%s", defaultInstanceTypeFamily, defaultInstanceType),
		Size:           1,
		TemplateFilter: defaultTemplateFilter,
	}))
}
