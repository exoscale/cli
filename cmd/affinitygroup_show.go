package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type affinityGroupShowOutput struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Instances   []string `json:"instances"`
}

func (o *affinityGroupShowOutput) Type() string { return "Anti-Affinity Group" }
func (o *affinityGroupShowOutput) toJSON()      { outputJSON(o) }
func (o *affinityGroupShowOutput) toText()      { outputText(o) }
func (o *affinityGroupShowOutput) toTable()     { outputTable(o) }

func init() {
	affinitygroupCmd.AddCommand(&cobra.Command{
		Use:   "show <name | id>",
		Short: "Show affinity group details",
		Long: fmt.Sprintf(`This command shows an Anti-Affinity Group details.

Supported output template annotations: %s`,
			strings.Join(outputterTemplateAnnotations(&affinityGroupShowOutput{}), ", ")),
		Aliases: gShowAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Usage()
			}

			return output(showAffinityGroup(args[0]))
		},
	})
}

func showAffinityGroup(name string) (outputter, error) {
	ag, err := getAffinityGroupByName(name)
	if err != nil {
		return nil, err
	}

	out := affinityGroupShowOutput{
		ID:          ag.ID.String(),
		Name:        ag.Name,
		Description: ag.Description,
		Instances:   make([]string, len(ag.VirtualMachineIDs)),
	}

	if len(ag.VirtualMachineIDs) > 0 {
		resp, err := cs.ListWithContext(gContext, &egoscale.ListVirtualMachines{IDs: ag.VirtualMachineIDs})
		if err != nil {
			return nil, err
		}

		for i, r := range resp {
			vm := r.(*egoscale.VirtualMachine)
			out.Instances[i] = vm.Name
		}
	}

	return &out, nil
}
