package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var vmStopCmd = &cobra.Command{
	Use:   "stop <vm name> [vm name] ...",
	Short: "Stop virtual machine instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		errs := []error{}
		for _, v := range args {
			if err := stopVirtualMachine(v); err != nil {
				errs = append(errs, fmt.Errorf("could not stop %q: %s", v, err))
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

// stopVirtualMachine stop a virtual machine instance
func stopVirtualMachine(vmName string) error {
	vm, err := getVMWithNameOrID(vmName)
	if err != nil {
		return err
	}

	if vm.State != (string)(egoscale.VirtualMachineRunning) {
		return fmt.Errorf("virtual machine is not running")
	}

	fmt.Printf("Stopping %q ", vm.Name)
	var errorReq error
	cs.AsyncRequestWithContext(gContext, &egoscale.StopVirtualMachine{ID: vm.ID}, func(jobResult *egoscale.AsyncJobResult, err error) bool {

		fmt.Print(".")

		if err != nil {
			errorReq = err
			return false
		}

		if jobResult.JobStatus == egoscale.Success {
			fmt.Println(" success.")
			return false
		}

		return true
	})

	if errorReq != nil {
		fmt.Println(" failure!")
	}

	return errorReq
}

func init() {
	vmCmd.AddCommand(vmStopCmd)
}
