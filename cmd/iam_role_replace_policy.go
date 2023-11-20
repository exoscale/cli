package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	exoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type iamRolePolicyReplaceCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	Role   string `cli-arg:"#" cli-usage:"ID|NAME"`
	Policy string `cli-arg:"#"`

	_ bool `cli-cmd:"replace-policy"`
}

func (c *iamRolePolicyReplaceCmd) cmdAliases() []string { return nil }

func (c *iamRolePolicyReplaceCmd) cmdShort() string {
	return "Replace IAM Role Policy"
}

func (c *iamRolePolicyReplaceCmd) cmdLong() string {
	return fmt.Sprintf(`This command replaces complete IAM Role Policy with the new one provided in JSON format.
To read Policy from STDIN provide '-' as argument.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&iamPolicyOutput{}), ", "))
}

func (c *iamRolePolicyReplaceCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamRolePolicyReplaceCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	if c.Role == "" {
		return errors.New("Role ID not provided")
	}

	zone := account.CurrentAccount.DefaultZone
	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone),
	)

	if _, err := uuid.Parse(c.Role); err != nil {
		roles, err := globalstate.EgoscaleClient.ListIAMRoles(ctx, zone)
		if err != nil {
			return err
		}

		for _, role := range roles {
			if role.Name != nil && *role.Name == c.Role {
				c.Role = *role.ID
				break
			}
		}
	}

	role, err := globalstate.EgoscaleClient.GetIAMRole(ctx, zone, c.Role)
	if err != nil {
		return err
	}

	if c.Policy == "-" {
		inputReader := cmd.InOrStdin()
		b, err := io.ReadAll(inputReader)
		if err != nil {
			return fmt.Errorf("failed to read policy from stdin: %w", err)
		}

		c.Policy = string(b)
	}

	var obj iamPolicyOutput
	err = json.Unmarshal([]byte(c.Policy), &obj)
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

	role.Policy = policy

	err = globalstate.EgoscaleClient.UpdateIAMRolePolicy(ctx, zone, role)
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&iamRoleShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			Role:               *role.ID,
			Policy:             true,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(iamRoleCmd, &iamRolePolicyReplaceCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
