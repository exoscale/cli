package cmd

import (
	"fmt"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
	"strings"
)

type antiAffinityGroupShowOutput struct {
	ID          v3.UUID  `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Instances   []string `json:"instances"`
}

func (o *antiAffinityGroupShowOutput) ToJSON()  { output.JSON(o) }
func (o *antiAffinityGroupShowOutput) ToText()  { output.Text(o) }
func (o *antiAffinityGroupShowOutput) ToTable() { output.Table(o) }

type antiAffinityGroupShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	AntiAffinityGroup string `cli-arg:"#" cli-usage:"NAME|ID"`
}

func (c *antiAffinityGroupShowCmd) cmdAliases() []string { return gShowAlias }

func (c *antiAffinityGroupShowCmd) cmdShort() string {
	return "Show an Anti-Affinity Group details"
}

func (c *antiAffinityGroupShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows a Compute instance Anti-Affinity Group details.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&antiAffinityGroupShowOutput{}), ", "))
}

func (c *antiAffinityGroupShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *antiAffinityGroupShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := gContext

	antiAffinityGroupsResp, err := globalstate.EgoscaleV3Client.ListAntiAffinityGroups(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve list of anti-affinity %q: %w", err)
	}

	antiAffinityGroup, err := antiAffinityGroupsResp.FindAntiAffinityGroup(c.AntiAffinityGroup)
	if err != nil {
		return fmt.Errorf("unable to find anti-affinity group %q: %w", c.AntiAffinityGroup, err)
	}
	out := antiAffinityGroupShowOutput{
		ID:          antiAffinityGroup.ID,
		Name:        antiAffinityGroup.Name,
		Description: antiAffinityGroup.Description,
	}

	antiAffinityGroupWithInstanceDetails, err := globalstate.EgoscaleV3Client.GetAntiAffinityGroup(ctx, antiAffinityGroup.ID)
	if err != nil {
		return fmt.Errorf("unable to retrieve anti-affinity group with instance details %q: %w", c.AntiAffinityGroup, err)
	}
	if antiAffinityGroupWithInstanceDetails.Instances != nil {
		out.Instances = make([]string, len(antiAffinityGroupWithInstanceDetails.Instances))
		for i, instance := range antiAffinityGroupWithInstanceDetails.Instances {
			out.Instances[i] = instance.ID.String()
		}
	}
	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(antiAffinityGroupCmd, &antiAffinityGroupShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
