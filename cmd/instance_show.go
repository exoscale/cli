package cmd

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/userdata"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceShowOutput struct {
	ID                 v3.UUID           `json:"id"`
	Name               string            `json:"name"`
	CreationDate       string            `json:"creation_date"`
	InstanceType       string            `json:"instance_type"`
	Template           string            `json:"template"`
	Zone               v3.ZoneName       `json:"zone"`
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
	State              v3.InstanceState  `json:"state"`
	Labels             map[string]string `json:"labels"`
	SecureBoot         bool              `json:"secureboot"`
	Tpm                bool              `json:"tpm"`
	ReverseDNS         v3.DomainName     `json:"reverse_dns" outputLabel:"Reverse DNS"`
}

func (o *instanceShowOutput) Type() string { return "Compute instance" }
func (o *instanceShowOutput) ToJSON()      { output.JSON(o) }
func (o *instanceShowOutput) ToText()      { output.Text(o) }
func (o *instanceShowOutput) ToTable()     { output.Table(o) }

type instanceShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`

	ShowUserData bool        `cli-flag:"user-data" cli-short:"u" cli-usage:"show instance cloud-init user data configuration"`
	Zone         v3.ZoneName `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceShowCmd) cmdAliases() []string { return gShowAlias }

func (c *instanceShowCmd) cmdShort() string { return "Show a Compute instance details" }

func (c *instanceShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows a Compute instance details.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instanceShowOutput{}), ", "))
}

func (c *instanceShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceShowCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	resp, err := client.ListInstances(ctx)
	if err != nil {
		return err
	}

	found_instance, err := resp.FindListInstancesResponseInstances(c.Instance)
	if err != nil {
		return err
	}

	instance, err := client.GetInstance(ctx, found_instance.ID)
	if err != nil {
		return err
	}

	if c.ShowUserData {
		if instance.UserData != "" {
			userData, err := userdata.DecodeUserData(instance.UserData)
			if err != nil {
				return fmt.Errorf("error decoding user data: %w", err)
			}

			cmd.Print(userData)
		}

		return nil
	}
	var ipV6 *net.IP
	if parsed := net.ParseIP(instance.Ipv6Address); parsed != nil {
		ipV6 = &parsed // only assign pointer if it's a valid IP
	}

	var sshKeyName *string
	if instance.SSHKey != nil {
		sshKeyName = &instance.SSHKey.Name
	}

	out := instanceShowOutput{
		AntiAffinityGroups: make([]string, 0),
		CreationDate:       instance.CreatedAT.String(),
		DiskSize:           humanize.IBytes(uint64(instance.DiskSize << 30)),
		ElasticIPs:         make([]string, 0),
		ID:                 instance.ID,
		IPAddress:          utils.DefaultIP(&instance.PublicIP, "-"),
		IPv6Address:        utils.DefaultIP(ipV6, "-"),
		Labels: func() (v map[string]string) {
			if instance.Labels != nil {
				v = instance.Labels
			}
			return
		}(),
		Name:            instance.Name,
		PrivateNetworks: make([]string, 0),
		SSHKey:          utils.DefaultString(sshKeyName, "-"),
		SecurityGroups:  make([]string, 0),
		SecureBoot:      *instance.SecurebootEnabled,
		Tpm:             *instance.TpmEnabled,
		State:           instance.State,
		Zone:            c.Zone,
	}

	out.PrivateInstance = "No"
	if instance.PublicIPAssignment != "" && instance.PublicIPAssignment == "none" {
		out.PrivateInstance = "Yes"
	}

	if instance.AntiAffinityGroups != nil {
		for _, group := range instance.AntiAffinityGroups {
			resp, err := client.ListAntiAffinityGroups(ctx)
			if err != nil {
				return err
			}
			found_group, err := resp.FindAntiAffinityGroup(group.ID.String())
			if err != nil {
				return fmt.Errorf("error retrieving Anti-Affinity Group: %w", err)
			}
			out.AntiAffinityGroups = append(out.AntiAffinityGroups, found_group.Name)
		}
	}

	out.DeployTarget = "-"
	if instance.DeployTarget != nil {
		resp, err := client.ListDeployTargets(ctx)
		if err != nil {
			return err
		}
		dt, err := resp.FindDeployTarget(instance.DeployTarget.ID.String())
		if err != nil {
			return fmt.Errorf("error retrieving Deploy Target: %w", err)
		}
		out.DeployTarget = dt.Name
	}

	if instance.ElasticIPS != nil {
		for _, eip := range instance.ElasticIPS {
			resp, err := client.ListElasticIPS(ctx)
			if err != nil {
				return err
			}
			found_eip, err := resp.FindElasticIP(eip.ID.String())
			if err != nil {
				return fmt.Errorf("error retrieving Elastic IP: %w", err)
			}
			out.ElasticIPs = append(out.ElasticIPs, found_eip.IP)
		}
	}

	it, err := client.GetInstanceType(ctx, instance.InstanceType.ID)
	if err != nil {
		return err
	}
	out.InstanceType = fmt.Sprintf("%s.%s", it.Family, it.Size)

	if instance.PrivateNetworks != nil {
		for _, pn := range instance.PrivateNetworks {
			resp, err := client.ListPrivateNetworks(ctx)
			if err != nil {
				return err
			}
			found_pn, err := resp.FindPrivateNetwork(pn.ID.String())
			if err != nil {
				return fmt.Errorf("error retrieving Private Network: %w", err)
			}
			out.PrivateNetworks = append(out.PrivateNetworks, found_pn.Name)
		}
	}

	if instance.SecurityGroups != nil {
		for _, sg := range instance.SecurityGroups {
			resp, err := client.ListSecurityGroups(ctx)
			if err != nil {
				return err
			}
			found_sg, err := resp.FindSecurityGroup(sg.ID.String())
			if err != nil {
				return fmt.Errorf("error retrieving Security Group: %w", err)
			}
			out.SecurityGroups = append(out.SecurityGroups, found_sg.Name)
		}
	}

	template, err := client.GetTemplate(ctx, instance.Template.ID)
	if err != nil {
		return err
	}
	out.Template = template.Name

	rdns, err := client.GetReverseDNSInstance(ctx, instance.ID)
	if err != nil {
		if errors.Is(err, v3.ErrNotFound) {
			out.ReverseDNS = ""
		} else {
			return err
		}
	} else if rdns == nil {
		out.ReverseDNS = ""
	} else {
		out.ReverseDNS = rdns.DomainName
	}

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceCmd, &instanceShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
