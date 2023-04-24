package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/spf13/cobra"
)

type configListItemOutput struct {
	Name    string `json:"name"`
	Default bool   `json:"default"`
}

type configListOutput []configListItemOutput

func (o *configListOutput) toJSON() { output.JSON(o) }

func (o *configListOutput) toText() { output.Text(o) }

func (o *configListOutput) toTable() {
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
			strings.Join(outputterTemplateAnnotations(&configListOutput{}), ", ")),
		Aliases: gListAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			return output(listConfigs(), nil)
		},
	})
}

func listConfigs() outputter {
	out := configListOutput{}

	if gAllAccount == nil {
		return &out
	}

	for _, a := range gAllAccount.Accounts {
		out = append(out, configListItemOutput{
			Name:    a.Name,
			Default: a.IsDefault(),
		})
	}

	return &out
}
