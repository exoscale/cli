package cmd

import (
	"fmt"
	"os"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var vmStartCmd = &cobra.Command{
	Use:               "start NAME|ID",
	Short:             "Start a stopped Compute instance",
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

func init() {
	vmStartCmd.Flags().StringP("rescue-profile", "", "", "option rescue profile when starting VM")
	vmCmd.AddCommand(vmStartCmd)
}
