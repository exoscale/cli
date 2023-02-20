package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
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

func (o *securityGroupShowOutput) toJSON() { outputJSON(o) }
func (o *securityGroupShowOutput) toText() { outputText(o) }
func (o *securityGroupShowOutput) toTable() {
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
					r = append(r, "SG:"+*rule.SecurityGroup)
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
		if len(instances) > 0 {
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

		return "-"
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
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	SecurityGroup string `cli-arg:"#" cli-usage:"NAME|ID"`
}

func (c *securityGroupShowCmd) cmdAliases() []string { return gShowAlias }

func (c *securityGroupShowCmd) cmdShort() string {
	return "Show a Security Group details"
}

func (c *securityGroupShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows a Compute instance Security Group details.

Supported output template annotations for Security Group: %s

Supported output template annotations for Security Group rules: %s`,
		strings.Join(outputterTemplateAnnotations(&securityGroupShowOutput{}), ", "),
		strings.Join(outputterTemplateAnnotations(&securityGroupRuleOutput{}), ", "))
}

func (c *securityGroupShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *securityGroupShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	zone := gCurrentAccount.DefaultZone

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

	securityGroup, err := cs.FindSecurityGroup(ctx, zone, c.SecurityGroup)
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
			ruleSecurityGroup, err := cs.GetSecurityGroup(ctx, zone, *rule.SecurityGroupID)
			if err != nil {
				return fmt.Errorf("error retrieving Security Group: %w", err)
			}
			or.SecurityGroup = ruleSecurityGroup.Name
		}

		if *rule.FlowDirection == "ingress" {
			out.IngressRules = append(out.IngressRules, or)
		} else {
			out.EgressRules = append(out.EgressRules, or)
		}
	}

	instances, err := utils.GetInstancesInSecurityGroup(ctx, cs, *securityGroup.ID, zone)
	if err != nil {
		return fmt.Errorf("error retrieving instances in Security Group: %w", err)
	}

	for _, instance := range instances {
		publicIP := emptyIPAddressVisualization
		if instance.PublicIPAddress != nil && (!instance.PublicIPAddress.IsUnspecified() || len(*instance.PublicIPAddress) > 0) {
			publicIP = instance.PublicIPAddress.String()
		}

		out.Instances = append(out.Instances, securityGroupInstanceOutput{
			Name:     utils.DefaultString(instance.Name, "-"),
			PublicIP: publicIP,
			ID:       utils.DefaultString(instance.ID, "-"),
			Zone:     utils.DefaultString(instance.Zone, "-"),
		})
	}

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(securityGroupCmd, &securityGroupShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
