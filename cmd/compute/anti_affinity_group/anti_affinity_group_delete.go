package anti_affinity_group

import (
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type antiAffinityGroupDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	AntiAffinityGroup string `cli-arg:"#" cli-usage:"ANTI-AFFINITY-GROUP-NAME|ID"`

	Force bool `cli-short:"f" cli-usage:"don't prompt for confirmation"`
}

func (c *antiAffinityGroupDeleteCmd) CmdAliases() []string { return exocmd.GRemoveAlias }

func (c *antiAffinityGroupDeleteCmd) CmdShort() string {
	return "Delete an Anti-Affinity Group"
}

func (c *antiAffinityGroupDeleteCmd) CmdLong() string { return "" }

func (c *antiAffinityGroupDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *antiAffinityGroupDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	antiAffinityGroupsResp, err := globalstate.EgoscaleV3Client.ListAntiAffinityGroups(ctx)
	if err != nil {
		return err
	}

	antiAffinityGroup, err := antiAffinityGroupsResp.FindAntiAffinityGroup(c.AntiAffinityGroup)
	if err != nil {
		return err
	}

	if !c.Force {
		if !utils.AskQuestion(ctx, fmt.Sprintf(
			"Are you sure you want to delete Anti-Affinity Group %s?",
			c.AntiAffinityGroup,
		)) {
			return nil
		}
	}

	return utils.DecorateAsyncOperations(fmt.Sprintf("Deleting Anti-Affinity Group %s...", c.AntiAffinityGroup), func() error {
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
	cobra.CheckErr(exocmd.RegisterCLICommand(antiAffinityGroupCmd, &antiAffinityGroupDeleteCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
