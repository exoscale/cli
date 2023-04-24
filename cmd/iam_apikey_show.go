package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/output"
	"github.com/spf13/cobra"
)

type apiKeyShowItemOutput struct {
	Name       string   `json:"name"`
	Key        string   `json:"key"`
	Operations []string `json:"operations,omitempty"`
	Resources  []string `json:"resources,omitempty"`
	Type       string   `json:"type"`
}

func (o *apiKeyShowItemOutput) toJSON()  { output.JSON(o) }
func (o *apiKeyShowItemOutput) toText()  { output.Text(o) }
func (o *apiKeyShowItemOutput) toTable() { output.Table(o) }

var apiKeyShowCmd = &cobra.Command{
	Use:   "show KEY|NAME",
	Short: "Show API key",
	Long: fmt.Sprintf(`This command shows an API key details.

Supported output template annotations: %s`,
		strings.Join(output.OutputterTemplateAnnotations(&apiKeyShowItemOutput{}), ", ")),
	Aliases: gShowAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Usage()
		}

		apiKey, err := getAPIKeyByName(args[0])
		if err != nil {
			return err
		}

		o := apiKeyShowItemOutput{
			Name:       apiKey.Name,
			Key:        apiKey.Key,
			Operations: apiKey.Operations,
			Resources:  apiKey.Resources,
			Type:       string(apiKey.Type),
		}

		return output(&o, err)
	},
}

func init() {
	apiKeyCmd.AddCommand(apiKeyShowCmd)
}
