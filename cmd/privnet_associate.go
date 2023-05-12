package cmd

import (
	"fmt"
	"net"
	"os"
	"text/tabwriter"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var privnetAssociateCmd = &cobra.Command{
	Use:     "associate NETWORK-NAME|ID INSTANCE-NAME|ID [IP-ADDRESS]",
	Short:   "Associate a Private Network to a Compute instance",
	Aliases: gAssociateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}

		network, err := getNetwork(args[0], nil)
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)
		dhcp := dhcpRange(*network)

		fmt.Fprintf(w, "Network:\t%s\n", network.Name)            // nolint: errcheck
		fmt.Fprintf(w, "Description:\t%s\n", network.DisplayText) // nolint: errcheck
		fmt.Fprintf(w, "Zone:\t%s\n", network.ZoneName)           // nolint: errcheck
		fmt.Fprintf(w, "IP Range:\t%s\n", dhcp)                   // nolint: errcheck

		// FIXME: this implementation mixes side effects with user reporting,
		// this is not great and should be split.
		table := table.NewTable(os.Stdout)
		table.SetHeader([]string{"Compute instance", "IP Address"})
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
						nicIP(*nic),
					})

					i++

					continue
				}
			}
			nic, vm, err := associatePrivNet(network, name, nil)
			if err != nil {
				return err
			}
			table.Append([]string{
				vm.DisplayName,
				nicIP(*nic),
			})
		}
		w.Flush() // nolint: errcheck

		if !globalstate.Quiet {
			table.Render()
		}

		return nil
	},
}

func associatePrivNet(privnet *egoscale.Network, vmName string, ip net.IP) (*egoscale.Nic, *egoscale.VirtualMachine, error) {
	vm, err := getVirtualMachineByNameOrID(vmName)
	if err != nil {
		return nil, nil, err
	}

	req := &egoscale.AddNicToVirtualMachine{NetworkID: privnet.ID, VirtualMachineID: vm.ID, IPAddress: ip}
	resp, err := globalstate.EgoscaleClient.RequestWithContext(gContext, req)
	if err != nil {
		return nil, nil, err
	}

	nic := resp.(*egoscale.VirtualMachine).NicByNetworkID(*privnet.ID)
	if nic == nil {
		return nil, nil, fmt.Errorf("no NIC found for network %q", privnet.ID)
	}

	return nic, vm, nil
}

func init() {
	privnetCmd.AddCommand(privnetAssociateCmd)
}
