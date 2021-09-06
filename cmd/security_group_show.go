package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/cli/table"
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
	ID           string                    `json:"id"`
	Name         string                    `json:"name"`
	Description  string                    `json:"description"`
	IngressRules []securityGroupRuleOutput `json:"ingress_rules"`
	EgressRules  []securityGroupRuleOutput `json:"egress_rules"`
}

func (o *securityGroupShowOutput) toJSON() { outputJSON(o) }
func (o *securityGroupShowOutput) toText() { outputText(o) }
func (o *securityGroupShowOutput) toTable() {
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

	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Security Group"})
	defer t.Render()

	t.Append([]string{"ID", o.ID})
	t.Append([]string{"Name", o.Name})
	t.Append([]string{"Description", o.Description})
	t.Append([]string{"Ingress Rules", formatRule(o.IngressRules)})
	t.Append([]string{"Egress Rules", formatRule(o.EgressRules)})
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
	return output(showSecurityGroup(gCurrentAccount.DefaultZone, c.SecurityGroup))
}

func showSecurityGroup(zone, x string) (outputter, error) {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

	securityGroup, err := cs.FindSecurityGroup(ctx, zone, x)
	if err != nil {
		return nil, err
	}

	out := securityGroupShowOutput{
		ID:           *securityGroup.ID,
		Name:         *securityGroup.Name,
		Description:  defaultString(securityGroup.Description, ""),
		IngressRules: make([]securityGroupRuleOutput, 0),
		EgressRules:  make([]securityGroupRuleOutput, 0),
	}

	for _, rule := range securityGroup.Rules {
		or := securityGroupRuleOutput{
			ID:          *rule.ID,
			Description: defaultString(rule.Description, ""),
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
				return nil, fmt.Errorf("error retrieving Security Group: %v", err)
			}
			or.SecurityGroup = ruleSecurityGroup.Name
		}

		if *rule.FlowDirection == "ingress" {
			out.IngressRules = append(out.IngressRules, or)
		} else {
			out.EgressRules = append(out.EgressRules, or)
		}
	}

	return &out, nil
}

func init() {
	cobra.CheckErr(registerCLICommand(securityGroupCmd, &securityGroupShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
