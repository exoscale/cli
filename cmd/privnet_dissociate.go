package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// dissociateCmd represents the dissociate command
var dissociateCmd = &cobra.Command{
	Use:     "dissociate <privnet name | id> <vm name | vm id> [vm name | vm id] [...]",
	Short:   "Dissociate a private network from instance(s)",
	Aliases: gDissociateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}

		network, err := getNetworkIDByName(cs, args[0])
		if err != nil {
			return err
		}

		for _, vm := range args[1:] {
			resp, err := dissociatePrivNet(network, vm)
			if err != nil {
				return err
			}
			println(resp)
		}
		return nil
	},
}

func dissociatePrivNet(privnet *egoscale.Network, vmName string) (string, error) {
	vm, err := getVMWithNameOrID(vmName)
	if err != nil {
		return "", err
	}

	nic, err := containNetID(privnet, vm.Nic)
	if err != nil {
		return "", err
	}

	_, err = cs.RequestWithContext(gContext, &egoscale.RemoveNicFromVirtualMachine{NicID: nic.ID, VirtualMachineID: vm.ID})
	if err != nil {
		return "", err
	}

	return nic.ID, nil
}

func containNetID(network *egoscale.Network, vmNics []egoscale.Nic) (egoscale.Nic, error) {

	for _, nic := range vmNics {
		if nic.NetworkID == network.ID {
			return nic, nil
		}
	}
	return egoscale.Nic{}, fmt.Errorf("NIC not found")
}

func init() {
	privnetCmd.AddCommand(dissociateCmd)
}
