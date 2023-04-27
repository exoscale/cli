package cmd

import (
	"fmt"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var vmResizeCmd = &cobra.Command{
	Use:               "resize NAME|ID",
	Short:             "Resize a Compute instance disk",
	ValidArgsFunction: completeVMNames,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		return cmdCheckRequiredFlags(cmd, []string{"disk"})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		diskValue, err := cmd.Flags().GetInt64("disk")
		if err != nil {
			return err
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		tasks := make([]task, 0, len(args))
		for _, v := range args {
			if !force {
				if !askQuestion(fmt.Sprintf("Are you sure you want to resize Compute instance disk %q?", v)) {
					continue
				}
			}

			task, err := resizeVirtualMachine(v, diskValue)
			if err != nil {
				return err
			}

			tasks = append(tasks, *task)
		}

		taskResponses := asyncTasks(tasks)
		errors := filterErrors(taskResponses)
		if len(errors) > 0 {
			return errors[0]
		}

		return nil
	},
}

func resizeVirtualMachine(vmName string, diskValue int64) (*task, error) {
	vm, err := getVirtualMachineByNameOrID(vmName)
	if err != nil {
		return nil, err
	}

	state := (string)(egoscale.VirtualMachineStopped)
	if vm.State != state {
		return nil, fmt.Errorf("this operation is not permitted while your Compute instance is running; stop it before issuing that command again")
	}

	resp, err := globalstate.EgoscaleClient.GetWithContext(gContext, egoscale.Volume{
		VirtualMachineID: vm.ID,
		Type:             "ROOT",
	})
	if err != nil {
		return nil, err
	}
	resizeVolume := &egoscale.ResizeVolume{
		ID:   resp.(*egoscale.Volume).ID,
		Size: diskValue,
	}

	return &task{
		resizeVolume,
		fmt.Sprintf("Resizing %q ", vm.Name),
	}, err
}

func init() {
	vmCmd.AddCommand(vmResizeCmd)
	vmResizeCmd.Flags().Int64P("disk", "d", 0, "Disk size in GB")
	vmResizeCmd.Flags().BoolP("force", "f", false, cmdFlagForceHelp)
}
