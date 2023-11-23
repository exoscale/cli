package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type iamRoleShowOutput struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Editable    bool              `json:"editable"`
	Labels      map[string]string `json:"labels"`
	Permissions []string          `json:"permission"`
}

func (o *iamRoleShowOutput) ToJSON()  { output.JSON(o) }
func (o *iamRoleShowOutput) ToText()  { output.Text(o) }
func (o *iamRoleShowOutput) ToTable() { output.Table(o) }

type iamRoleShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Policy bool `cli-flag:"policy" cli-usage:"Print IAM Role policy"`

	Role string `cli-arg:"#" cli-usage:"ID|NAME"`
}

func (c *iamRoleShowCmd) cmdAliases() []string { return gShowAlias }

func (c *iamRoleShowCmd) cmdShort() string {
	return "Show Role details"
}

func (c *iamRoleShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows IAM Role details.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&iamRoleShowOutput{}), ", "))
}

func (c *iamRoleShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamRoleShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	if c.Role == "" {
		return errors.New("Role ID not provided")
	}

	zone := account.CurrentAccount.DefaultZone
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

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

	if c.Policy {
		policy := role.Policy

		out := iamPolicyOutput{
			DefaultServiceStrategy: policy.DefaultServiceStrategy,
			Services:               map[string]iamPolicyServiceOutput{},
		}

		for name, service := range policy.Services {
			rules := []iamPolicyServiceRuleOutput{}
			if service.Type != nil && *service.Type == "rules" {
				for _, rule := range service.Rules {
					rules = append(rules, iamPolicyServiceRuleOutput{
						Action:     utils.DefaultString(rule.Action, ""),
						Expression: utils.DefaultString(rule.Expression, ""),
					})
				}
			}

			out.Services[name] = iamPolicyServiceOutput{
				Type:  utils.DefaultString(service.Type, ""),
				Rules: rules,
			}
		}

		return c.outputFunc(&out, nil)
	}

	out := iamRoleShowOutput{
		ID:          utils.DefaultString(role.ID, ""),
		Description: utils.DefaultString(role.Description, ""),
		Editable:    utils.DefaultBool(role.Editable, false),
		Labels:      role.Labels,
		Name:        utils.DefaultString(role.Name, ""),
		Permissions: role.Permissions,
	}

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(iamRoleCmd, &iamRoleShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
