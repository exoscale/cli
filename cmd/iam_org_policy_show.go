package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type iamOrgPolicyShowOutput struct {
	DefaultServiceStrategy string                                   `json:"default-service-strategy"`
	Services               map[string]iamOrgPolicyServiceShowOutput `json:"services"`
}

type iamOrgPolicyServiceShowOutput struct {
	Type  string                              `json:"type"`
	Rules []iamOrgPolicyServiceRuleShowOutput `json:"rules"`
}

type iamOrgPolicyServiceRuleShowOutput struct {
	Action     string   `json:"action"`
	Expression string   `json:"expression"`
	Resources  []string `json:"resources,omitempty"`
}

func (o *iamOrgPolicyShowOutput) ToJSON() { output.JSON(o) }
func (o *iamOrgPolicyShowOutput) ToText() { output.Text(o) }
func (o *iamOrgPolicyShowOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	t.SetAutoMergeCells(true)

	t.SetHeader([]string{
		"Service",
		fmt.Sprintf("Type (default strategy \"%s\")", o.DefaultServiceStrategy),
		"Rule Action",
		"Rule Expression",
		"Rule Resources",
	})
	defer t.Render()

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
				strings.Join(rule.Resources, ","),
			})
		}
	}
}

type iamOrgPolicyShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`
}

func (c *iamOrgPolicyShowCmd) cmdAliases() []string { return gShowAlias }

func (c *iamOrgPolicyShowCmd) cmdShort() string {
	return "Show Org policy details"
}

func (c *iamOrgPolicyShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows IAM Org Policy details.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&iamOrgPolicyShowOutput{}), ", "))
}

func (c *iamOrgPolicyShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamOrgPolicyShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	zone := account.CurrentAccount.DefaultZone
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

	policy, err := globalstate.EgoscaleClient.GetIAMOrgPolicy(ctx, zone)
	if err != nil {
		return err
	}

	out := iamOrgPolicyShowOutput{
		DefaultServiceStrategy: policy.DefaultServiceStrategy,
		Services:               map[string]iamOrgPolicyServiceShowOutput{},
	}

	for name, service := range policy.Services {
		rules := []iamOrgPolicyServiceRuleShowOutput{}
		for _, rule := range service.Rules {
			rules = append(rules, iamOrgPolicyServiceRuleShowOutput{
				Action:     utils.DefaultString(rule.Action, ""),
				Expression: utils.DefaultString(rule.Expression, ""),
				Resources:  rule.Resources,
			})
		}

		out.Services[name] = iamOrgPolicyServiceShowOutput{
			Type:  utils.DefaultString(service.Type, ""),
			Rules: rules,
		}
	}

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(iamOrgPolicyCmd, &iamOrgPolicyShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
