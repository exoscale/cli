package security_group

import (
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type securityGroupDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	DeleteRules   bool   `cli-short:"r" cli-usage:"Delete all rules but not the security group"`
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
	ctx := exocmd.GContext
	var err error
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	securityGroups, err := client.ListSecurityGroups(ctx)
	if err != nil {
		return err
	}

	securityGroup, err := securityGroups.FindSecurityGroup(c.SecurityGroup)
	if err != nil {
		return err
	}

	if !c.DeleteRules {
		if !c.Force {
			if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to delete Security Group %s?", c.SecurityGroup)) {
				return nil
			}
		}

		op, err := client.DeleteSecurityGroup(ctx, securityGroup.ID)
		if err != nil {
			return err
		}

		utils.DecorateAsyncOperation(fmt.Sprintf("Deleting Security Group %s...", c.SecurityGroup), func() {
			_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		})
	} else {
		if !c.Force {
			if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to delete the rules associated with Security Group %s?", c.SecurityGroup)) {
				return nil
			}
		}

		utils.DecorateAsyncOperation(fmt.Sprintf("Deleting Rules for Security Group %s...", c.SecurityGroup), func() {
			for _, rule := range securityGroup.Rules {
				op, err := client.DeleteRuleFromSecurityGroup(ctx, securityGroup.ID, rule.ID)
				if err != nil {
					return
				}
				_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
				if err != nil {
					return
				}
			}
		})

	}

	return err
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(securityGroupCmd, &securityGroupDeleteCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
