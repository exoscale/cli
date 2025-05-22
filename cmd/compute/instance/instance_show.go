package instance

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

type InstanceShowOutput struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	CreationDate       string            `json:"creation_date"`
	InstanceType       string            `json:"instance_type"`
	Template           string            `json:"template"`
	Zone               string            `json:"zone"`
	AntiAffinityGroups []string          `json:"anti_affinity_groups" outputLabel:"Anti-Affinity Groups"`
	DeployTarget       string            `json:"deploy_target"`
	SecurityGroups     []string          `json:"security_groups"`
	PrivateInstance    string            `json:"private-instance" outputLabel:"Private Instance"`
	PrivateNetworks    []string          `json:"private_networks"`
	ElasticIPs         []string          `json:"elastic_ips" outputLabel:"Elastic IPs"`
	IPAddress          string            `json:"ip_address"`
	IPv6Address        string            `json:"ipv6_address" outputLabel:"IPv6 Address"`
	SSHKey             string            `json:"ssh_key"`
	DiskSize           string            `json:"disk_size"`
	State              string            `json:"state"`
	Labels             map[string]string `json:"labels"`
	ReverseDNS         string            `json:"reverse_dns" outputLabel:"Reverse DNS"`
}

func (o *InstanceShowOutput) Type() string { return "Compute instance" }
func (o *InstanceShowOutput) ToJSON()      { output.JSON(o) }
func (o *InstanceShowOutput) ToText()      { output.Text(o) }
func (o *InstanceShowOutput) ToTable()     { output.Table(o) }

type instanceShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`

	ShowUserData bool   `cli-flag:"user-data" cli-short:"u" cli-usage:"show instance cloud-init user data configuration"`
	Zone         string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceShowCmd) CmdAliases() []string { return exocmd.GShowAlias }

func (c *instanceShowCmd) CmdShort() string { return "Show a Compute instance details" }

func (c *instanceShowCmd) CmdLong() string {
	return fmt.Sprintf(`This command shows a Compute instance details.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&InstanceShowOutput{}), ", "))
}

func (c *instanceShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceShowCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	instance, err := globalstate.EgoscaleClient.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	if c.ShowUserData {
		if instance.UserData != nil {
			userData, err := userdata.DecodeUserData(*instance.UserData)
			if err != nil {
				return fmt.Errorf("error decoding user data: %w", err)
			}

			cmd.Print(userData)
		}

		return nil
	}

	out := InstanceShowOutput{
		AntiAffinityGroups: make([]string, 0),
		CreationDate:       instance.CreatedAt.String(),
		DiskSize:           humanize.IBytes(uint64(*instance.DiskSize << 30)),
		ElasticIPs:         make([]string, 0),
		ID:                 *instance.ID,
		IPAddress:          utils.DefaultIP(instance.PublicIPAddress, "-"),
		IPv6Address:        utils.DefaultIP(instance.IPv6Address, "-"),
		Labels: func() (v map[string]string) {
			if instance.Labels != nil {
				v = *instance.Labels
			}
			return
		}(),
		Name:            *instance.Name,
		PrivateNetworks: make([]string, 0),
		SSHKey:          utils.DefaultString(instance.SSHKey, "-"),
		SecurityGroups:  make([]string, 0),
		State:           *instance.State,
		Zone:            c.Zone,
	}

	out.PrivateInstance = "No"
	if instance.PublicIPAssignment != nil && *instance.PublicIPAssignment == "none" {
		out.PrivateInstance = "Yes"
	}

	if instance.AntiAffinityGroupIDs != nil {
		for _, id := range *instance.AntiAffinityGroupIDs {
			antiAffinityGroup, err := globalstate.EgoscaleClient.GetAntiAffinityGroup(ctx, c.Zone, id)
			if err != nil {
				return fmt.Errorf("error retrieving Anti-Affinity Group: %w", err)
			}
			out.AntiAffinityGroups = append(out.AntiAffinityGroups, *antiAffinityGroup.Name)
		}
	}

	out.DeployTarget = "-"
	if instance.DeployTargetID != nil {
		DeployTarget, err := globalstate.EgoscaleClient.GetDeployTarget(ctx, c.Zone, *instance.DeployTargetID)
		if err != nil {
			return fmt.Errorf("error retrieving Deploy Target: %w", err)
		}
		out.DeployTarget = *DeployTarget.Name
	}

	if instance.ElasticIPIDs != nil {
		for _, id := range *instance.ElasticIPIDs {
			elasticIP, err := globalstate.EgoscaleClient.GetElasticIP(ctx, c.Zone, id)
			if err != nil {
				return fmt.Errorf("error retrieving Elastic IP: %w", err)
			}
			out.ElasticIPs = append(out.ElasticIPs, elasticIP.IPAddress.String())
		}
	}

	instanceType, err := globalstate.EgoscaleClient.GetInstanceType(ctx, c.Zone, *instance.InstanceTypeID)
	if err != nil {
		return err
	}
	out.InstanceType = fmt.Sprintf("%s.%s", *instanceType.Family, *instanceType.Size)

	if instance.PrivateNetworkIDs != nil {
		for _, id := range *instance.PrivateNetworkIDs {
			privateNetwork, err := globalstate.EgoscaleClient.GetPrivateNetwork(ctx, c.Zone, id)
			if err != nil {
				return fmt.Errorf("error retrieving Private Network: %w", err)
			}
			out.PrivateNetworks = append(out.PrivateNetworks, *privateNetwork.Name)
		}
	}

	if instance.SecurityGroupIDs != nil {
		for _, id := range *instance.SecurityGroupIDs {
			securityGroup, err := globalstate.EgoscaleClient.GetSecurityGroup(ctx, c.Zone, id)
			if err != nil {
				return fmt.Errorf("error retrieving Security Group: %w", err)
			}
			out.SecurityGroups = append(out.SecurityGroups, *securityGroup.Name)
		}
	}

	template, err := globalstate.EgoscaleClient.GetTemplate(ctx, c.Zone, *instance.TemplateID)
	if err != nil {
		return err
	}
	out.Template = *template.Name

	rdns, err := globalstate.EgoscaleClient.GetInstanceReverseDNS(ctx, c.Zone, *instance.ID)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			out.ReverseDNS = ""
		} else {
			return err
		}
	}

	out.ReverseDNS = rdns

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceCmd, &instanceShowCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
