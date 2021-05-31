package cmd

import (
	"fmt"
	"os"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

var instancePoolResetFields = []string{
	"anti-affinity-groups",
	"deploy-target",
	"description",
	"elastic-ips",
	"ipv6",
	"private-networks",
	"ssh-key",
	"security-groups",
	"user-data",
}

type instancePoolUpdateCmd struct {
	_ bool `cli-cmd:"update"`

	InstancePool string `cli-arg:"#" cli-usage:"NAME|ID"`

	AntiAffinityGroups []string `cli-flag:"anti-affinity-group" cli-short:"a" cli-usage:"managed Compute instances Anti-Affinity Group NAME|ID (can be specified multiple times)"`
	CloudInitFile      string   `cli-flag:"cloud-init" cli-short:"c" cli-usage:"cloud-init user data configuration file path"`
	DeployTarget       string   `cli-usage:"managed Compute instances Deploy Target NAME|ID"`
	Description        string   `cli-usage:"Instance Pool description"`
	DiskSize           int64    `cli-flag:"disk" cli-short:"d" cli-usage:"managed Compute instances disk size"`
	ElasticIPs         []string `cli-flag:"elastic-ip" cli-short:"e" cli-usage:"managed Compute instances Elastic IP ADDRESS|ID (can be specified multiple times)"`
	IPv6               bool     `cli-flag:"ipv6" cli-short:"6" cli-usage:"enable IPv6 on managed Compute instances"`
	InstancePrefix     string   `cli-usage:"string to prefix managed Compute instances names with"`
	InstanceType       string   `cli-flag:"service-offering" cli-short:"o" cli-usage:"managed Compute instances type"`
	Name               string   `cli-short:"n" cli-usage:"Instance Pool name"`
	PrivateNetworks    []string `cli-flag:"privnet" cli-short:"p" cli-usage:"managed Compute instances Private Network NAME|ID (can be specified multiple times)"`
	ResetFields        []string `cli-flag:"reset" cli-usage:"properties to reset to default value"`
	SSHKey             string   `cli-short:"k" cli-flag:"keypair" cli-usage:"SSH key to deploy on managed Compute instances"`
	SecurityGroups     []string `cli-flag:"security-group" cli-short:"s" cli-usage:"managed Compute instances Security Group NAME|ID (can be specified multiple times)"`
	Size               int64    `cli-usage:"Instance Pool size. Deprecated, replaced by the 'exo instancepool scale' command."`
	Template           string   `cli-short:"t" cli-usage:"managed Compute instances template NAME|ID"`
	TemplateFilter     string   `cli-usage:"managed Compute instances template filter"`
	Zone               string   `cli-short:"z" cli-usage:"Instance Pool zone"`
}

func (c *instancePoolUpdateCmd) cmdAliases() []string { return nil }

func (c *instancePoolUpdateCmd) cmdShort() string { return "Update an Instance Pool" }

func (c *instancePoolUpdateCmd) cmdLong() string {
	return fmt.Sprintf(`This command updates an Instance Pool.

Supported output template annotations: %s

Support values for --reset flag: %s`,
		strings.Join(outputterTemplateAnnotations(&instancePoolShowOutput{}), ", "),
		strings.Join(instancePoolResetFields, ", "),
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
		instancePool.AntiAffinityGroupIDs = make([]string, 0)
		for _, v := range c.AntiAffinityGroups {
			antiAffinityGroup, err := cs.FindAntiAffinityGroup(ctx, c.Zone, v)
			if err != nil {
				return fmt.Errorf("error retrieving Anti-Affinity Group: %s", err)
			}
			instancePool.AntiAffinityGroupIDs = append(instancePool.AntiAffinityGroupIDs, antiAffinityGroup.ID)
		}
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
		instancePool.Description = c.Description
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.DiskSize)) {
		instancePool.DiskSize = c.DiskSize
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.ElasticIPs)) {
		instancePool.ElasticIPIDs = make([]string, 0)
		for _, v := range c.ElasticIPs {
			elasticIP, err := cs.FindElasticIP(ctx, c.Zone, v)
			if err != nil {
				return fmt.Errorf("error retrieving Elastic IP: %s", err)
			}
			instancePool.ElasticIPIDs = append(instancePool.ElasticIPIDs, elasticIP.ID)
		}
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.InstancePrefix)) {
		instancePool.InstancePrefix = c.InstancePrefix
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.IPv6)) {
		instancePool.IPv6Enabled = c.IPv6
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Name)) {
		instancePool.Name = c.Name
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.PrivateNetworks)) {
		instancePool.PrivateNetworkIDs = make([]string, 0)
		for _, v := range c.PrivateNetworks {
			privateNetwork, err := cs.FindPrivateNetwork(ctx, c.Zone, v)
			if err != nil {
				return fmt.Errorf("error retrieving Private Network: %s", err)
			}
			instancePool.PrivateNetworkIDs = append(instancePool.PrivateNetworkIDs, privateNetwork.ID)
		}
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.SecurityGroups)) {
		instancePool.SecurityGroupIDs = make([]string, 0)
		for _, v := range c.SecurityGroups {
			securityGroup, err := cs.FindSecurityGroup(ctx, c.Zone, v)
			if err != nil {
				return fmt.Errorf("error retrieving Security Group: %s", err)
			}
			instancePool.SecurityGroupIDs = append(instancePool.SecurityGroupIDs, securityGroup.ID)
		}
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
		instancePool.SSHKey = c.SSHKey
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
		instancePool.TemplateID = template.ID.String()
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.CloudInitFile)) {
		if instancePool.UserData, err = getUserDataFromFile(c.CloudInitFile); err != nil {
			return fmt.Errorf("error parsing cloud-init user data: %s", err)
		}
		updated = true
	}

	decorateAsyncOperation(fmt.Sprintf("Updating Instance Pool %q...", instancePool.Name), func() {
		if updated {
			if err = cs.UpdateInstancePool(ctx, c.Zone, instancePool); err != nil {
				return
			}
		}

		for _, f := range c.ResetFields {
			switch f {
			case "anti-affinity-groups":
				err = instancePool.ResetField(ctx, &instancePool.AntiAffinityGroupIDs)
			case "elastic-ips":
				err = instancePool.ResetField(ctx, &instancePool.ElasticIPIDs)
			case "deploy-target":
				err = instancePool.ResetField(ctx, &instancePool.DeployTargetID)
			case "description":
				err = instancePool.ResetField(ctx, &instancePool.Description)
			case "ipv6":
				err = instancePool.ResetField(ctx, &instancePool.IPv6Enabled)
			case "private-networks":
				err = instancePool.ResetField(ctx, &instancePool.PrivateNetworkIDs)
			case "security-groups":
				err = instancePool.ResetField(ctx, &instancePool.SecurityGroupIDs)
			case "ssh-key":
				err = instancePool.ResetField(ctx, &instancePool.SSHKey)
			case "user-data":
				err = instancePool.ResetField(ctx, &instancePool.UserData)
			}
			if err != nil {
				return
			}
		}
	})
	if err != nil {
		return err
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Size)) {
		fmt.Fprintln(
			os.Stderr,
			`WARNING: the "--size" flag is deprecated and replaced by the `+
				`"exo instancepool scale" command, it will be removed in a future version.`,
		)

		decorateAsyncOperation(fmt.Sprintf("Scaling Instance Pool %q...", instancePool.Name), func() {
			err = instancePool.Scale(ctx, c.Size)
		})
	}

	if !gQuiet {
		return output(showInstancePool(c.Zone, instancePool.ID))
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(instancePoolCmd, &instancePoolUpdateCmd{
		TemplateFilter: defaultTemplateFilter,
	}))
}
