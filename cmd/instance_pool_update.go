package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/userdata"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type instancePoolUpdateCmd struct {
	CliCommandSettings `cli-cmd:"-"`

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
	MinAvailable       int64             `cli-usage:"Minimum number of running Instances"`
	Name               string            `cli-short:"n" cli-usage:"Instance Pool name"`
	PrivateNetworks    []string          `cli-flag:"private-network" cli-usage:"managed Compute instances Private Network NAME|ID (can be specified multiple times)"`
	SSHKey             string            `cli-flag:"ssh-key" cli-usage:"SSH key to deploy on managed Compute instances"`
	SecurityGroups     []string          `cli-flag:"security-group" cli-short:"s" cli-usage:"managed Compute instances Security Group NAME|ID (can be specified multiple times)"`
	Template           string            `cli-short:"t" cli-usage:"managed Compute instances template NAME|ID"`
	TemplateVisibility string            `cli-usage:"instance template visibility (public|private)"`
	Zone               v3.ZoneName       `cli-short:"z" cli-usage:"Instance Pool zone"`
}

func (c *instancePoolUpdateCmd) CmdAliases() []string { return nil }

func (c *instancePoolUpdateCmd) CmdShort() string { return "Update an Instance Pool" }

func (c *instancePoolUpdateCmd) CmdLong() string {
	return fmt.Sprintf(`This command updates an Instance Pool.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instancePoolShowOutput{}), ", "),
	)
}

func (c *instancePoolUpdateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instancePoolUpdateCmd) CmdRun(cmd *cobra.Command, _ []string) error { //nolint:gocyclo
	var updated bool

	ctx := GContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	instancePools, err := client.ListInstancePools(ctx)
	if err != nil {
		return err
	}

	instancePool, err := instancePools.FindInstancePool(c.InstancePool)
	if err != nil {
		return err
	}
	updateReq := v3.UpdateInstancePoolRequest{}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.AntiAffinityGroups)) {
		updateReq.AntiAffinityGroups = make([]v3.AntiAffinityGroup, len(c.AntiAffinityGroups))
		af, err := client.ListAntiAffinityGroups(ctx)
		if err != nil {
			return fmt.Errorf("error listing Anti-Affinity Group: %w", err)
		}
		for i := range c.AntiAffinityGroups {
			antiAffinityGroup, err := af.FindAntiAffinityGroup(c.AntiAffinityGroups[i])
			if err != nil {
				return fmt.Errorf("error retrieving Anti-Affinity Group: %w", err)
			}
			updateReq.AntiAffinityGroups[i] = v3.AntiAffinityGroup{ID: antiAffinityGroup.ID}
		}
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.DeployTarget)) {
		targets, err := client.ListDeployTargets(ctx)
		if err != nil {
			return fmt.Errorf("error listing Deploy Target: %w", err)
		}
		deployTarget, err := targets.FindDeployTarget(c.DeployTarget)
		if err != nil {
			return fmt.Errorf("error retrieving Deploy Target: %w", err)
		}
		updateReq.DeployTarget = &v3.DeployTarget{ID: deployTarget.ID}
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.Description)) {
		updateReq.Description = c.Description
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.DiskSize)) {
		updateReq.DiskSize = c.DiskSize
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.ElasticIPs)) {
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
			updateReq.ElasticIPS = result
			updated = true
		}
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.InstancePrefix)) {
		updateReq.InstancePrefix = &c.InstancePrefix
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.IPv6)) {
		updateReq.Ipv6Enabled = &c.IPv6
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.Labels)) {
		updateReq.Labels = convertIfSpecialEmptyMap(c.Labels)
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.MinAvailable)) {
		updateReq.MinAvailable = &c.MinAvailable
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.Name)) {
		updateReq.Name = c.Name
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.PrivateNetworks)) {
		updateReq.PrivateNetworks = make([]v3.PrivateNetwork, len(c.PrivateNetworks))
		pn, err := client.ListPrivateNetworks(ctx)
		if err != nil {
			return fmt.Errorf("error listing Elastic IP: %w", err)
		}
		for i := range c.PrivateNetworks {
			privateNetwork, err := pn.FindPrivateNetwork(c.PrivateNetworks[i])
			if err != nil {
				return fmt.Errorf("error retrieving Private Network: %w", err)
			}
			updateReq.PrivateNetworks[i] = v3.PrivateNetwork{ID: privateNetwork.ID}
		}
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.SecurityGroups)) {
		sgs, err := client.ListSecurityGroups(ctx)

		if err != nil {
			return fmt.Errorf("error listing Security Group: %w", err)
		}
		updateReq.SecurityGroups = make([]v3.SecurityGroup, len(c.SecurityGroups))

		for i := range c.SecurityGroups {
			securityGroup, err := sgs.FindSecurityGroup(c.SecurityGroups[i])
			if err != nil {
				return fmt.Errorf("error retrieving Security Group: %w", err)
			}
			updateReq.SecurityGroups[i] = v3.SecurityGroup{ID: securityGroup.ID}
		}
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.InstanceType)) {
		instanceTypes, err := client.ListInstanceTypes(ctx)
		if err != nil {
			return fmt.Errorf("error listing instance type: %w", err)
		}

		instanceType := utils.ParseInstanceType(c.InstanceType)
		for i, it := range instanceTypes.InstanceTypes {
			if it.Family == instanceType.Family && it.Size == instanceType.Size {
				updateReq.InstanceType = &v3.InstanceType{ID: instanceTypes.InstanceTypes[i].ID}
				break
			}
		}
		if updateReq.InstanceType == nil {
			return fmt.Errorf("error retrieving instance type %s: not found", c.InstanceType)
		}

		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.SSHKey)) {
		updateReq.SSHKey = &v3.SSHKey{Name: c.SSHKey}
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.Template)) {
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
		updateReq.Template = &template
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.CloudInitFile)) {
		userData, err := userdata.GetUserDataFromFile(c.CloudInitFile, c.CloudInitCompress)
		if err != nil {
			return fmt.Errorf("error parsing cloud-init user data: %w", err)
		}
		updateReq.UserData = &userData
		updated = true
	}

	if updated {
		decorateAsyncOperation(fmt.Sprintf("Updating Instance Pool %q...", c.InstancePool), func() {
			_, updateErr := client.UpdateInstancePool(ctx, instancePool.ID, updateReq)
			err = updateErr
		})
		if err != nil {
			return err
		}
	}

	if !globalstate.Quiet {
		return (&instancePoolShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			InstancePool:       instancePool.ID.String(),
			Zone:               string(c.Zone),
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(RegisterCLICommand(instancePoolCmd, &instancePoolUpdateCmd{
		CliCommandSettings: DefaultCLICmdSettings(),

		TemplateVisibility: defaultTemplateVisibility,
	}))
}
