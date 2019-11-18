package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type apiKeyCreateItemOutput struct {
	Name       string   `json:"name"`
	Key        string   `json:"key"`
	Secret     string   `json:"secret,omitempty"`
	Operations []string `json:"operations,omitempty"`
	Type       string   `json:"type"`
}

func (o *apiKeyCreateItemOutput) toJSON()  { outputJSON(o) }
func (o *apiKeyCreateItemOutput) toText()  { outputText(o) }
func (o *apiKeyCreateItemOutput) toTable() { outputTable(o) }

// apiKeyCreateCmd represents an API key creation command
var apiKeyCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create an API key",
	Long: fmt.Sprintf(`This command create an API key.

	Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&apiKeyCreateItemOutput{}), ", ")),
	Aliases: gCreateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Usage()
		}

		ops, err := cmd.Flags().GetStringSlice("operation")
		if err != nil {
			return err
		}

		resp, err := cs.RequestWithContext(gContext, &egoscale.CreateAPIKey{
			Name:       args[0],
			Operations: strings.Join(ops, ","),
		})
		if err != nil {
			return err
		}

		apiKey := resp.(*egoscale.APIKey)

		if !gQuiet {
			o := apiKeyCreateItemOutput{
				Name:       apiKey.Name,
				Key:        apiKey.Key,
				Secret:     apiKey.Secret,
				Operations: apiKey.Operations,
				Type:       string(apiKey.Type),
			}

			if err := output(&o, err); err != nil {
				return err
			}
		}

		fmt.Print(`
/!\  Ensure to save your API Secret somewhere,   /!\
/!\ as there is no way to recover it afterwards. /!\

`)

		return nil
	},
}

func init() {
	apiKeyCreateCmd.Flags().StringSliceP("operation", "o", []string{}, "API key allowed operation")
	apiKeyCmd.AddCommand(apiKeyCreateCmd)
}
