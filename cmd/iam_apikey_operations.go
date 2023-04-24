package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type apiKeyOperationsItemOutput struct {
	Account []string `json:"account,omitempty"`
	Compute []string `json:"compute,omitempty"`
	DNS     []string `json:"dns,omitempty"`
	IAM     []string `json:"iam,omitempty"`
	SOS     []string `json:"sos,omitempty"`
}

func (o *apiKeyOperationsItemOutput) toJSON()  { output.JSON(o) }
func (o *apiKeyOperationsItemOutput) toText()  { output.Text(o) }
func (o *apiKeyOperationsItemOutput) toTable() { output.Table(o) }

var apiKeyOperationsCmd = &cobra.Command{
	Use:   "operations [FILTER]...",
	Short: "List supported API key operations",
	Long: fmt.Sprintf(`This command lists all supported operations for an API key.
Optional patterns can be provided to filter results by compute, DNS, IAM or SOS operations.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&apiKeyOperationsItemOutput{}), ", ")),
	RunE: func(cmd *cobra.Command, args []string) error {
		return output(listAPIKeyOperations(args))
	},
}

func listAPIKeyOperations(filters []string) (outputter, error) {
	resp, err := cs.RequestWithContext(gContext, &egoscale.ListAPIKeyOperations{})
	if err != nil {
		return nil, err
	}

	opes := resp.(*egoscale.ListAPIKeyOperationsResponse)

	out := apiKeyOperationsItemOutput{}

	for _, o := range opes.Operations {
		operation := strings.ToLower(o)

		result := operation
		for _, f := range filters {
			result = ""
			filter := strings.ToLower(f)
			if strings.Contains(operation, filter) {
				result = operation
				break
			}
		}

		switch {
		case strings.HasPrefix(result, "account/"):
			out.Account = append(out.Account, o)
		case strings.HasPrefix(result, "compute/"):
			out.Compute = append(out.Compute, o)
		case strings.HasPrefix(result, "dns/"):
			out.DNS = append(out.DNS, o)
		case strings.HasPrefix(result, "iam/"):
			out.IAM = append(out.IAM, o)
		case strings.HasPrefix(result, "sos/"):
			out.SOS = append(out.SOS, o)
		}
	}

	return &out, nil
}

func init() {
	apiKeyCmd.AddCommand(apiKeyOperationsCmd)
}
