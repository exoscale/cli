package cmd

import (
	"fmt"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instanceTypeListItemOutput struct {
	ID     string `json:"id"`
	Family string `json:"family"`
	Size   string `json:"name"`
}

type instanceTypeListOutput []instanceTypeListItemOutput

func (o *instanceTypeListOutput) toJSON()  { outputJSON(o) }
func (o *instanceTypeListOutput) toText()  { outputText(o) }
func (o *instanceTypeListOutput) toTable() { outputTable(o) }

type instanceTypeListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`
}

func (c *instanceTypeListCmd) cmdAliases() []string { return nil }

func (c *instanceTypeListCmd) cmdShort() string { return "List Compute instance types" }

func (c *instanceTypeListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists available Compute instance types.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&instanceTypeListItemOutput{}), ", "))
}

func (c *instanceTypeListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceTypeListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(gCurrentAccount.Environment, gCurrentAccount.DefaultZone),
	)

	instanceTypes, err := cs.ListInstanceTypes(ctx, gCurrentAccount.DefaultZone)
	if err != nil {
		return err
	}

	out := make(instanceTypeListOutput, 0)

	for _, t := range instanceTypes {
		out = append(out, instanceTypeListItemOutput{
			ID:     *t.ID,
			Family: *t.Family,
			Size:   *t.Size,
		})
	}

	return output(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceTypeCmd, &instanceTypeListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
