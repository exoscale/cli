package cmd

import (
	"fmt"
	"strings"

	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type securityGroupDeleteRuleCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	SecurityGroup string `cli-arg:"#" cli-usage:"SECURITY-GROUP-ID|NAME"`
	Rule          string `cli-arg:"#"`

	Force bool `cli-short:"f" cli-usage:"don't prompt for confirmation"`
}

func (c *securityGroupDeleteRuleCmd) cmdAliases() []string { return gRemoveAlias }

func (c *securityGroupDeleteRuleCmd) cmdShort() string {
	return "Delete a Security Group rule"
}

func (c *securityGroupDeleteRuleCmd) cmdLong() string {
	return fmt.Sprintf(`This command deletes a rule from a Compute instance Security Group.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&securityGroupShowOutput{}), ", "))
}

func (c *securityGroupDeleteRuleCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *securityGroupDeleteRuleCmd) cmdRun(_ *cobra.Command, _ []string) error {
	zone := gCurrentAccount.DefaultZone

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

	securityGroup, err := cs.FindSecurityGroup(ctx, zone, c.SecurityGroup)
	if err != nil {
		return err
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf(
			"Are you sure you want to delete rule %s from Security Group %q?",
			c.Rule,
			*securityGroup.Name,
		)) {
			return nil
		}
	}

	decorateAsyncOperation(fmt.Sprintf("Deleting Security Group rule %s...", c.Rule), func() {
		err = cs.DeleteSecurityGroupRule(ctx, zone, securityGroup, &egoscale.SecurityGroupRule{ID: &c.Rule})
	})
	if err != nil {
		return err
	}

	return output(showSecurityGroup(zone, *securityGroup.ID))
}

func init() {
	cobra.CheckErr(registerCLICommand(securityGroupRuleCmd, &securityGroupDeleteRuleCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
