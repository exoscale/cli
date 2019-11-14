package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type apiKeyOperationsItemOutput struct {
	Compute []string `json:"compute,omitempty"`
	DNS     []string `json:"dns,omitempty"`
	IAM     []string `json:"iam,omitempty"`
	SOS     []string `json:"sos,omitempty"`
}

func (o *apiKeyOperationsItemOutput) toJSON()  { outputJSON(o) }
func (o *apiKeyOperationsItemOutput) toText()  { outputText(o) }
func (o *apiKeyOperationsItemOutput) toTable() { outputTable(o) }

// apiKeyOperationsCmd represents the supported operations listing command for an API key
var apiKeyOperationsCmd = &cobra.Command{
	Use:   "operations [filter ...]",
	Short: "List Operations",
	Long: fmt.Sprintf(`This command lists all suported Operations for an API key.
	
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

		switch true {
		case strings.Contains(result, "compute"):
			out.Compute = append(out.Compute, result)
		case strings.Contains(result, "dns"):
			out.DNS = append(out.DNS, result)
		case strings.Contains(result, "iam"):
			out.IAM = append(out.IAM, result)
		case strings.Contains(result, "sos"):
			out.SOS = append(out.SOS, result)
		}
	}

	return &out, nil
}

func init() {
	apiKeyCmd.AddCommand(apiKeyOperationsCmd)
}
