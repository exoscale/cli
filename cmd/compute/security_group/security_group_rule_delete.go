package security_group

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type securityGroupDeleteRuleCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	SecurityGroup string `cli-arg:"#" cli-usage:"SECURITY-GROUP-ID|NAME"`
	Rule          string `cli-arg:"#"`

	Force bool `cli-short:"f" cli-usage:"don't prompt for confirmation"`
}

func (c *securityGroupDeleteRuleCmd) CmdAliases() []string { return exocmd.GRemoveAlias }

func (c *securityGroupDeleteRuleCmd) CmdShort() string {
	return "Delete a Security Group rule"
}

func (c *securityGroupDeleteRuleCmd) CmdLong() string {
	return fmt.Sprintf(`This command deletes a rule from a Compute instance Security Group.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&securityGroupShowOutput{}), ", "))
}

func (c *securityGroupDeleteRuleCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *securityGroupDeleteRuleCmd) CmdRun(_ *cobra.Command, _ []string) error {
	zone := account.CurrentAccount.DefaultZone

	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

	securityGroup, err := globalstate.EgoscaleClient.FindSecurityGroup(ctx, zone, c.SecurityGroup)
	if err != nil {
		return err
	}

	if !c.Force {
		if !utils.AskQuestion(
			ctx,
			fmt.Sprintf(
				"Are you sure you want to delete rule %s from Security Group %q?",
				c.Rule,
				*securityGroup.Name,
			)) {
			return nil
		}
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Deleting Security Group rule %s...", c.Rule), func() {
		err = globalstate.EgoscaleClient.DeleteSecurityGroupRule(ctx, zone, securityGroup, &egoscale.SecurityGroupRule{ID: &c.Rule})
	})
	if err != nil {
		return err
	}

	return (&securityGroupShowCmd{
		CliCommandSettings: c.CliCommandSettings,
		SecurityGroup:      *securityGroup.ID,
	}).CmdRun(nil, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(securityGroupRuleCmd, &securityGroupDeleteRuleCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
