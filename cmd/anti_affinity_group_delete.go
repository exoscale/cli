package cmd

import (
	"fmt"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type antiAffinityGroupDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	AntiAffinityGroup string `cli-arg:"#" cli-usage:"ANTI-AFFINITY-GROUP-NAME|ID"`

	Force bool `cli-short:"f" cli-usage:"don't prompt for confirmation"`
}

func (c *antiAffinityGroupDeleteCmd) cmdAliases() []string { return gRemoveAlias }

func (c *antiAffinityGroupDeleteCmd) cmdShort() string {
	return "Delete an Anti-Affinity Group"
}

func (c *antiAffinityGroupDeleteCmd) cmdLong() string { return "" }

func (c *antiAffinityGroupDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *antiAffinityGroupDeleteCmd) cmdRun(_ *cobra.Command, _ []string) error {
	zone := gCurrentAccount.DefaultZone

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

	antiAffinityGroup, err := cs.FindAntiAffinityGroup(ctx, zone, c.AntiAffinityGroup)
	if err != nil {
		return err
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf(
			"Are you sure you want to delete Anti-Affinity Group %s?",
			c.AntiAffinityGroup,
		)) {
			return nil
		}
	}

	decorateAsyncOperation(fmt.Sprintf("Deleting Anti-Affinity Group %s...", c.AntiAffinityGroup), func() {
		err = cs.DeleteAntiAffinityGroup(ctx, zone, antiAffinityGroup)
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(antiAffinityGroupCmd, &antiAffinityGroupDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
