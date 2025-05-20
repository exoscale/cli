package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

type iamRoleDeleteCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Role string `cli-arg:"#" cli-usage:"ID|NAME"`

	Force bool `cli-short:"f" cli-usage:"don't prompt for confirmation"`
}

func (c *iamRoleDeleteCmd) CmdAliases() []string { return GDeleteAlias }

func (c *iamRoleDeleteCmd) CmdShort() string {
	return "Delete IAM Role"
}

func (c *iamRoleDeleteCmd) CmdLong() string {
	return `This command deletes an existing IAM Role.
It will fail if the Role is attached to an IAM Key.`
}

func (c *iamRoleDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamRoleDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := GContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
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

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete IAM Role %s?", role.ID.String())) {
			return nil
		}
	}

	op, err := client.DeleteIAMRole(ctx, role.ID)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Deleting IAM role %s...", role.ID.String()), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	return err
}

func init() {
	cobra.CheckErr(RegisterCLICommand(iamRoleCmd, &iamRoleDeleteCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
