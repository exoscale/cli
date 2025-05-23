package instance_pool

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/userdata"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
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
	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	instancePool, err := globalstate.EgoscaleClient.FindInstancePool(ctx, c.Zone, c.InstancePool)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}

		return err
	}

	if c.ShowUserData {
		if instancePool.UserData != nil {
			userData, err := userdata.DecodeUserData(*instancePool.UserData)
			if err != nil {
				return fmt.Errorf("error decoding user data: %w", err)
			}

			cmd.Print(userData)
		}

		return nil
	}

	out := instancePoolShowOutput{
		AntiAffinityGroups: make([]string, 0),
		Description:        utils.DefaultString(instancePool.Description, ""),
		DiskSize:           humanize.IBytes(uint64(*instancePool.DiskSize << 30)),
		ElasticIPs:         make([]string, 0),
		ID:                 *instancePool.ID,
		IPv6:               utils.DefaultBool(instancePool.IPv6Enabled, false),
		InstancePrefix:     utils.DefaultString(instancePool.InstancePrefix, ""),
		Instances:          make([]string, 0),
		Labels: func() (v map[string]string) {
			if instancePool.Labels != nil {
				v = *instancePool.Labels
			}
			return
		}(),
		Name:            *instancePool.Name,
		PrivateNetworks: make([]string, 0),
		SSHKey:          utils.DefaultString(instancePool.SSHKey, "-"),
		SecurityGroups:  make([]string, 0),
		Size:            *instancePool.Size,
		State:           *instancePool.State,
		Zone:            c.Zone,
	}

	if instancePool.AntiAffinityGroupIDs != nil {
		for _, id := range *instancePool.AntiAffinityGroupIDs {
			antiAffinityGroup, err := globalstate.EgoscaleClient.GetAntiAffinityGroup(ctx, c.Zone, id)
			if err != nil {
				return fmt.Errorf("error retrieving Anti-Affinity Group: %w", err)
			}
			out.AntiAffinityGroups = append(out.AntiAffinityGroups, *antiAffinityGroup.Name)
		}
	}

	if instancePool.ElasticIPIDs != nil {
		for _, id := range *instancePool.ElasticIPIDs {
			elasticIP, err := globalstate.EgoscaleClient.GetElasticIP(ctx, c.Zone, id)
			if err != nil {
				return fmt.Errorf("error retrieving Elastic IP: %w", err)
			}
			out.ElasticIPs = append(out.ElasticIPs, elasticIP.IPAddress.String())
		}
	}

	if instancePool.InstanceIDs != nil {
		for _, id := range *instancePool.InstanceIDs {
			instance, err := globalstate.EgoscaleClient.GetInstance(ctx, c.Zone, id)
			if err != nil {
				return fmt.Errorf("error retrieving Compute instance: %w", err)
			}
			out.Instances = append(out.Instances, *instance.Name)
		}
	}

	instanceType, err := globalstate.EgoscaleClient.GetInstanceType(ctx, c.Zone, *instancePool.InstanceTypeID)
	if err != nil {
		return err
	}
	out.InstanceType = fmt.Sprintf("%s.%s", *instanceType.Family, *instanceType.Size)

	if instancePool.PrivateNetworkIDs != nil {
		for _, id := range *instancePool.PrivateNetworkIDs {
			privateNetwork, err := globalstate.EgoscaleClient.GetPrivateNetwork(ctx, c.Zone, id)
			if err != nil {
				return fmt.Errorf("error retrieving Private Network: %w", err)
			}
			out.PrivateNetworks = append(out.PrivateNetworks, *privateNetwork.Name)
		}
	}

	if instancePool.SecurityGroupIDs != nil {
		for _, id := range *instancePool.SecurityGroupIDs {
			securityGroup, err := globalstate.EgoscaleClient.GetSecurityGroup(ctx, c.Zone, id)
			if err != nil {
				return fmt.Errorf("error retrieving Security Group: %w", err)
			}
			out.SecurityGroups = append(out.SecurityGroups, *securityGroup.Name)
		}
	}

	template, err := globalstate.EgoscaleClient.GetTemplate(ctx, c.Zone, *instancePool.TemplateID)
	if err != nil {
		return fmt.Errorf("error retrieving template: %w", err)
	}
	out.Template = *template.Name

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instancePoolCmd, &instancePoolShowCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
