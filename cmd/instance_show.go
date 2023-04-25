package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instanceShowOutput struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	CreationDate       string            `json:"creation_date"`
	InstanceType       string            `json:"instance_type"`
	Template           string            `json:"template_id"`
	Zone               string            `json:"zoneid"`
	AntiAffinityGroups []string          `json:"anti_affinity_groups" outputLabel:"Anti-Affinity Groups"`
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

func (o *instanceShowOutput) Type() string { return "Compute instance" }
func (o *instanceShowOutput) ToJSON()      { output.JSON(o) }
func (o *instanceShowOutput) ToText()      { output.Text(o) }
func (o *instanceShowOutput) ToTable()     { output.Table(o) }

type instanceShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`

	ShowUserData bool   `cli-flag:"user-data" cli-short:"u" cli-usage:"show instance cloud-init user data configuration"`
	Zone         string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceShowCmd) cmdAliases() []string { return gShowAlias }

func (c *instanceShowCmd) cmdShort() string { return "Show a Compute instance details" }

func (c *instanceShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows a Compute instance details.

Supported output template annotations: %s`,
		strings.Join(output.OutputterTemplateAnnotations(&instanceShowOutput{}), ", "))
}

func (c *instanceShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceShowCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	instance, err := globalstate.GlobalEgoscaleClient.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	if c.ShowUserData {
		if instance.UserData != nil {
			userData, err := decodeUserData(*instance.UserData)
			if err != nil {
				return fmt.Errorf("error decoding user data: %w", err)
			}

			cmd.Print(userData)
		}

		return nil
	}

	out := instanceShowOutput{
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
			antiAffinityGroup, err := globalstate.GlobalEgoscaleClient.GetAntiAffinityGroup(ctx, c.Zone, id)
			if err != nil {
				return fmt.Errorf("error retrieving Anti-Affinity Group: %w", err)
			}
			out.AntiAffinityGroups = append(out.AntiAffinityGroups, *antiAffinityGroup.Name)
		}
	}

	if instance.ElasticIPIDs != nil {
		for _, id := range *instance.ElasticIPIDs {
			elasticIP, err := globalstate.GlobalEgoscaleClient.GetElasticIP(ctx, c.Zone, id)
			if err != nil {
				return fmt.Errorf("error retrieving Elastic IP: %w", err)
			}
			out.ElasticIPs = append(out.ElasticIPs, elasticIP.IPAddress.String())
		}
	}

	instanceType, err := globalstate.GlobalEgoscaleClient.GetInstanceType(ctx, c.Zone, *instance.InstanceTypeID)
	if err != nil {
		return err
	}
	out.InstanceType = fmt.Sprintf("%s.%s", *instanceType.Family, *instanceType.Size)

	if instance.PrivateNetworkIDs != nil {
		for _, id := range *instance.PrivateNetworkIDs {
			privateNetwork, err := globalstate.GlobalEgoscaleClient.GetPrivateNetwork(ctx, c.Zone, id)
			if err != nil {
				return fmt.Errorf("error retrieving Private Network: %w", err)
			}
			out.PrivateNetworks = append(out.PrivateNetworks, *privateNetwork.Name)
		}
	}

	if instance.SecurityGroupIDs != nil {
		for _, id := range *instance.SecurityGroupIDs {
			securityGroup, err := globalstate.GlobalEgoscaleClient.GetSecurityGroup(ctx, c.Zone, id)
			if err != nil {
				return fmt.Errorf("error retrieving Security Group: %w", err)
			}
			out.SecurityGroups = append(out.SecurityGroups, *securityGroup.Name)
		}
	}

	template, err := globalstate.GlobalEgoscaleClient.GetTemplate(ctx, c.Zone, *instance.TemplateID)
	if err != nil {
		return err
	}
	out.Template = *template.Name

	rdns, err := globalstate.GlobalEgoscaleClient.GetInstanceReverseDNS(ctx, c.Zone, *instance.ID)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			out.ReverseDNS = ""
		} else {
			return err
		}
	}

	out.ReverseDNS = rdns

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceCmd, &instanceShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
