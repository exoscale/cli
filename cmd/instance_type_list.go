package cmd

import (
	"fmt"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type computeInstanceTypeListItemOutput struct {
	ID     string `json:"id"`
	Family string `json:"family"`
	Size   string `json:"name"`
}

type computeInstanceTypeListOutput []computeInstanceTypeListItemOutput

func (o *computeInstanceTypeListOutput) toJSON()  { outputJSON(o) }
func (o *computeInstanceTypeListOutput) toText()  { outputText(o) }
func (o *computeInstanceTypeListOutput) toTable() { outputTable(o) }

type computeInstanceTypeListCmd struct {
	_ bool `cli-cmd:"list"`
}

func (c *computeInstanceTypeListCmd) cmdAliases() []string { return nil }

func (c *computeInstanceTypeListCmd) cmdShort() string { return "List Compute instance types" }

func (c *computeInstanceTypeListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists available Compute instance types.

	Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&computeInstanceTypeListItemOutput{}), ", "))
}

func (c *computeInstanceTypeListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *computeInstanceTypeListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(gCurrentAccount.Environment, gCurrentAccount.DefaultZone),
	)

	computeInstanceTypes, err := cs.ListInstanceTypes(ctx, gCurrentAccount.DefaultZone)
	if err != nil {
		return err
	}

	out := make(computeInstanceTypeListOutput, 0)

	for _, t := range computeInstanceTypes {
		out = append(out, computeInstanceTypeListItemOutput{
			ID:     *t.ID,
			Family: *t.Family,
			Size:   *t.Size,
		})
	}

	return output(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(computeInstanceTypeCmd, &computeInstanceTypeListCmd{}))
}
