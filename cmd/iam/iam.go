package iam

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"

	v3 "github.com/exoscale/egoscale/v3"
)

var iamCmd = &cobra.Command{
	Use:   "iam",
	Short: "Identity and Access Management",
}

func init() {
	exocmd.RootCmd.AddCommand(iamCmd)
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
			t.Append([]string{name, service.Type, "", ""})
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

func iamPolicyFromJSON(data []byte) (*v3.IAMPolicy, error) {
	var obj iamPolicyOutput
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return nil, fmt.Errorf("failed to parse policy: %w", err)
	}

	policy := v3.IAMPolicy{
		DefaultServiceStrategy: v3.IAMPolicyDefaultServiceStrategy(obj.DefaultServiceStrategy),
		Services:               map[string]v3.IAMServicePolicy{},
	}

	if len(obj.Services) > 0 {
		for name, sv := range obj.Services {
			service := v3.IAMServicePolicy{
				Type: v3.IAMServicePolicyType(sv.Type),
			}

			if len(sv.Rules) > 0 {
				service.Rules = []v3.IAMServicePolicyRule{}
				for _, rl := range sv.Rules {

					rule := v3.IAMServicePolicyRule{
						Action: v3.IAMServicePolicyRuleAction(rl.Action),
					}

					if rl.Expression != "" {
						rule.Expression = rl.Expression
					}

					service.Rules = append(service.Rules, rule)
				}
			}

			policy.Services[name] = service
		}
	}

	return &policy, nil
}

type iamAssumeRolePolicyOutput struct {
	Rules []iamPolicyServiceRuleOutput `json:"rules"`
}

func (o *iamAssumeRolePolicyOutput) ToJSON() { output.JSON(o) }
func (o *iamAssumeRolePolicyOutput) ToText() { output.Text(o) }
func (o *iamAssumeRolePolicyOutput) ToTable() {
	t := table.NewTable(os.Stdout)

	t.SetHeader([]string{
		"Rule Action",
		"Rule Expression",
	})

	// use underlying tablewriter.Render to display table even with empty rows.
	defer t.Table.Render()

	for _, rule := range o.Rules {
		t.Append([]string{
			rule.Action,
			rule.Expression,
		})
	}
}

func iamAssumeRolePolicyToOutput(policy *v3.IAMAssumeRolePolicy) *iamAssumeRolePolicyOutput {
	if policy == nil {
		return nil
	}

	out := &iamAssumeRolePolicyOutput{
		Rules: []iamPolicyServiceRuleOutput{},
	}

	for _, rule := range policy.Rules {
		out.Rules = append(out.Rules, iamPolicyServiceRuleOutput{
			Action:     string(rule.Action),
			Expression: rule.Expression,
		})
	}

	return out
}

func iamAssumeRolePolicyFromJSON(data []byte) (*v3.IAMAssumeRolePolicy, error) {
	var obj iamAssumeRolePolicyOutput
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return nil, fmt.Errorf("failed to parse assume role policy: %w", err)
	}

	policy := v3.IAMAssumeRolePolicy{}

	if len(obj.Rules) > 0 {
		policy.Rules = []v3.IAMServicePolicyRule{}
		for _, rl := range obj.Rules {
			rule := v3.IAMServicePolicyRule{
				Action: v3.IAMServicePolicyRuleAction(rl.Action),
			}

			if rl.Expression != "" {
				rule.Expression = rl.Expression
			}

			policy.Rules = append(policy.Rules, rule)
		}
	}

	return &policy, nil
}
