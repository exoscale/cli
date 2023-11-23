package cmd

import (
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	exoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type iamOrgPolicyResetCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	Force bool `cli-short:"f" cli-usage:"don't prompt for confirmation"`

	_ bool `cli-cmd:"reset"`
}

func (c *iamOrgPolicyResetCmd) cmdAliases() []string { return nil }

func (c *iamOrgPolicyResetCmd) cmdShort() string {
	return "Reset Org policy to default"
}

func (c *iamOrgPolicyResetCmd) cmdLong() string {
	return `This command resets the IAM Organization Policy to the default (allow all).
This will remove any constraints that were set in the Org Policy.`
}

func (c *iamOrgPolicyResetCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamOrgPolicyResetCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	if !c.Force {
		if !askQuestion("This action will remove any resource constraints you may have set in your Org Policy. Proceed?") {
			return nil
		}
	}

	zone := account.CurrentAccount.DefaultZone
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

	policy := &exoscale.IAMPolicy{
		DefaultServiceStrategy: "allow",
		Services:               map[string]exoscale.IAMPolicyService{},
	}

	return globalstate.EgoscaleClient.UpdateIAMOrgPolicy(ctx, zone, policy)
}

func init() {
	cobra.CheckErr(registerCLICommand(iamOrgPolicyCmd, &iamOrgPolicyResetCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
