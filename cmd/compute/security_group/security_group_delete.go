package security_group

import (
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type securityGroupDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	DeleteRules   bool   `cli-short:"r" cli-usage:"delete rules before deleting the Security Group"`
	SecurityGroup string `cli-arg:"#" cli-usage:"SECURITY-GROUP-NAME|ID"`

	Force bool `cli-short:"f" cli-usage:"don't prompt for confirmation"`
}

func (c *securityGroupDeleteCmd) CmdAliases() []string { return exocmd.GRemoveAlias }

func (c *securityGroupDeleteCmd) CmdShort() string {
	return "Delete a Security Group"
}

func (c *securityGroupDeleteCmd) CmdLong() string { return "" }

func (c *securityGroupDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *securityGroupDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	zone := account.CurrentAccount.DefaultZone

	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

	securityGroup, err := globalstate.EgoscaleClient.FindSecurityGroup(ctx, zone, c.SecurityGroup)
	if err != nil {
		return err
	}

	if !c.Force {
		if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to delete Security Group %s?", c.SecurityGroup)) {
			return nil
		}
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Deleting Security Group %s...", c.SecurityGroup), func() {
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
	cobra.CheckErr(exocmd.RegisterCLICommand(securityGroupCmd, &securityGroupDeleteCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
