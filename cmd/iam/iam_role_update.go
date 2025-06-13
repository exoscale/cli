package iam

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type iamRoleUpdateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	Role string `cli-arg:"#" cli-usage:"ID|NAME"`

	Description string            `cli-flag:"description" cli-usage:"Role description"`
	Permissions []string          `cli-flag:"permissions" cli-usage:"Role permissions"`
	Labels      map[string]string `cli-flag:"label" cli-usage:"Role labels (format: key=value)"`
	Policy      string            `cli-flag:"policy" cli-usage:"Role policy (use '-' to read from STDIN)"`

	_ bool `cli-cmd:"update"`
}

func (c *iamRoleUpdateCmd) CmdAliases() []string { return nil }

func (c *iamRoleUpdateCmd) CmdShort() string {
	return "Update an IAM Role"
}

func (c *iamRoleUpdateCmd) CmdLong() string {
	return fmt.Sprintf(`This command updates an IAM Role.
When you supply '-' as a flag argument to '--policy', the new policy will be read from STDIN.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&iamPolicyOutput{}), ", "))
}

func (c *iamRoleUpdateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamRoleUpdateCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	if c.Role == "" {
		return errors.New("role not provided")
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

	updateRole := v3.UpdateIAMRoleRequest{}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Description)) {
		updateRole.Description = c.Description
	}
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Labels)) {
		updateRole.Labels = c.Labels
	}
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Permissions)) {
		updateRole.Permissions = c.Permissions
	}

	op, err := client.UpdateIAMRole(ctx, role.ID, updateRole)
	if err != nil {
		return err
	}
	utils.DecorateAsyncOperation("Updating IAM role...", func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	// If we don't need to update Policy we can exit now
	if c.Policy == "" {
		if !globalstate.Quiet {
			return (&iamRoleShowCmd{
				CliCommandSettings: c.CliCommandSettings,
				Role:               role.ID.String(),
			}).CmdRun(nil, nil)
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

	op, err = client.UpdateIAMRolePolicy(ctx, role.ID, *policy)
	if err != nil {
		return err
	}
	utils.DecorateAsyncOperation("Updating IAM role policy...", func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&iamRoleShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Role:               role.ID.String(),
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(iamRoleCmd, &iamRoleUpdateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
