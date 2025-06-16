package iam

import (
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type iamOrgPolicyResetCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	Force bool `cli-short:"f" cli-usage:"don't prompt for confirmation"`

	_ bool `cli-cmd:"reset"`
}

func (c *iamOrgPolicyResetCmd) CmdAliases() []string { return nil }

func (c *iamOrgPolicyResetCmd) CmdShort() string {
	return "Reset Org policy to default"
}

func (c *iamOrgPolicyResetCmd) CmdLong() string {
	return `This command resets the IAM Organization Policy to the default (allow all).
This will remove any constraints that were set in the Org Policy.`
}

func (c *iamOrgPolicyResetCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamOrgPolicyResetCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	if !c.Force {
		if !utils.AskQuestion(ctx, "This action will reset your Org Policy to the default, removing any constraints that were set in the Org Policy. Proceed?") {
			return nil
		}
	}

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	policy := &v3.IAMPolicy{
		DefaultServiceStrategy: "allow",
		Services:               map[string]v3.IAMServicePolicy{},
	}

	op, err := client.UpdateIAMOrganizationPolicy(ctx, *policy)
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation("Resetting IAM org policy...", func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	return err
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(iamOrgPolicyCmd, &iamOrgPolicyResetCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
