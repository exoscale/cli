package cmd

import (
	"fmt"
	"os"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/spf13/cobra"
)

var iamCmd = &cobra.Command{
	Use:   "iam",
	Short: "Identity and Access Management",
}

func init() {
	RootCmd.AddCommand(iamCmd)
}

type iamPolicyOutput struct {
	DefaultServiceStrategy string                            `json:"default-service-strategy"`
	Services               map[string]iamPolicyServiceOutput `json:"services"`
}

type iamPolicyServiceOutput struct {
	Type  string                       `json:"type"`
	Rules []iamPolicyServiceRuleOutput `json:"rules"`
}

type iamPolicyServiceRuleOutput struct {
	Action     string `json:"action"`
	Expression string `json:"expression"`
}

func (o *iamPolicyOutput) ToJSON() { output.JSON(o) }
func (o *iamPolicyOutput) ToText() { output.Text(o) }
func (o *iamPolicyOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	t.SetAutoMergeCellsByColumnIndex([]int{0, 1})

	t.SetHeader([]string{
		"Service",
		fmt.Sprintf("Type (default strategy \"%s\")", o.DefaultServiceStrategy),
		"Rule Action",
		"Rule Expression",
	})

	// use underlying tablewriter.Render to display table even with empty rows
	// as default strategy is in header.
	defer t.Table.Render()

	for name, service := range o.Services {
		if len(service.Rules) == 0 {
			t.Append([]string{name, service.Type, "", "", ""})
			continue
		}

		for _, rule := range service.Rules {
			t.Append([]string{
				name,
				service.Type,
				rule.Action,
				rule.Expression,
			})
		}
	}
}
