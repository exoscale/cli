package cmd

import (
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type apiKeyItem struct {
	Description string              `json:"description"`
	Key         string              `json:"key"`
	Operations  string              `json:"operations,omitempty"`
	Type        egoscale.APIKeyType `json:"type"`
}

type apiKeyListItemOutput []apiKeyItem

func (o *apiKeyListItemOutput) toJSON()  { outputJSON(o) }
func (o *apiKeyListItemOutput) toText()  { outputText(o) }
func (o *apiKeyListItemOutput) toTable() { outputTable(o) }

// apiKeyListCmd represents the List command
var apiKeyListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List APIKeys",
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
				Description: i.Description,
				Key:         i.Key,
				Operations:  strings.Join(i.Operations, ", "),
				Type:        i.Type,
			})
		}

		return output(&o, err)
	},
}

func init() {
	iamAPIKeyCmd.AddCommand(apiKeyListCmd)
}
