package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceTemplateListItemOutput struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Family       string `json:"family"`
	CreationDate string `json:"creation_date"`
}

type instanceTemplateListOutput []instanceTemplateListItemOutput

func (o *instanceTemplateListOutput) ToJSON()  { output.JSON(o) }
func (o *instanceTemplateListOutput) ToText()  { output.Text(o) }
func (o *instanceTemplateListOutput) ToTable() { output.Table(o) }

type instanceTemplateListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Family     string `cli-short:"f" cli-usage:"template family to filter results to"`
	Visibility string `cli-short:"v" cli-usage:"template visibility (public|private)"`
	Zone       string `cli-short:"z" cli-usage:"zone to filter results to (default: current account's default zone)"`
}

func (c *instanceTemplateListCmd) CmdAliases() []string { return nil }

func (c *instanceTemplateListCmd) CmdShort() string { return "List Compute instance templates" }

func (c *instanceTemplateListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists available Compute instance templates.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instanceTemplateListItemOutput{}), ", "))
}

func (c *instanceTemplateListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceTemplateListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	if c.Zone == "" {
		c.Zone = account.CurrentAccount.DefaultZone
	}

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	opts := []v3.ListTemplatesOpt{v3.ListTemplatesWithVisibility(v3.ListTemplatesVisibility(c.Visibility))}
	if c.Family != "" {
		opts = append(opts, v3.ListTemplatesWithFamily(c.Family))
	}
	templates, err := client.ListTemplates(
		ctx,
		opts...,
	)
	if err != nil {
		return err
	}

	// Sort private templates by name for better visibility
	// Public templates are sorted by Family
	if c.Visibility == "private" {
		slices.SortFunc(templates.Templates, func(i, j v3.Template) int {
			return strings.Compare(i.Name, j.Name)

		})
	} else {
		slices.SortFunc(templates.Templates, func(i, j v3.Template) int {
			return strings.Compare(i.Family, j.Family)

		})
	}

	out := make(instanceTemplateListOutput, 0)

	for _, t := range templates.Templates {
		out = append(out, instanceTemplateListItemOutput{
			ID:           t.ID.String(),
			Name:         t.Name,
			Family:       t.Family,
			CreationDate: t.CreatedAT.String(),
		})
	}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceTemplateCmd, &instanceTemplateListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),

		Visibility: "public",
	}))
}
