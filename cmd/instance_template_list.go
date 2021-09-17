package cmd

import (
	"fmt"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instanceTemplateListItemOutput struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Family       string `json:"family"`
	CreationDate string `json:"creation_date"`
}

type instanceTemplateListOutput []instanceTemplateListItemOutput

func (o *instanceTemplateListOutput) toJSON()  { outputJSON(o) }
func (o *instanceTemplateListOutput) toText()  { outputText(o) }
func (o *instanceTemplateListOutput) toTable() { outputTable(o) }

type instanceTemplateListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Family     string `cli-short:"f" cli-usage:"template family to filter results to"`
	Visibility string `cli-short:"v" cli-usage:"template visibility (public|private)"`
	Zone       string `cli-short:"z" cli-usage:"zone to filter results to (default: current account's default zone)"`
}

func (c *instanceTemplateListCmd) cmdAliases() []string { return nil }

func (c *instanceTemplateListCmd) cmdShort() string { return "List Compute instance templates" }

func (c *instanceTemplateListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists available Compute instance templates.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&instanceTemplateListItemOutput{}), ", "))
}

func (c *instanceTemplateListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceTemplateListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	if c.Zone == "" {
		c.Zone = gCurrentAccount.DefaultZone
	}

	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone),
	)

	templates, err := cs.ListTemplates(ctx, c.Zone, c.Visibility, c.Family)
	if err != nil {
		return err
	}

	out := make(instanceTemplateListOutput, 0)

	for _, t := range templates {
		out = append(out, instanceTemplateListItemOutput{
			ID:           *t.ID,
			Name:         *t.Name,
			Family:       *t.Family,
			CreationDate: t.CreatedAt.String(),
		})
	}

	return output(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceTemplateCmd, &instanceTemplateListCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		Visibility: "public",
	}))
}
