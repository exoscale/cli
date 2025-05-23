package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type antiAffinityGroupCreateCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#"`

	Description string `cli-usage:"Anti-Affinity Group description"`
}

func (c *antiAffinityGroupCreateCmd) CmdAliases() []string { return GCreateAlias }

func (c *antiAffinityGroupCreateCmd) CmdShort() string {
	return "Create an Anti-Affinity Group"
}

func (c *antiAffinityGroupCreateCmd) CmdLong() string {
	return fmt.Sprintf(`This command creates a Compute instance Anti-Affinity Group.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&antiAffinityGroupShowOutput{}), ", "))
}

func (c *antiAffinityGroupCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *antiAffinityGroupCreateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := GContext

	antiAffinityGroupCreateRequest := v3.CreateAntiAffinityGroupRequest{
		Description: c.Description,
		Name:        c.Name,
	}

	err := decorateAsyncOperations(fmt.Sprintf("Creating Anti-Affinity Group %q...", c.Name), func() error {
		op, err := globalstate.EgoscaleV3Client.CreateAntiAffinityGroup(ctx, antiAffinityGroupCreateRequest)
		if err != nil {
			return fmt.Errorf("exoscale: error while creating Anti Affinity Group: %w", err)
		}

		_, err = globalstate.EgoscaleV3Client.Wait(ctx, op, v3.OperationStateSuccess)
		if err != nil {
			return fmt.Errorf("exoscale: error while waiting for Anti Affinity Group creation: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	antiAffinityGroupsResp, err := globalstate.EgoscaleV3Client.ListAntiAffinityGroups(ctx)
	if err != nil {
		return err
	}

	antiAffinityGroup, err := antiAffinityGroupsResp.FindAntiAffinityGroup(c.Name)
	if err != nil {
		return err
	}

	return (&antiAffinityGroupShowCmd{
		CliCommandSettings: c.CliCommandSettings,
		AntiAffinityGroup:  antiAffinityGroup.ID.String(),
	}).CmdRun(nil, nil)
}

func init() {
	cobra.CheckErr(RegisterCLICommand(antiAffinityGroupCmd, &antiAffinityGroupCreateCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
