package cmd

import (
	"fmt"
	"os"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// rebootCmd represents the reboot command
var vmRebootCmd = &cobra.Command{
	Use:   "reboot <vm name> [vm name] ...",
	Short: "Reboot virtual machine instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		tasks := make([]task, 0, len(args))
		for _, v := range args {
			if !force {
				if !askQuestion(fmt.Sprintf("Are you sure you want to reboot virtual machine %q?", v)) {
					continue
				}
			}

			vm, err := getVirtualMachineByNameOrID(v)
			if err != nil {
				return err
			}

			state := (string)(egoscale.VirtualMachineRunning)
			if vm.State != state {
				fmt.Fprintf(os.Stderr, "%q is not in a %s state, got %s\n", vm.Name, state, vm.State)
				continue
			}

			tasks = append(tasks, task{
				&egoscale.RebootVirtualMachine{ID: vm.ID},
				fmt.Sprintf("Rebooting %q ", vm.Name),
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

// rebootVirtualMachine reboot a virtual machine instance Async
func rebootVirtualMachine(vmName string) error {
	vm, err := getVirtualMachineByNameOrID(vmName)
	if err != nil {
		return err
	}

	state := (string)(egoscale.VirtualMachineRunning)
	if vm.State != state {
		return fmt.Errorf("%q is not in a %s state, got %s", vmName, state, vm.State)
	}

	_, err = asyncRequest(&egoscale.RebootVirtualMachine{ID: vm.ID}, fmt.Sprintf("Rebooting %q ", vm.Name))
	return err
}

func init() {
	vmRebootCmd.Flags().BoolP("force", "f", false, "Attempt to reboot virtual machine without prompting for confirmation")
	vmCmd.AddCommand(vmRebootCmd)
}
