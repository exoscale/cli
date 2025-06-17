package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
)

type configListItemOutput struct {
	Name    string `json:"name"`
	Default bool   `json:"default"`
}

type configListOutput []configListItemOutput

func (o *configListOutput) ToJSON() { output.JSON(o) }

func (o *configListOutput) ToText() { output.Text(o) }

func (o *configListOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Accounts"})

	for _, i := range *o {
		a := i.Name
		if i.Default {
			a += "*"
		}

		t.Append([]string{a})
	}

	t.Render()
}

func init() {
	configCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List available accounts",
		Long: fmt.Sprintf(`This command lists configured Exoscale accounts. The default account is marked with (*).

Supported output template annotations: %s`,
			strings.Join(output.TemplateAnnotations(&configListOutput{}), ", ")),
		Aliases: exocmd.GListAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			return utils.PrintOutput(listConfigs(), nil)
		},
	})
}

func listConfigs() output.Outputter {
	out := configListOutput{}

	if account.GAllAccount == nil {
		return &out
	}

	for _, a := range account.GAllAccount.Accounts {
		out = append(out, configListItemOutput{
			Name:    a.Name,
			Default: a.IsDefault(),
		})
	}

	return &out
}
