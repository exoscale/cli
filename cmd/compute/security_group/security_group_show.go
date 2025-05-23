package security_group

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/cmd/compute/instance"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type securityGroupRuleOutput struct {
	ID            v3.UUID `json:"id"`
	Description   string  `json:"description"`
	ICMPCode      int64   `json:"icmp_code,omitempty"`
	ICMPType      int64   `json:"icmp_type,omitempty"`
	Network       string  `json:"network,omitempty"`
	Protocol      string  `json:"protocol"`
	SecurityGroup string  `json:"security_group,omitempty"`
	StartPort     uint16  `json:"start_port,omitempty"`
	EndPort       uint16  `json:"end_port,omitempty"`
}

type securityGroupShowOutput struct {
	ID              v3.UUID                       `json:"id"`
	Name            string                        `json:"name"`
	Description     string                        `json:"description"`
	ExternalSources []string                      `json:"external_sources"`
	IngressRules    []securityGroupRuleOutput     `json:"ingress_rules"`
	EgressRules     []securityGroupRuleOutput     `json:"egress_rules"`
	Instances       []securityGroupInstanceOutput `json:"instances"`
}

type securityGroupInstanceOutput struct {
	Name     string      `json:"name"`
	PublicIP string      `json:"public_ip"`
	ID       string      `json:"id"`
	Zone     v3.ZoneName `json:"zone"`
}

func (o *securityGroupShowOutput) ToJSON() { output.JSON(o) }
func (o *securityGroupShowOutput) ToText() { output.Text(o) }
func (o *securityGroupShowOutput) ToTable() {
	formatExternalSources := func(sources []string) string {
		if len(sources) > 0 {
			return strings.Join(sources, ", ")
		}
		return "-"
	}

	formatRule := func(rules []securityGroupRuleOutput) string {
		if len(rules) > 0 {
			buf := bytes.NewBuffer(nil)
			at := table.NewEmbeddedTable(buf)
			at.SetHeader([]string{" "})
			at.SetAlignment(tablewriter.ALIGN_LEFT)

			for _, rule := range rules {
				r := []string{rule.ID.String(), rule.Description, strings.ToUpper(rule.Protocol)}

				if rule.Network != "" {
					r = append(r, rule.Network)
				} else {
					r = append(r, rule.SecurityGroup)
				}

				if strings.HasPrefix(rule.Protocol, "icmp") {
					r = append(r, fmt.Sprintf("ICMP code:%d type:%d", rule.ICMPCode, rule.ICMPType))
				} else if rule.StartPort != 0 {
					r = append(r, func() string {
						if rule.StartPort == rule.EndPort {
							return fmt.Sprint(rule.StartPort)
						}
						return fmt.Sprintf("%d-%d", rule.StartPort, rule.EndPort)
					}())
				}

				at.Append(r)
			}
			at.Render()

			return buf.String()
		}

		return "-"
	}

	formatInstances := func(instances []securityGroupInstanceOutput) string {
		if len(instances) < 1 {
			return "-"
		}

		buf := bytes.NewBuffer(nil)
		at := table.NewEmbeddedTable(buf)
		at.SetHeader([]string{" "})
		at.SetAlignment(tablewriter.ALIGN_LEFT)

		for _, instance := range instances {
			r := []string{instance.Name, instance.ID}

			r = append(r, instance.PublicIP)
			r = append(r, string(instance.Zone))

			at.Append(r)
		}

		at.Render()

		return buf.String()
	}

	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Security Group"})
	defer t.Render()

	t.Append([]string{"ID", o.ID.String()})
	t.Append([]string{"Name", o.Name})
	t.Append([]string{"Description", o.Description})
	t.Append([]string{"Ingress Rules", formatRule(o.IngressRules)})
	t.Append([]string{"Egress Rules", formatRule(o.EgressRules)})
	t.Append([]string{"External Sources", formatExternalSources(o.ExternalSources)})
	t.Append([]string{"Instances", formatInstances(o.Instances)})
}

type securityGroupShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	SecurityGroup string `cli-arg:"#" cli-usage:"NAME|ID"`
}

func (c *securityGroupShowCmd) CmdAliases() []string { return exocmd.GShowAlias }

func (c *securityGroupShowCmd) CmdShort() string {
	return "Show a Security Group details"
}

func (c *securityGroupShowCmd) CmdLong() string {
	return fmt.Sprintf(`This command shows a Compute instance Security Group details.

Supported output template annotations for Security Group: %s

Supported output template annotations for Security Group rules: %s`,
		strings.Join(output.TemplateAnnotations(&securityGroupShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&securityGroupRuleOutput{}), ", "))
}

func (c *securityGroupShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *securityGroupShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}
	securityGroups, err := client.ListSecurityGroups(ctx)
	if err != nil {
		return err
	}
	securityGroup, err := securityGroups.FindSecurityGroup(c.SecurityGroup)

	if err != nil {
		return err
	}

	externalSources := make([]string, 0)
	if securityGroup.ExternalSources != nil {
		externalSources = securityGroup.ExternalSources
	}
	out := securityGroupShowOutput{
		ID:              securityGroup.ID,
		Name:            securityGroup.Name,
		Description:     securityGroup.Description,
		ExternalSources: externalSources,
		IngressRules:    make([]securityGroupRuleOutput, 0),
		EgressRules:     make([]securityGroupRuleOutput, 0),
	}

	for _, rule := range securityGroup.Rules {
		or := securityGroupRuleOutput{
			ID:          rule.ID,
			Description: rule.Description,
			Network:     rule.Network,
			Protocol:    string(rule.Protocol),
			StartPort:   uint16(rule.StartPort),
			EndPort:     uint16(rule.EndPort),
		}

		if rule.ICMP != nil {
			or.ICMPCode = rule.ICMP.Code
			or.ICMPCode = rule.ICMP.Code
		}

		if rule.SecurityGroup != nil {
			if rule.SecurityGroup.ID != "" {
				ruleSecurityGroup, err := client.GetSecurityGroup(ctx, rule.SecurityGroup.ID)
				if err != nil {
					return fmt.Errorf("error retrieving Security Group: %w", err)
				}
				ruleSG := "SG:" + ruleSecurityGroup.Name
				or.SecurityGroup = ruleSG
			}
			if rule.SecurityGroup.Name != "" {
				ruleSG := "PUBLIC-SG:" + rule.SecurityGroup.Name
				or.SecurityGroup = ruleSG
			}

		}

		if rule.FlowDirection == "ingress" {
			out.IngressRules = append(out.IngressRules, or)
		} else {
			out.EgressRules = append(out.EgressRules, or)
		}
	}

	instancesByZone, err := utils.GetInstancesInSecurityGroup(ctx, globalstate.EgoscaleV3Client, securityGroup.ID)
	if err != nil {
		return fmt.Errorf("error retrieving instances in Security Group: %w", err)
	}

	for zone, instances := range instancesByZone {
		for _, instance := range instances {
			publicIP := emptyIPAddressVisualization
			if instance.PublicIP != nil && (!instance.PublicIP.IsUnspecified() || len(instance.PublicIP) > 0) {
				publicIP = instance.PublicIP.String()
			}

			out.Instances = append(out.Instances, securityGroupInstanceOutput{
				Name:     utils.DefaultString(&instance.Name, "-"),
				PublicIP: publicIP,
				ID:       instance.ID.String(),
				Zone:     zone,
			})
		}
	}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(securityGroupCmd, &securityGroupShowCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
