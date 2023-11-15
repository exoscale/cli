package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

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
		strings.Join(output.TemplateAnnotations(&iamPolicyServiceOutput{}), ", "))
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

	out := iamPolicyOutput{
		DefaultServiceStrategy: policy.DefaultServiceStrategy,
		Services:               map[string]iamPolicyServiceOutput{},
	}

	for name, service := range policy.Services {
		rules := []iamPolicyServiceRuleOutput{}
		for _, rule := range service.Rules {
			rules = append(rules, iamPolicyServiceRuleOutput{
				Action:     utils.DefaultString(rule.Action, ""),
				Expression: utils.DefaultString(rule.Expression, ""),
			})
		}

		out.Services[name] = iamPolicyServiceOutput{
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
