package cmd

import (
	"errors"
	"fmt"
	"strings"

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

		errs := []error{}
		for _, v := range args {
			if !force {
				if !askQuestion(fmt.Sprintf("Are you sure you want to reboot virtual machine %q?", v)) {
					return nil
				}
			}

			if err := rebootVirtualMachine(v); err != nil {
				errs = append(errs, fmt.Errorf("could not reboot %q: %s", v, err))
			}
		}

		if len(errs) == 1 {
			return errs[0]
		}
		if len(errs) > 1 {
			var b strings.Builder
			for _, err := range errs {
				if _, e := fmt.Fprintln(&b, err); e != nil {
					return e
				}
			}
			return errors.New(b.String())
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
