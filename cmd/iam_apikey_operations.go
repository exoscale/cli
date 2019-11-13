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

// apiKeyShowCmd represents the API key showing command
var apiKeyOperationsCmd = &cobra.Command{
	Use:   "operations [Filter ...]",
	Short: "List Operations",
	Long: fmt.Sprintf(`This command lists all Operations for an API key.
	
	Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&apiKeyOperationsItemOutput{}), ", ")),
	Aliases: gShowAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		return output(listApiKeyOperations(args))
	},
}

func listApiKeyOperations(filters []string) (outputter, error) {
	resp, err := cs.RequestWithContext(gContext, &egoscale.ListAPIKeyOperations{})
	if err != nil {
		return nil, err
	}

	opes := resp.(*egoscale.ListAPIKeyOperationsResponse)

	out := apiKeyOperationsItemOutput{}

	for _, s := range opes.Operations {
		st := strings.ToLower(s)

		keep := true
		if len(filters) > 0 {
			keep = false

			for _, filter := range filters {
				substr := strings.ToLower(filter)
				if strings.Contains(st, substr) {
					keep = true
					break
				}
			}
		}

		if !keep {
			continue
		}

		switch true {
		case strings.Contains(s, "compute"):
			out.Compute = append(out.Compute, s)
		case strings.Contains(s, "dns"):
			out.DNS = append(out.DNS, s)
		case strings.Contains(s, "iam"):
			out.IAM = append(out.IAM, s)
		case strings.Contains(s, "sos"):
			out.SOS = append(out.SOS, s)
		}
	}

	return &out, nil
}

func init() {
	apiKeyCmd.AddCommand(apiKeyOperationsCmd)
}
