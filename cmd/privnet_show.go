package cmd

import (
	"os"
	"strings"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// showCmd represents the show command
var privnetShowCmd = &cobra.Command{
	Use:   "show <privnet name | id>",
	Short: "Show a private network details",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Usage()
		}

		network, err := getNetworkByName(args[0])
		if err != nil {
			return err
		}

		vms, err := privnetDetails(network)
		if err != nil {
			return err
		}

		dhcp := dhcpRange(*network)

		t := table.NewTable(os.Stdout)
		t.SetHeader([]string{network.Name})

		t.Append([]string{"ID", network.ID.String()})
		t.Append([]string{"Name", network.Name})
		t.Append([]string{"Description", network.DisplayText})
		t.Append([]string{"Zone", network.ZoneName})
		t.Append([]string{"DHCP IP Range", dhcp})

		if len(vms) == 0 {
			t.Append([]string{"Instances", "n/a"})
			t.Render()
			return nil
		}

		if len(vms) > 0 {
			instances := make([]string, len(vms))
			for i, vm := range vms {
				instances[i] = strings.Join([]string{vm.ID.String(), vm.Name}, " │ ")
			}
			t.Append([]string{"Instances", strings.Join(instances, "\n")})
		}

		t.Render()

		return nil
	},
}

func privnetDetails(network *egoscale.Network) ([]egoscale.VirtualMachine, error) {
	vms, err := cs.ListWithContext(gContext, &egoscale.VirtualMachine{
		ZoneID: network.ZoneID,
	})
	if err != nil {
		return nil, err
	}

	var vmsRes []egoscale.VirtualMachine
	for _, v := range vms {
		vm := v.(*egoscale.VirtualMachine)

		nic := vm.NicByNetworkID(*network.ID)
		if nic != nil {
			vmsRes = append(vmsRes, *vm)
		}
	}

	return vmsRes, nil
}

func init() {
	privnetCmd.AddCommand(privnetShowCmd)
}
