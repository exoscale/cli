package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type configShowOutput struct {
	Name               string `json:"name"`
	APIKey             string `json:"api_key"`
	APISecret          string `json:"api_secret"`
	DefaultZone        string `json:"default_zone"`
	DefaultTemplate    string `json:"default_template,omitempty"`
	ComputeAPIEndpoint string `json:"compute_api_endpoint,omitempty"`
	StorageAPIEndpoint string `json:"storage_api_endpoint,omitempty"`
	DNSAPIEndpoint     string `json:"dns_api_endpoint,omitempty" outputLabel:"DNS API Endpoint"`
}

func (o *configShowOutput) Type() string { return "Account" }
func (o *configShowOutput) toJSON()      { outputJSON(o) }
func (o *configShowOutput) toText()      { outputText(o) }
func (o *configShowOutput) toTable()     { outputTable(o) }

func init() {
	configCmd.AddCommand(&cobra.Command{
		Use:   "show <account name>",
		Short: "Show an account details",
		Long: fmt.Sprintf(`This command shows an Exoscale account details.

Supported output template annotations: %s`,
			strings.Join(outputterTemplateAnnotations(&configShowOutput{}), ", ")),
		Aliases: gShowAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			if gAllAccount == nil {
				return fmt.Errorf("no accounts are defined")
			}
			name := gCurrentAccount.AccountName()

			if len(args) > 0 {
				name = args[0]
			}

			return output(showConfig(name))
		},
	})
}

func showConfig(name string) (outputter, error) {
	account := getAccountByName(name)
	if account == nil {
		return nil, fmt.Errorf("account %q was not found", name)
	}

	secret := strings.Repeat("×", len(account.Key))
	if len(account.SecretCommand) > 0 {
		secret = strings.Join(account.SecretCommand, " ")
	}

	out := configShowOutput{
		Name:               account.Name,
		APIKey:             account.Key,
		APISecret:          secret,
		DefaultZone:        account.DefaultZone,
		DefaultTemplate:    account.DefaultTemplate,
		ComputeAPIEndpoint: account.Endpoint,
		StorageAPIEndpoint: account.SosEndpoint,
		DNSAPIEndpoint:     account.DNSEndpoint,
	}

	return &out, nil
}
