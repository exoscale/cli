package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	exoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type iamOrgPolicyReplaceCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	Policy string `cli-arg:"#"`

	_ bool `cli-cmd:"replace"`
}

func (c *iamOrgPolicyReplaceCmd) cmdAliases() []string { return nil }

func (c *iamOrgPolicyReplaceCmd) cmdShort() string {
	return "Replace Org policy"
}

func (c *iamOrgPolicyReplaceCmd) cmdLong() string {
	return fmt.Sprintf(`This command replaces complete IAM Organization Policy with the new one provided in JSON format.
To read Policy from STDIN provide '-' as argument.

Pro Tip: you can get policy in JSON format with command:

	exo iam org-policy show --output-format json

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&iamOrgPolicyShowOutput{}), ", "))
}

func (c *iamOrgPolicyReplaceCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamOrgPolicyReplaceCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	zone := account.CurrentAccount.DefaultZone
	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone),
	)

	if c.Policy == "-" {
		inputReader := cmd.InOrStdin()
		b, err := io.ReadAll(inputReader)
		if err != nil {
			return fmt.Errorf("failed to read policy from stdin: %w", err)
		}

		c.Policy = string(b)
	}

	var obj iamOrgPolicyShowOutput
	err := json.Unmarshal([]byte(c.Policy), &obj)
	if err != nil {
		return fmt.Errorf("failed to parse policy: %w", err)
	}

	policy := &exoscale.IAMPolicy{
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

	err = globalstate.EgoscaleClient.UpdateIAMOrgPolicy(ctx, zone, policy)
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&iamOrgPolicyShowCmd{
			cliCommandSettings: c.cliCommandSettings,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(iamOrgPolicyCmd, &iamOrgPolicyReplaceCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
