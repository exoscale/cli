package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/exoscale/egoscale/cmd/exo/table"

	"github.com/spf13/cobra"
)

// serviceofferingCmd represents the serviceoffering command
var serviceofferingCmd = &cobra.Command{
	Use:   "serviceoffering",
	Short: "List available services offerings with details",
	RunE: func(cmd *cobra.Command, args []string) error {
		return listServiceOffering()
	},
}

func listServiceOffering() error {
	serviceOffering, err := cs.List(&egoscale.ServiceOffering{})
	if err != nil {
		return err
	}

	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"Name", "cpu", "ram"})

	for _, soff := range serviceOffering {
		f := soff.(*egoscale.ServiceOffering)

		ram := ""
		if f.Memory > 1000 {
			ram = fmt.Sprintf("%d GB", f.Memory>>10)
		} else if f.Memory < 1000 {
			ram = fmt.Sprintf("%d MB", f.Memory)
		}

		table.Append([]string{f.Name, fmt.Sprintf("%d× %d MHz", f.CPUNumber, f.CPUSpeed), ram})
	}

	table.Render()

	return nil

}

func getServiceOfferingByName(cs *egoscale.Client, servOffering string) (*egoscale.ServiceOffering, error) {
	servOReq := &egoscale.ServiceOffering{}

	servOffs, err := cs.List(servOReq)
	if err != nil {
		return nil, err
	}

	for _, servoff := range servOffs {
		r := servoff.(*egoscale.ServiceOffering)
		if strings.Compare(strings.ToLower(servOffering), strings.ToLower(r.Name)) == 0 {
			return r, nil
		}
		if strings.Compare(servOffering, r.ID) == 0 {
			return r, nil
		}
	}
	return nil, fmt.Errorf("service offering %q was not found", servOffering)
}

func init() {
	vmCmd.AddCommand(serviceofferingCmd)
}
