package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/output"
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
	ConfigFile         string `json:"config_file" outputLabel:"Configuration File"`
	ClientTimeout      int    `json:"client_timeout" outputLabel:"API Timeout (in minutes)"`
}

func (o *configShowOutput) Type() string { return "Account" }
func (o *configShowOutput) ToJSON()      { output.JSON(o) }
func (o *configShowOutput) ToText()      { output.Text(o) }
func (o *configShowOutput) ToTable()     { output.Table(o) }

func init() {
	configCmd.AddCommand(&cobra.Command{
		Use:   "show NAME",
		Short: "Show an account details",
		Long: fmt.Sprintf(`This command shows an Exoscale account details.

Supported output template annotations: %s`,
			strings.Join(output.OutputterTemplateAnnotations(&configShowOutput{}), ", ")),
		Aliases: gShowAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			if account.GAllAccount == nil {
				return fmt.Errorf("no accounts configured")
			}
			name := gCurrentAccount.AccountName(gContext)

			if len(args) > 0 {
				name = args[0]
			}

			return printOutput(showConfig(name))
		},
	})
}

func showConfig(name string) (output.Outputter, error) {
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
		ConfigFile:         gConfigFilePath,
		APIKey:             account.Key,
		APISecret:          secret,
		DefaultZone:        account.DefaultZone,
		DefaultTemplate:    account.DefaultTemplate,
		ComputeAPIEndpoint: account.Endpoint,
		StorageAPIEndpoint: account.SosEndpoint,
		DNSAPIEndpoint:     account.DNSEndpoint,
		ClientTimeout:      account.ClientTimeout,
	}

	return &out, nil
}
