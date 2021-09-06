package cmd

import (
	"fmt"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type securityGroupDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

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
	zone := gCurrentAccount.DefaultZone

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

	securityGroup, err := cs.FindSecurityGroup(ctx, zone, c.SecurityGroup)
	if err != nil {
		return err
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete Security Group %s?", c.SecurityGroup)) {
			return nil
		}
	}

	decorateAsyncOperation(fmt.Sprintf("Deleting Security Group %s...", c.SecurityGroup), func() {
		err = cs.DeleteSecurityGroup(ctx, zone, securityGroup)
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
