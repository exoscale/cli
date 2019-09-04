package cmd

import (
	"fmt"
	"os"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
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
	serviceOffering, err := cs.ListWithContext(gContext, &egoscale.ServiceOffering{})
	if err != nil {
		return err
	}

	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "cpu", "ram"})

	for _, soff := range serviceOffering {
		f := soff.(*egoscale.ServiceOffering)

		ram := ""
		if f.Memory > 1000 {
			ram = fmt.Sprintf("%d GB", f.Memory>>10)
		} else if f.Memory < 1000 {
			ram = fmt.Sprintf("%d MB", f.Memory)
		}

		table.Append([]string{f.ID.String(), f.Name, fmt.Sprintf("%d× %d MHz", f.CPUNumber, f.CPUSpeed), ram})
	}

	table.Render()

	return nil

}

func getServiceOfferingByName(name string) (*egoscale.ServiceOffering, error) {
	so := &egoscale.ServiceOffering{}

	id, err := egoscale.ParseUUID(name)
	if err != nil {
		so.Name = name
	} else {
		so.ID = id
	}

	resp, err := cs.GetWithContext(gContext, so)
	if err != nil {
		return nil, err
	}

	return resp.(*egoscale.ServiceOffering), nil
}

func init() {
	vmCmd.AddCommand(serviceofferingCmd)
}
