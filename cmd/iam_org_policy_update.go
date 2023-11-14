package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	exoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"

	"github.com/exoscale/cli/utils"
)

type iamOrgPolicyUpdate struct {
	cliCommandSettings `cli-cmd:"-"`

	DeleteService string `cli-flag:"delete-service" cli-usage:"Delete service class"`
	AddService    string `cli-flag:"add-service" cli-usage:"Add service class"`
	UpdateService string `cli-flag:"update-service" cli-usage:"Update service class"`

	ServiceType            string `cli-flag:"service-type" cli-usage:"Default Strategy for service type. Allowed values: 'allow', 'deny' and 'rules'. Used with --add-service"`
	AppendServiceRuleAllow string `cli-flag:"append-service-rule-allow" cli-usage:"Append service rule of type 'allow' to the end of the rules list"`
	AppendServiceRuleDeny  string `cli-flag:"append-service-rule-deny" cli-usage:"Append service rule of type 'deny' to the end of the rules list"`

	_ bool `cli-cmd:"update"`
}

func (c *iamOrgPolicyUpdate) cmdAliases() []string { return nil }

func (c *iamOrgPolicyUpdate) cmdShort() string {
	return "Update IAM Org policy"
}

func (c *iamOrgPolicyUpdate) cmdLong() string {
	return fmt.Sprintf(`This command updates an IAM Organization Policy.

Command can update only a single service and a single service rule.
For bigger changes to Org Policy command must be executed multiple times.

Pro Tip: use 'exo iam org-policy replace' command to edit Org Policy as JSON document.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&iamOrgPolicyShowOutput{}), ", "))
}

func (c *iamOrgPolicyUpdate) cmdPreRun(cmd *cobra.Command, args []string) error {
	err := cliCommandDefaultPreRun(c, cmd, args)
	if err != nil {
		return err
	}

	counter := 0
	if c.DeleteService != "" {
		counter++
	}
	if c.AddService != "" {
		counter++
	}
	if c.UpdateService != "" {
		counter++
	}

	if counter != 1 {
		return errors.New("only one service can be updated")
	}

	if c.AddService != "" || c.UpdateService != "" {
		if c.AppendServiceRuleAllow != "" && c.AppendServiceRuleDeny != "" {
			return errors.New("only one service rule can be added")
		}
	}

	return nil
}

func (c *iamOrgPolicyUpdate) cmdRun(cmd *cobra.Command, _ []string) error {
	zone := account.CurrentAccount.DefaultZone
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

	policy, err := globalstate.EgoscaleClient.GetIAMOrgPolicy(ctx, zone)
	if err != nil {
		return err
	}

	switch {
	case c.DeleteService != "":
		if _, found := policy.Services[c.DeleteService]; !found {
			return fmt.Errorf("service class %q not found", c.DeleteService)
		}

		delete(policy.Services, c.DeleteService)
	case c.AddService != "":
		if _, found := policy.Services[c.AddService]; found {
			return fmt.Errorf("service class %q already exists", c.AddService)
		}

		service := exoscale.IAMPolicyService{}

		if c.ServiceType == "" {
			return errors.New("--service-type must be specified when --add-service is used")
		}
		if c.ServiceType != "allow" && c.ServiceType != "deny" && c.ServiceType != "rules" {
			return errors.New("allowed values for --service-type are: 'allow', 'deny' and 'rules'")
		}

		service.Type = &c.ServiceType

		if c.ServiceType == "rules" {
			if c.AppendServiceRuleAllow == "" && c.AppendServiceRuleDeny == "" {
				return errors.New("service rule must be specified, use --append-service-rule-allow or --append-service-rule-deny")
			}

			var rule exoscale.IAMPolicyServiceRule

			if c.AppendServiceRuleAllow != "" {
				rule = exoscale.IAMPolicyServiceRule{
					Action:     utils.NonEmptyStringPtr("allow"),
					Expression: &c.AppendServiceRuleAllow,
				}
			}

			if c.AppendServiceRuleDeny != "" {
				rule = exoscale.IAMPolicyServiceRule{
					Action:     utils.NonEmptyStringPtr("deny"),
					Expression: &c.AppendServiceRuleDeny,
				}
			}

			service.Rules = []exoscale.IAMPolicyServiceRule{rule}
		}

		policy.Services[c.AddService] = service
	case c.UpdateService != "":
		service, found := policy.Services[c.UpdateService]
		if !found {
			return fmt.Errorf("service class %q does not exist", c.UpdateService)
		}

		if c.AppendServiceRuleAllow == "" && c.AppendServiceRuleDeny == "" {
			return errors.New("service rule must be specified, use --append-service-rule-allow or --append-service-rule-deny")
		}

		var rule exoscale.IAMPolicyServiceRule

		if c.AppendServiceRuleAllow != "" {
			rule = exoscale.IAMPolicyServiceRule{
				Action:     utils.NonEmptyStringPtr("allow"),
				Expression: &c.AppendServiceRuleAllow,
			}
		}

		if c.AppendServiceRuleDeny != "" {
			rule = exoscale.IAMPolicyServiceRule{
				Action:     utils.NonEmptyStringPtr("deny"),
				Expression: &c.AppendServiceRuleDeny,
			}
		}

		service.Rules = append(service.Rules, rule)

		policy.Services[c.UpdateService] = service
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
	cobra.CheckErr(registerCLICommand(iamOrgPolicyCmd, &iamOrgPolicyUpdate{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
