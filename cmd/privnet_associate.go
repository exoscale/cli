package cmd

import (
	"fmt"
	"net"
	"os"
	"text/tabwriter"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// privnetAssociateCmd represents the associate command
var privnetAssociateCmd = &cobra.Command{
	Use:     "associate <privnet name | id> <vm name | vm id> [<ip>] [<vm name | vm id> [<ip>]] [...]",
	Short:   "Associate a private network to instance(s)",
	Aliases: gAssociateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}

		network, err := getNetworkByName(args[0])
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)
		dhcp := dhcpRange(network.StartIP, network.EndIP, network.Netmask)

		fmt.Fprintf(w, "Network:\t%s\n", network.Name)            // nolint: errcheck
		fmt.Fprintf(w, "Description:\t%s\n", network.DisplayText) // nolint: errcheck
		fmt.Fprintf(w, "Zone:\t%s\n", network.ZoneName)           // nolint: errcheck
		fmt.Fprintf(w, "IP Range:\t%s\n", dhcp)                   // nolint: errcheck

		table := table.NewTable(os.Stdout)
		table.SetHeader([]string{"Virtual Machine", "IP Address"})
		for i := 1; i < len(args); i++ {
			name := args[i]
			if i != len(args)-1 {
				ip := net.ParseIP(args[i+1])
				if ip != nil {
					// the next param is an ip
					nic, vm, err := associatePrivNet(network, name, ip)
					if err != nil {
						return err
					}
					table.Append([]string{
						vm.DisplayName,
						nicIP(*nic)})
					i = i + 1
					continue
				}
			}
			nic, vm, err := associatePrivNet(network, name, nil)
			if err != nil {
				return err
			}
			table.Append([]string{
				vm.DisplayName,
				nicIP(*nic)})
		}
		w.Flush() // nolint: errcheck
		table.Render()
		return nil
	},
}

func associatePrivNet(privnet *egoscale.Network, vmName string, ip net.IP) (*egoscale.Nic, *egoscale.VirtualMachine, error) {
	vm, err := getVMWithNameOrID(vmName)
	if err != nil {
		return nil, nil, err
	}

	req := &egoscale.AddNicToVirtualMachine{NetworkID: privnet.ID, VirtualMachineID: vm.ID, IPAddress: ip}
	resp, err := cs.RequestWithContext(gContext, req)
	if err != nil {
		return nil, nil, err
	}

	nic := resp.(*egoscale.VirtualMachine).NicByNetworkID(*privnet.ID)
	if nic == nil {
		return nil, nil, fmt.Errorf("no nics found for network %q", privnet.ID)
	}

	return nic, vm, nil

}

func init() {
	privnetCmd.AddCommand(privnetAssociateCmd)
}
