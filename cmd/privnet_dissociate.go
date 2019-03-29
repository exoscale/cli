package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// dissociateCmd represents the dissociate command
var dissociateCmd = &cobra.Command{
	Use:     "dissociate <privnet name | id> <vm name | vm id>+",
	Short:   "Dissociate a private network from instance(s)",
	Aliases: gDissociateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		network, err := getNetworkByName(args[0])
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
				if !askQuestion(fmt.Sprintf("sure you want to dissociate %q private network", args[0])) {
					continue
				}
			}

			tasks = append(tasks, task{
				cmd,
				fmt.Sprintf("dissociate %q privnet from %q", network.Name, vm.Name),
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
		return nil, fmt.Errorf("no nics found for network %q", privnet.ID)
	}

	return &egoscale.RemoveNicFromVirtualMachine{
		NicID:            nic.ID,
		VirtualMachineID: vm.ID,
	}, nil
}

func init() {
	dissociateCmd.Flags().BoolP("force", "f", false, "Attempt to dissociate private network without prompting for confirmation")
	privnetCmd.AddCommand(dissociateCmd)
}
