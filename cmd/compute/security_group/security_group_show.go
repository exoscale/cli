package security_group

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type securityGroupRuleOutput struct {
	ID            string  `json:"id"`
	Description   string  `json:"description"`
	ICMPCode      *int64  `json:"icmp_code,omitempty"`
	ICMPType      *int64  `json:"icmp_type,omitempty"`
	Network       *string `json:"network,omitempty"`
	Protocol      string  `json:"protocol"`
	SecurityGroup *string `json:"security_group,omitempty"`
	StartPort     *uint16 `json:"start_port,omitempty"`
	EndPort       *uint16 `json:"end_port,omitempty"`
}

type securityGroupShowOutput struct {
	ID              string                        `json:"id"`
	Name            string                        `json:"name"`
	Description     string                        `json:"description"`
	ExternalSources []string                      `json:"external_sources"`
	IngressRules    []securityGroupRuleOutput     `json:"ingress_rules"`
	EgressRules     []securityGroupRuleOutput     `json:"egress_rules"`
	Instances       []securityGroupInstanceOutput `json:"instances"`
}

type securityGroupInstanceOutput struct {
	Name     string `json:"name"`
	PublicIP string `json:"public_ip"`
	ID       string `json:"id"`
	Zone     string `json:"zone"`
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
				r := []string{rule.ID, rule.Description, strings.ToUpper(rule.Protocol)}

				if rule.Network != nil {
					r = append(r, *rule.Network)
				} else {
					r = append(r, *rule.SecurityGroup)
				}

				if strings.HasPrefix(rule.Protocol, "icmp") {
					r = append(r, fmt.Sprintf("ICMP code:%d type:%d", *rule.ICMPCode, *rule.ICMPType))
				} else if rule.StartPort != nil {
					r = append(r, func() string {
						if *rule.StartPort == *rule.EndPort {
							return fmt.Sprint(*rule.StartPort)
						}
						return fmt.Sprintf("%d-%d", *rule.StartPort, *rule.EndPort)
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
			r = append(r, instance.Zone)

			at.Append(r)
		}

		at.Render()

		return buf.String()
	}

	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Security Group"})
	defer t.Render()

	t.Append([]string{"ID", o.ID})
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

func (c *securityGroupShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	zone := account.CurrentAccount.DefaultZone

	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

	securityGroup, err := globalstate.EgoscaleClient.FindSecurityGroup(ctx, zone, c.SecurityGroup)
	if err != nil {
		return err
	}

	externalSources := make([]string, 0)
	if securityGroup.ExternalSources != nil {
		externalSources = *securityGroup.ExternalSources
	}
	out := securityGroupShowOutput{
		ID:              *securityGroup.ID,
		Name:            *securityGroup.Name,
		Description:     utils.DefaultString(securityGroup.Description, ""),
		ExternalSources: externalSources,
		IngressRules:    make([]securityGroupRuleOutput, 0),
		EgressRules:     make([]securityGroupRuleOutput, 0),
	}

	for _, rule := range securityGroup.Rules {
		or := securityGroupRuleOutput{
			ID:          *rule.ID,
			Description: utils.DefaultString(rule.Description, ""),
			ICMPCode:    rule.ICMPCode,
			ICMPType:    rule.ICMPType,
			Network: func() *string {
				if rule.Network != nil {
					v := rule.Network.String()
					return &v
				}
				return nil
			}(),
			Protocol:  *rule.Protocol,
			StartPort: rule.StartPort,
			EndPort:   rule.EndPort,
		}

		if rule.SecurityGroupID != nil {
			ruleSecurityGroup, err := globalstate.EgoscaleClient.GetSecurityGroup(ctx, zone, *rule.SecurityGroupID)
			if err != nil {
				return fmt.Errorf("error retrieving Security Group: %w", err)
			}
			ruleSG := "SG:" + *ruleSecurityGroup.Name
			or.SecurityGroup = &ruleSG
		}
		if rule.SecurityGroupName != nil {
			ruleSG := "PUBLIC-SG:" + *rule.SecurityGroupName
			or.SecurityGroup = &ruleSG
		}

		if *rule.FlowDirection == "ingress" {
			out.IngressRules = append(out.IngressRules, or)
		} else {
			out.EgressRules = append(out.EgressRules, or)
		}
	}

	instances, err := utils.GetInstancesInSecurityGroup(ctx, globalstate.EgoscaleClient, *securityGroup.ID)
	if err != nil {
		return fmt.Errorf("error retrieving instances in Security Group: %w", err)
	}

	for _, vm := range instances {
		publicIP := exocmd.EmptyIPAddressVisualization
		if vm.PublicIPAddress != nil && (!vm.PublicIPAddress.IsUnspecified() || len(*vm.PublicIPAddress) > 0) {
			publicIP = vm.PublicIPAddress.String()
		}

		out.Instances = append(out.Instances, securityGroupInstanceOutput{
			Name:     utils.DefaultString(vm.Name, "-"),
			PublicIP: publicIP,
			ID:       utils.DefaultString(vm.ID, "-"),
			Zone:     utils.DefaultString(vm.Zone, "-"),
		})
	}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(securityGroupCmd, &securityGroupShowCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
