package cmd

import (
	"fmt"
	"os"

	"github.com/exoscale/egoscale"

	"github.com/spf13/cobra"
)

var vmScaleCmd = &cobra.Command{
	Use:               "scale NAME|ID",
	Short:             "Scale a Compute instance (change service offering)",
	ValidArgsFunction: completeVMNames,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		return cmdCheckRequiredFlags(cmd, []string{"service-offering"})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		so, err := cmd.Flags().GetString("service-offering")
		if err != nil {
			return err
		}

		serviceoffering, err := getServiceOfferingByNameOrID(so)
		if err != nil {
			return err
		}

		tasks := make([]task, 0, len(args))
		for _, v := range args {
			vm, err := getVirtualMachineByNameOrID(v)
			if err != nil {
				return err
			}

			if vm.State != (string)(egoscale.VirtualMachineStopped) {
				fmt.Fprintf(os.Stderr, "this operation is not permitted while your Compute instance is not running; stop it before issuing that command again\n")
			}

			tasks = append(tasks, task{
				&egoscale.ScaleVirtualMachine{ID: vm.ID, ServiceOfferingID: serviceoffering.ID},
				fmt.Sprintf("Scaling %q ", vm.Name),
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
	vmCmd.AddCommand(vmScaleCmd)
	vmScaleCmd.Flags().StringP("service-offering", "o", "", serviceOfferingHelp)
}
