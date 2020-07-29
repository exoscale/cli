package cmd

import (
	"fmt"
	"os"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var vmStartCmd = &cobra.Command{
	Use:               "start <vm name | id>+",
	Short:             "Start virtual machine",
	ValidArgsFunction: completeVMNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		rescueProfile, err := cmd.Flags().GetString("rescue-profile")
		if err != nil {
			return err
		}

		tasks := make([]task, 0, len(args))
		for _, v := range args {
			vm, err := getVirtualMachineByNameOrID(v)
			if err != nil {
				return err
			}
			state := (string)(egoscale.VirtualMachineStopped)
			if vm.State != state {
				fmt.Fprintf(os.Stderr, "%q is not in a %s state, got: %s\n", vm.Name, state, vm.State)
				continue
			}
			tasks = append(tasks, task{
				egoscale.StartVirtualMachine{ID: vm.ID, RescueProfile: rescueProfile},
				fmt.Sprintf("Starting %q ", vm.Name),
			})

		}

		taskResponses := asyncTasks(tasks)
		errors := filterErrors(taskResponses)
		if len(errors) > 0 {
			return errors[0]
		}

		return nil
	},
}

// startVirtualMachine start a virtual machine instance Async
func startVirtualMachine(vmName string, vmRescueProfile string) error {
	vm, err := getVirtualMachineByNameOrID(vmName)
	if err != nil {
		return err
	}

	state := (string)(egoscale.VirtualMachineStopped)
	if vm.State != state {
		return fmt.Errorf("%q is not in a %s state, got: %s", vmName, state, vm.State)
	}

	_, err = asyncRequest(&egoscale.StartVirtualMachine{ID: vm.ID, RescueProfile: vmRescueProfile},
		fmt.Sprintf("Starting %q ", vm.Name))
	return err
}

func init() {
	vmStartCmd.Flags().StringP("rescue-profile", "", "", "option rescue profile when starting VM")
	vmCmd.AddCommand(vmStartCmd)
}
