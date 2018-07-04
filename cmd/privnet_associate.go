package cmd

import (
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// privnetAssociateCmd represents the associate command
var privnetAssociateCmd = &cobra.Command{
	Use:     "associate <privnet name | id> <vm name | vm id> [vm name | vm id] [...]",
	Short:   "Associate a private network to instance(s)",
	Aliases: gAssociateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}

		network, err := getNetworkIDByName(cs, args[0])
		if err != nil {
			return err
		}

		for _, vm := range args[1:] {
			resp, err := associatePrivNet(network, vm)
			if err != nil {
				return err
			}
			println(resp)
		}
		return nil
	},
}

func associatePrivNet(privnet *egoscale.Network, vmName string) (string, error) {
	vm, err := getVMWithNameOrID(cs, vmName)
	if err != nil {
		return "", err
	}

	_, err = cs.Request(&egoscale.AddNicToVirtualMachine{NetworkID: privnet.ID, VirtualMachineID: vm.ID})
	if err != nil {
		return "", err
	}

	return vm.ID, nil

}

func init() {
	privnetCmd.AddCommand(privnetAssociateCmd)
}
