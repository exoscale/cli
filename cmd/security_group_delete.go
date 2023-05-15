package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type securityGroupDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	DeleteRules   bool   `cli-short:"r" cli-usage:"delete rules before deleting the Security Group"`
	SecurityGroup string `cli-arg:"#" cli-usage:"SECURITY-GROUP-NAME|ID"`

	Force bool `cli-short:"f" cli-usage:"don't prompt for confirmation"`
}

func (c *securityGroupDeleteCmd) cmdAliases() []string { return gRemoveAlias }

func (c *securityGroupDeleteCmd) cmdShort() string {
	return "Delete a Security Group"
}

func (c *securityGroupDeleteCmd) cmdLong() string { return "" }

func (c *securityGroupDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *securityGroupDeleteCmd) cmdRun(_ *cobra.Command, _ []string) error {
	zone := account.CurrentAccount.DefaultZone

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

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
	cobra.CheckErr(registerCLICommand(securityGroupCmd, &securityGroupDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
