package cmd

import (
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type apiKeyCreateItemOutput egoscale.APIKey

func (o *apiKeyCreateItemOutput) toJSON()  { outputJSON(o) }
func (o *apiKeyCreateItemOutput) toText()  { outputText(o) }
func (o *apiKeyCreateItemOutput) toTable() { outputTable(o) }

// apiKeyCreateCmd represents an API key creation command
var apiKeyCreateCmd = &cobra.Command{
	Use:     "create <description>",
	Short:   "Create an API key",
	Aliases: gCreateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Usage()
		}

		ops, err := cmd.Flags().GetStringSlice("operations")
		if err != nil {
			return err
		}

		resp, err := cs.RequestWithContext(gContext, &egoscale.CreateAPIKey{
			Description: args[0],
			Operations:  strings.Join(ops, ","),
		})
		if err != nil {
			return err
		}

		apiKey := resp.(*egoscale.APIKey)
		o := apiKeyCreateItemOutput(*apiKey)
		return output(&o, err)
	},
}

func init() {
	apiKeyCreateCmd.Flags().StringSliceP("operations", "o", []string{}, "API key operation")
	iamAPIKeyCmd.AddCommand(apiKeyCreateCmd)
}
