package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/egoscale"
)

type firewallShowItemOutput struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Source      string `json:"source"`
	Port        string `json:"port"`
	Protocol    string `json:"protocol"`
	Description string `json:"description,omitempty"`
}

type firewallShowOutput []firewallShowItemOutput

func (o *firewallShowOutput) ToJSON()  { output.JSON(o) }
func (o *firewallShowOutput) ToText()  { output.Text(o) }
func (o *firewallShowOutput) ToTable() { output.Table(o) }

func init() {
	firewallCmd.AddCommand(&cobra.Command{
		Use:   "show NAME|ID",
		Short: "Show a Security Group rules details",
		Long: fmt.Sprintf(`This command shows a Security Group details.

Supported output template annotations: %s`,
			strings.Join(output.TemplateAnnotations(&firewallShowOutput{}), ", ")),
		Aliases: gListAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("show expects one Security Group by name or id")
			}

			return printOutput(showFirewall(args[0]))
		},
	})
}

func showFirewall(name string) (output.Outputter, error) {
	sg, err := getSecurityGroupByNameOrID(name)
	if err != nil {
		return nil, err
	}

	out := firewallShowOutput{}

	for _, rule := range sg.IngressRule {
		out = append(out, firewallShowItemOutput{
			Type:        "ingress",
			ID:          rule.RuleID.String(),
			Source:      formatRuleSource(rule),
			Port:        formatRulePort(rule),
			Protocol:    rule.Protocol,
			Description: rule.Description,
		})
	}

	for _, rule := range sg.EgressRule {
		out = append(out, firewallShowItemOutput{
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
