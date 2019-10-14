package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type apiKeyItem struct {
	Name       string `json:"name"`
	Key        string `json:"key"`
	Operations string `json:"operations,omitempty"`
	Type       string `json:"type"`
}

type apiKeyListItemOutput []apiKeyItem

func (o *apiKeyListItemOutput) toJSON()  { outputJSON(o) }
func (o *apiKeyListItemOutput) toText()  { outputText(o) }
func (o *apiKeyListItemOutput) toTable() { outputTable(o) }

// apiKeyListCmd represents the API keys Listing command
var apiKeyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List API keys",
	Long: fmt.Sprintf(`This command lists existing API keys.
	
	Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&apiKeyListItemOutput{}), ", ")),
	Aliases: gListAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := cs.RequestWithContext(gContext, &egoscale.ListAPIKeys{})
		if err != nil {
			return err
		}

		r := resp.(*egoscale.ListAPIKeysResponse)

		o := make(apiKeyListItemOutput, 0, r.Count)
		for _, i := range r.APIKeys {
			o = append(o, apiKeyItem{
				Name:       i.Name,
				Key:        i.Key,
				Operations: strings.Join(i.Operations, ", "),
				Type:       string(i.Type),
			})
		}

		return output(&o, err)
	},
}

func init() {
	iamAPIKeyCmd.AddCommand(apiKeyListCmd)
}
