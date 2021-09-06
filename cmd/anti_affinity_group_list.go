package cmd

import (
	"fmt"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type antiAffinityGroupListItemOutput struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type antiAffinityGroupListOutput []antiAffinityGroupListItemOutput

func (o *antiAffinityGroupListOutput) toJSON()  { outputJSON(o) }
func (o *antiAffinityGroupListOutput) toText()  { outputText(o) }
func (o *antiAffinityGroupListOutput) toTable() { outputTable(o) }

type antiAffinityGroupListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`
}

func (c *antiAffinityGroupListCmd) cmdAliases() []string { return gListAlias }

func (c *antiAffinityGroupListCmd) cmdShort() string { return "List Anti-Affinity Groups" }

func (c *antiAffinityGroupListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists Compute instance Anti-Affinity Groups.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&antiAffinityGroupListItemOutput{}), ", "))
}

func (c *antiAffinityGroupListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *antiAffinityGroupListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(gCurrentAccount.Environment, gCurrentAccount.DefaultZone),
	)

	antiAffinityGroups, err := cs.ListAntiAffinityGroups(ctx, gCurrentAccount.DefaultZone)
	if err != nil {
		return err
	}

	out := make(antiAffinityGroupListOutput, 0)

	for _, t := range antiAffinityGroups {
		out = append(out, antiAffinityGroupListItemOutput{
			ID:   *t.ID,
			Name: *t.Name,
		})
	}

	return output(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(antiAffinityGroupCmd, &antiAffinityGroupListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
