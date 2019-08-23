package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var vmUpdateIPCmd = &cobra.Command{
	Use:   "updateip <vm name|id> <network name|id> [flags]",
	Short: "Update the static DHCP lease of an instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return cmd.Usage()
		}
		vmName := args[0]
		netName := args[1]

		vm, err := getVirtualMachineByNameOrID(vmName)
		if err != nil {
			return err
		}
		network, err := getNetwork(netName, nil)
		if err != nil {
			return err
		}
		nic := vm.NicByNetworkID(*network.ID)
		if nic == nil {
			return fmt.Errorf("the virtual machine %s is not associated to the network %s", vm.DisplayName, network.Name)
		}

		ipAddress, err := getIPValue(cmd, "ip")
		if err != nil {
			return err
		}

		var msg string
		if ipAddress.IP != nil {
			msg = fmt.Sprintf("setting the static lease to %q, network %q: %s", vmName, netName, ipAddress.IP.String())
		} else {
			msg = fmt.Sprintf("removing the static lease from %q, network %q", vmName, netName)
		}

		req := &egoscale.UpdateVMNicIP{
			IPAddress: ipAddress.Value(),
			NicID:     nic.ID,
		}

		resp, err := asyncRequest(req, msg)
		if err != nil {
			return err
		}

		return showVMWithNics(resp.(*egoscale.VirtualMachine))
	},
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
			resp, err := cs.GetWithContext(gContext, &egoscale.Network{
				ID: nic.NetworkID,
			})

			if err != nil {
				return err
			}

			network := resp.(*egoscale.Network)
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
	ipAddress := new(ipValue)

	vmUpdateIPCmd.Flags().VarP(ipAddress, "ip", "i", "the static IP address lease, no values unsets it.")

	vmCmd.AddCommand(vmUpdateIPCmd)
}
