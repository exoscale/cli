package iam

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
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
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Policy bool `cli-flag:"policy" cli-usage:"Print IAM Role policy"`

	Role string `cli-arg:"#" cli-usage:"ID|NAME"`
}

func (c *iamRoleShowCmd) CmdAliases() []string { return exocmd.GShowAlias }

func (c *iamRoleShowCmd) CmdShort() string {
	return "Show Role details"
}

func (c *iamRoleShowCmd) CmdLong() string {
	return fmt.Sprintf(`This command shows IAM Role details.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&iamRoleShowOutput{}), ", "))
}

func (c *iamRoleShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamRoleShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	if c.Role == "" {
		return errors.New("role ID not provided")
	}

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	roles, err := client.ListIAMRoles(ctx)
	if err != nil {
		return err
	}

	role, err := roles.FindIAMRole(c.Role)
	if err != nil {
		return err
	}

	if c.Policy {
		policy := role.Policy

		out := iamPolicyOutput{
			DefaultServiceStrategy: string(policy.DefaultServiceStrategy),
			Services:               map[string]iamPolicyServiceOutput{},
		}

		for name, service := range policy.Services {
			rules := []iamPolicyServiceRuleOutput{}
			if service.Type == "rules" {
				for _, rule := range service.Rules {
					rules = append(rules, iamPolicyServiceRuleOutput{
						Action:     string(rule.Action),
						Expression: rule.Expression,
					})
				}
			}

			out.Services[name] = iamPolicyServiceOutput{
				Type:  string(service.Type),
				Rules: rules,
			}
		}

		return c.OutputFunc(&out, nil)
	}

	out := iamRoleShowOutput{
		ID:          role.ID.String(),
		Description: role.Description,
		Editable:    utils.DefaultBool(role.Editable, false),
		Labels:      role.Labels,
		Name:        role.Name,
		Permissions: role.Permissions,
	}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(iamRoleCmd, &iamRoleShowCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
