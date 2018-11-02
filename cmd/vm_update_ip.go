package cmd

import (
	"errors"
	"fmt"
	"net"
	"os"
	"text/tabwriter"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var vmUpdateIPCmd = &cobra.Command{
	Use:   "updateip <vm name|id> <network name|id> <ip address>",
	Short: "Update the static DHCP lease of an instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return cmd.Usage()
		}
		vm, err := getVMWithNameOrID(args[0])
		if err != nil {
			return err
		}
		network, err := getNetworkByName(args[1])
		if err != nil {
			return err
		}
		nic := vm.NicByNetworkID(*network.ID)
		if nic == nil {
			return fmt.Errorf("the virtual machine %s is not associated to the network %s", vm.DisplayName, network.Name)
		}

		newVM, err := updateNicIP(nic.ID, args[2])
		if err != nil {
			return err
		}

		return showVMWithNics(newVM)
	},
}

func updateNicIP(nicID *egoscale.UUID, newIP string) (*egoscale.VirtualMachine, error) {
	IP := net.ParseIP(newIP)
	if IP == nil {
		return nil, errors.New("invalid IP address")
	}

	req := &egoscale.UpdateVMNicIP{
		IPAddress: IP,
		NicID:     nicID,
	}

	resp, err := asyncRequest(req, fmt.Sprintf("updating the static lease of NIC %q", nicID))
	if err != nil {
		return nil, err
	}

	return resp.(*egoscale.VirtualMachine), nil
}

func showVMWithNics(vm *egoscale.VirtualMachine) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)
	fmt.Fprintf(w, "\nInstance ID:\t%s\n", vm.ID) // nolint: errcheck
	fmt.Fprintf(w, "Name:\t%s\n", vm.DisplayName) // nolint: errcheck
	if vm.DisplayName != vm.Name {
		fmt.Fprintf(w, "Hostname:\t%s\n", vm.Name) // nolint: errcheck
	}

	fmt.Fprintf(w, "Network Interfaces:\n") // nolint: errcheck
	defaultNic := vm.DefaultNic()
	if defaultNic != nil {
		fmt.Fprintf(w, "-\tNetwork:\tPublic\n")                               // nolint: errcheck
		fmt.Fprintf(w, " \tIP Address:\t%s\n", defaultNic.IPAddress.String()) // nolint: errcheck
	}
	for _, nic := range vm.Nic {
		if nic.IsDefault {
		} else {
			network := &egoscale.Network{ID: nic.NetworkID}
			if err := cs.GetWithContext(gContext, network); err != nil {
				return err
			}

			networkName := network.Name
			if networkName == "" {
				networkName = network.ID.String()
			}

			fmt.Fprintf(w, "-\tNetwork:\t%s\n", networkName)
			if network.Name == "" {
				fmt.Fprintf(w, " \tID:\t%s\n", network.ID.String())
			}

			fmt.Fprintf(w, " \tIP Address:\t%s\n", nicIP(nic))
		}
	}

	fmt.Fprintln(w) // nolint: errcheck
	return w.Flush()
}

func init() {
	vmCmd.AddCommand(vmUpdateIPCmd)
}
