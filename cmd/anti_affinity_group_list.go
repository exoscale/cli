package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type antiAffinityGroupListItemOutput struct {
	ID   v3.UUID `json:"id"`
	Name string  `json:"name"`
}

type antiAffinityGroupListOutput []antiAffinityGroupListItemOutput

func (o *antiAffinityGroupListOutput) ToJSON()  { output.JSON(o) }
func (o *antiAffinityGroupListOutput) ToText()  { output.Text(o) }
func (o *antiAffinityGroupListOutput) ToTable() { output.Table(o) }

type antiAffinityGroupListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`
}

func (c *antiAffinityGroupListCmd) cmdAliases() []string { return gListAlias }

func (c *antiAffinityGroupListCmd) cmdShort() string { return "List Anti-Affinity Groups" }

func (c *antiAffinityGroupListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists Compute instance Anti-Affinity Groups.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&antiAffinityGroupListItemOutput{}), ", "))
}

func (c *antiAffinityGroupListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *antiAffinityGroupListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := gContext

	antiAffinityGroups, err := globalstate.EgoscaleV3Client.ListAntiAffinityGroups(ctx)
	if err != nil {
		return err
	}

	out := make(antiAffinityGroupListOutput, 0)

	for _, t := range antiAffinityGroups.AntiAffinityGroups {
		out = append(out, antiAffinityGroupListItemOutput{
			ID:   t.ID,
			Name: t.Name,
		})
	}

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(antiAffinityGroupCmd, &antiAffinityGroupListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
