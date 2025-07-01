package instance_pool

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/userdata"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type instancePoolShowOutput struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	Description        string            `json:"description"`
	InstanceType       string            `json:"instance_type"`
	Template           string            `json:"template"`
	Zone               string            `json:"zone"`
	AntiAffinityGroups []string          `json:"anti_affinity_groups" outputLabel:"Anti-Affinity Groups"`
	SecurityGroups     []string          `json:"security_groups"`
	PrivateNetworks    []string          `json:"private_networks"`
	ElasticIPs         []string          `json:"elastic_ips" outputLabel:"Elastic IPs"`
	IPv6               bool              `json:"ipv6" outputLabel:"IPv6"`
	SSHKey             string            `json:"ssh_key"`
	Size               int64             `json:"size"`
	DiskSize           string            `json:"disk_size"`
	InstancePrefix     string            `json:"instance_prefix"`
	State              string            `json:"state"`
	Labels             map[string]string `json:"labels"`
	Instances          []string          `json:"instances"`
}

func (o *instancePoolShowOutput) Type() string { return "Instance Pool" }
func (o *instancePoolShowOutput) ToJSON()      { output.JSON(o) }
func (o *instancePoolShowOutput) ToText()      { output.Text(o) }
func (o *instancePoolShowOutput) ToTable()     { output.Table(o) }

type instancePoolShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	InstancePool string `cli-arg:"#" cli-usage:"NAME|ID"`

	ShowUserData bool   `cli-flag:"user-data" cli-short:"u" cli-usage:"show cloud-init user data configuration"`
	Zone         string `cli-short:"z" cli-usage:"Instance Pool zone"`
}

func (c *instancePoolShowCmd) CmdAliases() []string { return exocmd.GShowAlias }

func (c *instancePoolShowCmd) CmdShort() string { return "Show an Instance Pool details" }

func (c *instancePoolShowCmd) CmdLong() string {
	return fmt.Sprintf(`This command shows an Instance Pool details.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instancePoolShowOutput{}), ", "))
}

func (c *instancePoolShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instancePoolShowCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
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

	if c.ShowUserData {
		if instancePool.UserData != "" {
			userData, err := userdata.DecodeUserData(instancePool.UserData)
			if err != nil {
				return fmt.Errorf("error decoding user data: %w", err)
			}

			cmd.Print(userData)
		}

		return nil
	}

	out := instancePoolShowOutput{
		AntiAffinityGroups: make([]string, 0),
		Description:        instancePool.Description,
		DiskSize:           humanize.IBytes(uint64(instancePool.DiskSize << 30)),
		ElasticIPs:         make([]string, 0),
		ID:                 instancePool.ID.String(),
		IPv6:               utils.DefaultBool(instancePool.Ipv6Enabled, false),
		InstancePrefix:     instancePool.InstancePrefix,
		Instances:          make([]string, 0),
		Labels: func() (v map[string]string) {
			if instancePool.Labels != nil {
				v = instancePool.Labels
			}
			return
		}(),
		Name:            instancePool.Name,
		PrivateNetworks: make([]string, 0),
		SSHKey: func() string {
			if instancePool.SSHKey != nil {
				return instancePool.SSHKey.Name
			}
			return "-"
		}(),
		SecurityGroups: make([]string, 0),
		Size:           instancePool.Size,
		State:          string(instancePool.State),
		Zone:           c.Zone,
	}

	if instancePool.AntiAffinityGroups != nil {
		for _, aag := range instancePool.AntiAffinityGroups {
			aag, err := client.GetAntiAffinityGroup(ctx, aag.ID)
			if err != nil {
				return err
			}
			out.AntiAffinityGroups = append(out.AntiAffinityGroups, aag.Name)
		}
	}

	if instancePool.ElasticIPS != nil {
		for _, ip := range instancePool.ElasticIPS {
			ip, err := client.GetElasticIP(ctx, ip.ID)
			if err != nil {
				return err
			}
			out.ElasticIPs = append(out.ElasticIPs, ip.IP)
		}
	}

	if instancePool.Instances != nil {
		for _, instance := range instancePool.Instances {
			instance, err := client.GetInstance(ctx, instance.ID)
			if err != nil {
				return err
			}
			out.Instances = append(out.Instances, instance.Name)
		}
	}

	instanceType, err := client.GetInstanceType(ctx, instancePool.InstanceType.ID)
	if err != nil {
		return err
	}
	out.InstanceType = fmt.Sprintf("%s.%s", instanceType.Family, instanceType.Size)

	if instancePool.PrivateNetworks != nil {
		for _, privateNetwork := range instancePool.PrivateNetworks {
			privateNetwork, err := client.GetPrivateNetwork(ctx, privateNetwork.ID)
			if err != nil {
				return err
			}
			out.PrivateNetworks = append(out.PrivateNetworks, privateNetwork.Name)
		}
	}

	if instancePool.SecurityGroups != nil {
		for _, securityGroup := range instancePool.SecurityGroups {
			securityGroup, err := client.GetSecurityGroup(ctx, securityGroup.ID)
			if err != nil {
				return err
			}

			out.SecurityGroups = append(out.SecurityGroups, securityGroup.Name)
		}
	}

	template, err := client.GetTemplate(ctx, instancePool.Template.ID)
	if err != nil {
		return fmt.Errorf("error retrieving template: %w", err)
	}
	out.Template = template.Name

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instancePoolCmd, &instancePoolShowCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
