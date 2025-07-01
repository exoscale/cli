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
	v3 "github.com/exoscale/egoscale/v3"
)

type securityGroupDeleteRuleCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	SecurityGroup string  `cli-arg:"#" cli-usage:"SECURITY-GROUP-ID|NAME"`
	Rule          v3.UUID `cli-arg:"#"`

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
	ctx := exocmd.GContext
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

	var rule *v3.SecurityGroupRule
	for _, r := range securityGroup.Rules {
		if r.ID == v3.UUID(c.Rule) {
			rule = &r
		}
	}
	if rule == nil {
		return fmt.Errorf("could not find rule %q in security group %s", c.Rule, c.SecurityGroup)
	}

	if !c.Force {
		if !utils.AskQuestion(ctx, fmt.Sprintf(
			"Are you sure you want to delete rule %s from Security Group %q?",
			c.Rule,
			securityGroup.Name,
		)) {
			return nil
		}
	}

	op, err := client.DeleteRuleFromSecurityGroup(ctx, securityGroup.ID, c.Rule)
	utils.DecorateAsyncOperation(fmt.Sprintf("Deleting Security Group rule %s...", c.Rule), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	return (&securityGroupShowCmd{
		CliCommandSettings: c.CliCommandSettings,
		SecurityGroup:      securityGroup.ID.String(),
	}).CmdRun(nil, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(securityGroupRuleCmd, &securityGroupDeleteRuleCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
