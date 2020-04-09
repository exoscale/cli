package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type LimitsItemOutput struct {
	Resource  string `json:"resource"`
	Used      int    `json:"used"`
	Available int    `json:"available"`
	Max       int    `json:"max"`
}

type LimitsOutput []LimitsItemOutput

func (o *LimitsOutput) toJSON()  { outputJSON(o) }
func (o *LimitsOutput) toText()  { outputText(o) }
func (o *LimitsOutput) toTable() { outputTable(o) }

var limitsCmd = &cobra.Command{
	Use:   "limits",
	Short: "Current account limits",
	Long: fmt.Sprintf(`This command lists the safety limits currently enforced on your account.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&LimitsOutput{}), ", ")),
	RunE: func(cmd *cobra.Command, args []string) error {
		return output(listLimits())
	},
}

func listLimits() (outputter, error) {
	limits, err := cs.ListWithContext(gContext, &egoscale.ResourceLimit{})
	if err != nil {
		return nil, err
	}

	display := map[string]string{
		"user_vm":           "Instances",
		"snapshot":          "Snapshots",
		"template":          "Templates",
		"public_elastic_ip": "IP Addresses",
		"network":           "Private Networks",
	}

	out := LimitsOutput{}

	for _, key := range limits {
		limit := key.(*egoscale.ResourceLimit)

		if display[limit.ResourceTypeName] != "" {
			rUsed, err := fetchUsedResources(limit.ResourceTypeName)
			if err != nil {
				return nil, err
			}

			out = append(out, LimitsItemOutput{
				Resource:  display[limit.ResourceTypeName],
				Max:       int(limit.Max),
				Available: (int(limit.Max) - rUsed),
				Used:      rUsed,
			})
		}
	}

	return &out, nil
}

func fetchUsedResources(rType string) (int, error) {
	var rUsed int

	switch rType {
	case "user_vm":
		instances, err := cs.ListWithContext(gContext, &egoscale.VirtualMachine{})
		if err != nil {
			return 0, err
		}

		rUsed = len(instances)
	case "snapshot":
		snapshots, err := cs.ListWithContext(gContext, &egoscale.Snapshot{})
		if err != nil {
			return 0, err
		}

		rUsed = len(snapshots)
	case "template":
		templates, err := cs.ListWithContext(gContext, &egoscale.Template{})
		if err != nil {
			return 0, err
		}

		rUsed = len(templates)
	case "network":
		networks, err := cs.ListWithContext(gContext, &egoscale.Network{})
		if err != nil {
			return 0, err
		}

		rUsed = len(networks)
	case "public_elastic_ip":
		eips, err := cs.ListWithContext(gContext, &egoscale.IPAddress{IsElastic: true})
		if err != nil {
			return 0, err
		}

		rUsed = len(eips)
	}

	return rUsed, nil
}

func init() {
	RootCmd.AddCommand(limitsCmd)
}
