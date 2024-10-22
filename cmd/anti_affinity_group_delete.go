package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
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
	ctx := gContext

	antiAffinityGroupsResp, err := globalstate.EgoscaleV3Client.ListAntiAffinityGroups(ctx)
	if err != nil {
		return err
	}

	antiAffinityGroup, err := antiAffinityGroupsResp.FindAntiAffinityGroup(c.AntiAffinityGroup)
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

	return decorateAsyncOperations(fmt.Sprintf("Deleting Anti-Affinity Group %s...", c.AntiAffinityGroup), func() error {
		op, err := globalstate.EgoscaleV3Client.DeleteAntiAffinityGroup(ctx, antiAffinityGroup.ID)
		if err != nil {
			return fmt.Errorf("exoscale: error while deleting Anti Affinity Group: %w", err)
		}

		_, err = globalstate.EgoscaleV3Client.Wait(ctx, op, v3.OperationStateSuccess)
		if err != nil {
			return fmt.Errorf("exoscale: error while waiting for Anti Affinity Group deletion: %w", err)
		}

		return nil
	})
}

func init() {
	cobra.CheckErr(registerCLICommand(antiAffinityGroupCmd, &antiAffinityGroupDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
