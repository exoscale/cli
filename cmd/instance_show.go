package cmd

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instanceShowOutput struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	CreationDate       string            `json:"created_at"`
	InstanceType       string            `json:"instance_type"`
	Template           string            `json:"template_id"`
	Zone               string            `json:"zoneid"`
	AntiAffinityGroups []string          `json:"anti_affinity_groups" outputLabel:"Anti-Affinity Groups"`
	SecurityGroups     []string          `json:"security_groups"`
	PrivateNetworks    []string          `json:"private_networks"`
	ElasticIPs         []string          `json:"elastic_ips" outputLabel:"Elastic IPs"`
	IPAddress          string            `json:"ip_address"`
	IPv6Address        string            `json:"ipv6_address" outputLabel:"IPv6 Address"`
	SSHKey             string            `json:"ssh_key"`
	DiskSize           string            `json:"disk_size"`
	State              string            `json:"state"`
	Labels             map[string]string `json:"labels"`
}

func (o *instanceShowOutput) toJSON()  { outputJSON(o) }
func (o *instanceShowOutput) toText()  { outputText(o) }
func (o *instanceShowOutput) toTable() { outputTable(o) }

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
		strings.Join(outputterTemplateAnnotations(&instanceShowOutput{}), ", "))
}

func (c *instanceShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceShowCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	if c.ShowUserData {
		ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

		instance, err := cs.FindInstance(ctx, c.Zone, c.Instance)
		if err != nil {
			return err
		}

		if instance.UserData != nil {
			userData, err := decodeUserData(*instance.UserData)
			if err != nil {
				return fmt.Errorf("error decoding user data: %s", err)
			}

			cmd.Print(userData)
		}

		return nil
	}

	return output(showInstance(c.Zone, c.Instance))
}

func showInstance(zone, i string) (outputter, error) {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

	instance, err := cs.FindInstance(ctx, zone, i)
	if err != nil {
		return nil, err
	}

	out := instanceShowOutput{
		AntiAffinityGroups: make([]string, 0),
		CreationDate:       instance.CreatedAt.String(),
		DiskSize:           humanize.IBytes(uint64(*instance.DiskSize << 30)),
		ElasticIPs:         make([]string, 0),
		ID:                 *instance.ID,
		IPAddress:          instance.PublicIPAddress.String(),
		IPv6Address: func() string {
			if instance.IPv6Address != nil {
				return instance.IPv6Address.String()
			}
			return "-"
		}(),
		Labels: func() (v map[string]string) {
			if instance.Labels != nil {
				v = *instance.Labels
			}
			return
		}(),
		Name:            *instance.Name,
		PrivateNetworks: make([]string, 0),
		SSHKey:          defaultString(instance.SSHKey, "-"),
		SecurityGroups:  make([]string, 0),
		State:           *instance.State,
		Zone:            zone,
	}

	antiAffinityGroups, err := instance.AntiAffinityGroups(ctx)
	if err != nil {
		return nil, err
	}
	for _, antiAffinityGroup := range antiAffinityGroups {
		out.AntiAffinityGroups = append(out.AntiAffinityGroups, *antiAffinityGroup.Name)
	}

	elasticIPs, err := instance.ElasticIPs(ctx)
	if err != nil {
		return nil, err
	}
	for _, elasticIP := range elasticIPs {
		out.ElasticIPs = append(out.ElasticIPs, elasticIP.IPAddress.String())
	}

	instanceType, err := cs.GetInstanceType(ctx, zone, *instance.InstanceTypeID)
	if err != nil {
		return nil, err
	}
	out.InstanceType = fmt.Sprintf("%s.%s", *instanceType.Family, *instanceType.Size)

	privateNetworks, err := instance.PrivateNetworks(ctx)
	if err != nil {
		return nil, err
	}
	for _, privateNetwork := range privateNetworks {
		out.PrivateNetworks = append(out.PrivateNetworks, *privateNetwork.Name)
	}

	securityGroups, err := instance.SecurityGroups(ctx)
	if err != nil {
		return nil, err
	}
	for _, securityGroup := range securityGroups {
		out.SecurityGroups = append(out.SecurityGroups, *securityGroup.Name)
	}

	template, err := cs.GetTemplate(ctx, zone, *instance.TemplateID)
	if err != nil {
		return nil, err
	}
	out.Template = *template.Name

	return &out, nil
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceCmd, &instanceShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
