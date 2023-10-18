package cmd

import (
	"encoding/json"
	"errors"
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

type iamOrgPolicyUpdate struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	// Root settings
	DefaultServiceStrategy string `cli-flag:"default-service-strategy" cli-usage:"The default service strategy applies to all service classes that have not been explicitly configured. Allowed values are 'allow' and 'deny'"`

	// Actions (mutually exclusive)
	Clear           bool   `cli-flag:"clear" cli-usage:"Remove all existing service classes"`
	ReplacePolicy   string `cli-flag:"replace-policy" cli-usage:"Replace the whole policy. New policy must be provided in JSON format. If value '-' is used, policy is read from stdin"`
	DeleteService   string `cli-flag:"delete-service" cli-usage:"Delete service class"`
	AddService      string `cli-flag:"add-service" cli-usage:"Add service class"`
	AddServiceRules string `cli-flag:"add-service-rules" cli-usage:"Update service class by adding more rules"`

	// Service level settings
	ServiceType string `cli-flag:"service-type" cli-usage:"Default Strategy for service type. Allowed values: 'allow', 'deny' and 'rules'. Required for --add-service, optional for --update-service"`
}

func (c *iamOrgPolicyUpdate) cmdAliases() []string { return nil }

func (c *iamOrgPolicyUpdate) cmdShort() string {
	return "Update IAM Org policy"
}

func (c *iamOrgPolicyUpdate) cmdLong() string {
	return fmt.Sprintf(`This command updates an IAM Organization Policy.

Command requires exacly one flag to be set from the following: --clear, --replace-policy, --delete-service, --add-service, --update-service.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&iamOrgPolicyShowOutput{}), ", "))
}

func (c *iamOrgPolicyUpdate) cmdUse() string {
	return "exo iam org-policy update [flags] test"
}

func (c *iamOrgPolicyUpdate) cmdPreRun(cmd *cobra.Command, args []string) error {
	err := cliCommandDefaultPreRun(c, cmd, args)
	if err != nil {
		return err
	}

	counter := 0
	if c.Clear {
		counter++
	}
	if c.ReplacePolicy != "" {
		counter++
	}
	if c.DeleteService != "" {
		counter++
	}
	if c.AddService != "" {
		counter++
	}
	if c.AddServiceRules != "" {
		counter++
	}

	if counter == 0 || counter > 1 {
		return errors.New("command requires exacly one flag to be set from the following: --clear, --replace-policy, --delete-service, --add-service, --update-service")
	}

	return nil
}

func (c *iamOrgPolicyUpdate) cmdRun(cmd *cobra.Command, args []string) error {
	zone := account.CurrentAccount.DefaultZone
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

	policy, err := globalstate.EgoscaleClient.GetIAMOrgPolicy(ctx, zone)
	if err != nil {
		return err
	}

	switch {
	case c.ReplacePolicy != "":
		if c.ReplacePolicy == "-" {
			inputReader := cmd.InOrStdin()
			b, err := io.ReadAll(inputReader)
			if err != nil {
				return fmt.Errorf("failed to read policy from stdin: %w", err)
			}

			c.ReplacePolicy = string(b)
		}

		var obj iamOrgPolicyShowOutput
		err := json.Unmarshal([]byte(c.ReplacePolicy), &obj)
		if err != nil {
			return fmt.Errorf("failed to parse policy: %w", err)
		}

		policy = &exoscale.IAMPolicy{
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

						if len(rl.Resources) > 0 {
							rule.Resources = rl.Resources
						}

						service.Rules = append(service.Rules, rule)
					}
				}

				policy.Services[name] = service
			}
		}
	case c.Clear:
		policy.Services = map[string]exoscale.IAMPolicyService{}
	case c.DeleteService != "":
		if _, found := policy.Services[c.DeleteService]; !found {
			return fmt.Errorf("service class %q not found", c.DeleteService)
		}

		delete(policy.Services, c.DeleteService)
	case c.AddService != "":
		if _, found := policy.Services[c.AddService]; found {
			return fmt.Errorf("service class %q already exists in policy", c.AddService)
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
			// For rules service type arguments will hold pairs of "action" and "expression" definitions.
			if len(args) == 0 || len(args)%2 != 0 {
				return errors.New("at least one rule must be specified when --service-type is 'rules'")
			}

			service.Rules = []exoscale.IAMPolicyServiceRule{}

			for i := 0; i < len(args); i = i + 2 {
				rule := exoscale.IAMPolicyServiceRule{
					Action:     &args[i],
					Expression: &args[i+1],
				}

				service.Rules = append(service.Rules, rule)
			}
		}

		policy.Services[c.AddService] = service
	case c.AddServiceRules != "":
		if _, found := policy.Services[c.AddServiceRules]; !found {
			return fmt.Errorf("service class %q not found", c.AddService)
		}

		service := policy.Services[c.AddServiceRules]

		if *service.Type != "rules" {
			return fmt.Errorf("cannot add rules to service class of type %q", *service.Type)
		}

		// For rules service type arguments must hold pairs of "action" and "expression" definitions.
		if len(args) == 0 || len(args)%2 != 0 {
			return errors.New("at least one rule must be specified when --service-type is 'rules'")
		}

		for i := 0; i < len(args); i = i + 2 {
			rule := exoscale.IAMPolicyServiceRule{
				Action:     &args[i],
				Expression: &args[i+1],
			}

			service.Rules = append(service.Rules, rule)
		}

		policy.Services[c.AddServiceRules] = service
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
