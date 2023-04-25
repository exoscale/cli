package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/output"
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
func (o *affinityGroupShowOutput) toJSON()      { output.JSON(o) }
func (o *affinityGroupShowOutput) toText()      { output.Text(o) }
func (o *affinityGroupShowOutput) toTable()     { output.Table(o) }

func init() {
	affinitygroupCmd.AddCommand(&cobra.Command{
		Use:   "show NAME|ID",
		Short: "Show affinity group details",
		Long: fmt.Sprintf(`This command shows an Anti-Affinity Group details.

Supported output template annotations: %s`,
			strings.Join(output.OutputterTemplateAnnotations(&affinityGroupShowOutput{}), ", ")),
		Aliases: gShowAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Usage()
			}

			ag, err := getAntiAffinityGroupByNameOrID(args[0])
			if err != nil {
				return err
			}

			return printOutput(showAffinityGroup(ag))
		},
	})
}

func showAffinityGroup(ag *egoscale.AffinityGroup) (output.Outputter, error) {
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
