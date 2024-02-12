package cmd

import (
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

type iamRoleUpdateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	Role string `cli-arg:"#" cli-usage:"ID|NAME"`

	Description string            `cli-flag:"description" cli-usage:"Role description"`
	Permissions []string          `cli-flag:"permissions" cli-usage:"Role permissions"`
	Labels      map[string]string `cli-flag:"label" cli-usage:"Role labels (format: key=value)"`
	Policy      string            `cli-flag:"policy" cli-usage:"Role policy (use '-' to read from STDIN)"`

	_ bool `cli-cmd:"update"`
}

func (c *iamRoleUpdateCmd) cmdAliases() []string { return nil }

func (c *iamRoleUpdateCmd) cmdShort() string {
	return "Update an IAM Role"
}

func (c *iamRoleUpdateCmd) cmdLong() string {
	return fmt.Sprintf(`This command updates an IAM Role.
When you supply '-' as a flag argument to '--policy', the new policy will be read from STDIN.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&iamPolicyOutput{}), ", "))
}

func (c *iamRoleUpdateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamRoleUpdateCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	if c.Role == "" {
		return errors.New("Role not provided")
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

	role := &exoscale.IAMRole{
		ID: &c.Role,
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Description)) {
		role.Description = &c.Description
	}
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Labels)) {
		role.Labels = c.Labels
	}
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Permissions)) {
		role.Permissions = c.Permissions
	}

	err := globalstate.EgoscaleClient.UpdateIAMRole(ctx, zone, role)
	if err != nil {
		return err
	}

	// If we don't need to update Policy we can exit now
	if c.Policy == "" {
		if !globalstate.Quiet {
			return (&iamRoleShowCmd{
				cliCommandSettings: c.cliCommandSettings,
				Role:               *role.ID,
			}).cmdRun(nil, nil)
		}

		return nil
	}

	if c.Policy == "-" {
		inputReader := cmd.InOrStdin()
		b, err := io.ReadAll(inputReader)
		if err != nil {
			return fmt.Errorf("failed to read policy from stdin: %w", err)
		}

		c.Policy = string(b)
	}

	policy, err := iamPolicyFromJSON([]byte(c.Policy))
	if err != nil {
		return fmt.Errorf("failed to parse IAM policy: %w", err)
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
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(iamRoleCmd, &iamRoleUpdateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
