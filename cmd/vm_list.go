package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type vmListItemOutput struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Size      string `json:"size"`
	Zone      string `json:"zone"`
	State     string `json:"state"`
	IPAddress string `json:"ip_address"`
}

type vmListOutput []vmListItemOutput

func (o *vmListOutput) toJSON()  { outputJSON(o) }
func (o *vmListOutput) toText()  { outputText(o) }
func (o *vmListOutput) toTable() { outputTable(o) }

func init() {
	vmCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all the virtual machines instances",
		Long: fmt.Sprintf(`This command lists existing Compute instances.

Supported output template annotations: %s`,
			strings.Join(outputterTemplateAnnotations(&vmListOutput{}), ", ")),
		Aliases: gListAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return cmd.Usage()
			}

			return output(listVMs())
		},
	})
}

func listVMs() (outputter, error) {
	vm := &egoscale.VirtualMachine{}
	vms, err := cs.ListWithContext(gContext, vm)
	if err != nil {
		return nil, err
	}

	out := vmListOutput{}

	for _, key := range vms {
		vm := key.(*egoscale.VirtualMachine)

		out = append(out, vmListItemOutput{
			ID:        vm.ID.String(),
			Name:      vm.Name,
			Size:      vm.ServiceOfferingName,
			Zone:      vm.ZoneName,
			State:     vm.State,
			IPAddress: vm.IP().String(),
		})
	}

	return &out, nil
}
