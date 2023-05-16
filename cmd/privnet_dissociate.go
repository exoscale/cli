package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/egoscale"
)

var dissociateCmd = &cobra.Command{
	Use:     "dissociate NETWORK-NAME|ID INSTANCE-NAME|ID",
	Short:   "Dissociate a Private Network from a Compute instance",
	Aliases: gDissociateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		network, err := getNetwork(args[0], nil)
		if err != nil {
			return err
		}

		tasks := make([]task, 0, len(args[1:]))
		for _, arg := range args[1:] {
			vm, err := getVirtualMachineByNameOrID(arg)
			if err != nil {
				return err
			}

			cmd, err := prepareDissociatePrivNet(network, vm)
			if err != nil {
				return err
			}

			if !force {
				if !askQuestion(fmt.Sprintf("Are you sure you want to dissociate Private Network %q?", args[0])) {
					continue
				}
			}

			tasks = append(tasks, task{
				cmd,
				fmt.Sprintf("Dissociating Private Network %q from %q", network.Name, vm.Name),
			})
		}

		resps := asyncTasks(tasks)
		errs := filterErrors(resps)
		if len(errs) > 0 {
			return errs[0]
		}
		return nil
	},
}

func prepareDissociatePrivNet(privnet *egoscale.Network, vm *egoscale.VirtualMachine) (*egoscale.RemoveNicFromVirtualMachine, error) {
	nic := vm.NicByNetworkID(*privnet.ID)
	if nic == nil {
		return nil, fmt.Errorf("no NIC found for Private Network %q", privnet.ID)
	}

	return &egoscale.RemoveNicFromVirtualMachine{
		NicID:            nic.ID,
		VirtualMachineID: vm.ID,
	}, nil
}

func init() {
	dissociateCmd.Flags().BoolP("force", "f", false, cmdFlagForceHelp)
	privnetCmd.AddCommand(dissociateCmd)
}
