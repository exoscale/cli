package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type apiKeyItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type apiKeyListItemOutput []apiKeyItem

func (o *apiKeyListItemOutput) ToJSON()  { output.JSON(o) }
func (o *apiKeyListItemOutput) ToText()  { output.Text(o) }
func (o *apiKeyListItemOutput) ToTable() { output.Table(o) }

var apiKeyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List API keys",
	Long: fmt.Sprintf(`This command lists existing API keys.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&apiKeyListItemOutput{}), ", ")),
	Aliases: gListAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := globalstate.EgoscaleClient.RequestWithContext(gContext, &egoscale.ListAPIKeys{})
		if err != nil {
			return err
		}

		r := resp.(*egoscale.ListAPIKeysResponse)

		o := make(apiKeyListItemOutput, 0, r.Count)
		for _, i := range r.APIKeys {
			o = append(o, apiKeyItem{
				Name: i.Name,
				Key:  i.Key,
				Type: string(i.Type),
			})
		}

		return printOutput(&o, err)
	},
}

func init() {
	apiKeyCmd.AddCommand(apiKeyListCmd)
}
