package cmd

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var vmDeleteCmd = &cobra.Command{
	Use:   "delete <name | id>",
	Short: "Delete a virtual machine instance",
}

func vmDeleteCmdRun(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		vmDeleteCmd.Usage()
		return
	}

	if err := deleteVM(args[0]); err != nil {
		log.Fatal(err)
	}
}

func deleteVM(name string) error {
	vm, err := getVMWithNameOrID(cs, name)
	if err != nil {
		return err
	}

	var errorReq error

	req := &egoscale.DestroyVirtualMachine{ID: vm.ID}
	print("Destroying")
	cs.AsyncRequest(req, func(jobResult *egoscale.AsyncJobResult, err error) bool {

		if err != nil {
			errorReq = err
			return false
		}

		if jobResult.JobStatus == egoscale.Success {
			println("")
			return false
		}
		fmt.Printf(".")
		return true
	})

	if errorReq != nil {
		return errorReq
	}

	folder := path.Join(configFolder, "instances", vm.ID)

	if _, err := os.Stat(folder); !os.IsNotExist(err) {
		if err := os.RemoveAll(folder); err != nil {
			return err
		}
	}

	println(vm.ID)

	return nil
}

func init() {
	vmDeleteCmd.Run = vmDeleteCmdRun
	vmCmd.AddCommand(vmDeleteCmd)
}
