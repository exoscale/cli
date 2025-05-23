package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type securityGroupDeleteCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	DeleteRules   bool   `cli-short:"r" cli-usage:"delete rules before deleting the Security Group"`
	SecurityGroup string `cli-arg:"#" cli-usage:"SECURITY-GROUP-NAME|ID"`

	Force bool `cli-short:"f" cli-usage:"don't prompt for confirmation"`
}

func (c *securityGroupDeleteCmd) CmdAliases() []string { return GRemoveAlias }

func (c *securityGroupDeleteCmd) CmdShort() string {
	return "Delete a Security Group"
}

func (c *securityGroupDeleteCmd) CmdLong() string { return "" }

func (c *securityGroupDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *securityGroupDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	zone := account.CurrentAccount.DefaultZone

	ctx := exoapi.WithEndpoint(GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

	securityGroup, err := globalstate.EgoscaleClient.FindSecurityGroup(ctx, zone, c.SecurityGroup)
	if err != nil {
		return err
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete Security Group %s?", c.SecurityGroup)) {
			return nil
		}
	}

	decorateAsyncOperation(fmt.Sprintf("Deleting Security Group %s...", c.SecurityGroup), func() {
		if c.DeleteRules {
			for _, rule := range securityGroup.Rules {
				if err = globalstate.EgoscaleClient.DeleteSecurityGroupRule(ctx, zone, securityGroup, rule); err != nil {
					return
				}
			}
		}

		err = globalstate.EgoscaleClient.DeleteSecurityGroup(ctx, zone, securityGroup)
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(RegisterCLICommand(securityGroupCmd, &securityGroupDeleteCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
