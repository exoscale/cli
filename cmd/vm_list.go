package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/output"
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

func (o *vmListOutput) toJSON()  { output.JSON(o) }
func (o *vmListOutput) toText()  { output.Text(o) }
func (o *vmListOutput) toTable() { output.Table(o) }
func (o *vmListOutput) names() []string {
	names := make([]string, len(*o))
	for i, item := range *o {
		names[i] = item.Name
	}

	return names
}

func init() {
	vmCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List Compute instances",
		Long: fmt.Sprintf(`This command lists existing Compute instances.

Supported output template annotations: %s`,
			strings.Join(output.OutputterTemplateAnnotations(&vmListOutput{}), ", ")),
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
