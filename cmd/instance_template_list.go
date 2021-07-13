package cmd

import (
	"fmt"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type computeInstanceTemplateListItemOutput struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Family       string `json:"family"`
	CreationDate string `json:"creation_date"`
}

type computeInstanceTemplateListOutput []computeInstanceTemplateListItemOutput

func (o *computeInstanceTemplateListOutput) toJSON()  { outputJSON(o) }
func (o *computeInstanceTemplateListOutput) toText()  { outputText(o) }
func (o *computeInstanceTemplateListOutput) toTable() { outputTable(o) }

type computeInstanceTemplateListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Family     string `cli-short:"f" cli-usage:"template family to filter results to"`
	Visibility string `cli-short:"v" cli-usage:"template visibility (public|private)"`
	Zone       string `cli-short:"z" cli-usage:"zone to filter results to (default: current account's default zone)"`
}

func (c *computeInstanceTemplateListCmd) cmdAliases() []string { return nil }

func (c *computeInstanceTemplateListCmd) cmdShort() string { return "List Compute instance templates" }

func (c *computeInstanceTemplateListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists available Compute instance templates.

	Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&computeInstanceTemplateListItemOutput{}), ", "))
}

func (c *computeInstanceTemplateListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *computeInstanceTemplateListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(gCurrentAccount.Environment, gCurrentAccount.DefaultZone),
	)

	templates, err := cs.ListTemplates(ctx, gCurrentAccount.DefaultZone, c.Visibility, c.Family)
	if err != nil {
		return err
	}

	out := make(computeInstanceTemplateListOutput, 0)

	for _, t := range templates {
		out = append(out, computeInstanceTemplateListItemOutput{
			ID:           *t.ID,
			Name:         *t.Name,
			Family:       *t.Family,
			CreationDate: t.CreatedAt.String(),
		})
	}

	return output(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(computeInstanceTemplateCmd, &computeInstanceTemplateListCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		Visibility: "public",
	}))
}
