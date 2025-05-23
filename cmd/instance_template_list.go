package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
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
	CliCommandSettings `cli-cmd:"-"`

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
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceTemplateListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	if c.Zone == "" {
		c.Zone = account.CurrentAccount.DefaultZone
	}

	ctx := exoapi.WithEndpoint(
		GContext,
		exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone),
	)

	templates, err := globalstate.EgoscaleClient.ListTemplates(
		ctx,
		c.Zone,
		egoscale.ListTemplatesWithVisibility(c.Visibility),
		egoscale.ListTemplatesWithFamily(c.Family),
	)
	if err != nil {
		return err
	}

	// Sort private templates by name for better visibility
	// Public templates are sorted by Family
	if c.Visibility == "private" {
		sort.Sort(egoscale.ByName{Templates: templates})
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

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(RegisterCLICommand(instanceTemplateCmd, &instanceTemplateListCmd{
		CliCommandSettings: DefaultCLICmdSettings(),

		Visibility: "public",
	}))
}
