package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type securityGroupShowItemOutput struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Source      string `json:"source"`
	Port        string `json:"port"`
	Protocol    string `json:"protocol"`
	Description string `json:"description,omitempty"`
}

type securityGroupShowOutput []securityGroupShowItemOutput

func (o *securityGroupShowOutput) toJSON()  { outputJSON(o) }
func (o *securityGroupShowOutput) toText()  { outputText(o) }
func (o *securityGroupShowOutput) toTable() { outputTable(o) }

func init() {
	firewallCmd.AddCommand(&cobra.Command{
		Use:   "show NAME|ID",
		Short: "Show a Security Group rules details",
		Long: fmt.Sprintf(`This command shows a Security Group details.

Supported output template annotations: %s`,
			strings.Join(outputterTemplateAnnotations(&securityGroupShowOutput{}), ", ")),
		Aliases: gListAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("show expects one Security Group by name or id")
			}

			return output(showSecurityGroup(args[0]))
		},
	})
}

func showSecurityGroup(name string) (outputter, error) {
	sg, err := getSecurityGroupByNameOrID(name)
	if err != nil {
		return nil, err
	}

	out := securityGroupShowOutput{}

	for _, rule := range sg.IngressRule {
		out = append(out, securityGroupShowItemOutput{
			Type:        "ingress",
			ID:          rule.RuleID.String(),
			Source:      formatRuleSource(rule),
			Port:        formatRulePort(rule),
			Protocol:    rule.Protocol,
			Description: rule.Description,
		})
	}

	for _, rule := range sg.EgressRule {
		out = append(out, securityGroupShowItemOutput{
			Type:        "egress",
			ID:          rule.RuleID.String(),
			Source:      formatRuleSource((egoscale.IngressRule)(rule)),
			Port:        formatRulePort((egoscale.IngressRule)(rule)),
			Protocol:    rule.Protocol,
			Description: rule.Description,
		})
	}

	return &out, nil
}
