package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/exoscale/egoscale"
	"github.com/exoscale/egoscale/cmd/exo/table"

	"github.com/spf13/cobra"
)

// serviceofferingCmd represents the serviceoffering command
var serviceofferingCmd = &cobra.Command{
	Use:   "serviceoffering",
	Short: "List available services offerings with details",
	Run: func(cmd *cobra.Command, args []string) {
		if err := listServiceOffering(); err != nil {
			log.Fatal(err)
		}
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

func init() {
	vmCmd.AddCommand(serviceofferingCmd)
}
