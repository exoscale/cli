package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"

	exoscale "github.com/exoscale/egoscale/v2"
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

func iamPolicyFromJSON(data []byte) (*exoscale.IAMPolicy, error) {
	var obj iamPolicyOutput
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return nil, fmt.Errorf("failed to parse policy: %w", err)
	}

	policy := exoscale.IAMPolicy{
		DefaultServiceStrategy: obj.DefaultServiceStrategy,
		Services:               map[string]exoscale.IAMPolicyService{},
	}

	if len(obj.Services) > 0 {
		for name, sv := range obj.Services {
			service := exoscale.IAMPolicyService{
				Type: func() *string {
					t := sv.Type
					return &t
				}(),
			}

			if len(sv.Rules) > 0 {
				service.Rules = []exoscale.IAMPolicyServiceRule{}
				for _, rl := range sv.Rules {

					rule := exoscale.IAMPolicyServiceRule{
						Action: func() *string {
							t := rl.Action
							return &t
						}(),
					}

					if rl.Expression != "" {
						rule.Expression = func() *string {
							t := rl.Expression
							return &t
						}()
					}

					service.Rules = append(service.Rules, rule)
				}
			}

			policy.Services[name] = service
		}
	}

	return &policy, nil
}
