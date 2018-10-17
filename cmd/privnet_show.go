package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

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

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)
		dhcp := dhcpRange(network.StartIP, network.EndIP, network.Netmask)

		fmt.Fprintf(w, "Network:\t%s\n", network.Name)            // nolint: errcheck
		fmt.Fprintf(w, "Description:\t%s\n", network.DisplayText) // nolint: errcheck
		fmt.Fprintf(w, "Zone:\t%s\n", network.ZoneName)           // nolint: errcheck
		fmt.Fprintf(w, "IP Range:\t%s\n", dhcp)                   // nolint: errcheck
		fmt.Fprintf(w, "# Instances:\t%d\n", len(vms))            // nolint: errcheck
		w.Flush()                                                 // nolint: errcheck

		table := table.NewTable(os.Stdout)
		table.SetHeader([]string{"ID", "Name", "Public IP", "Internal IP"})

		if len(vms) > 0 {
			for _, vm := range vms {

				privateNic := vm.NicByNetworkID(*network.ID)
				table.Append([]string{
					vm.ID.String(),
					vm.Name,
					vm.IP().String(),
					nicIP(privateNic),
				})
			}
			table.Render()
		}
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
