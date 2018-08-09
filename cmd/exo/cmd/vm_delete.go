package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var vmDeleteCmd = &cobra.Command{
	Use:     "delete <name | id> [name | id] ...",
	Short:   "Delete virtual machine instance(s)",
	Aliases: gDeleteAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		for _, arg := range args {
			if err := deleteVM(arg, force); err != nil {
				_, err = fmt.Fprintf(os.Stderr, err.Error())
				if err != nil {
					return err
				}
			}
		}
		return nil
	},
}

func deleteVM(name string, force bool) error {
	vm, err := getVMWithNameOrID(name)
	if err != nil {
		return err
	}

	if !force {
		if !askQuestion(fmt.Sprintf("sure you want to delete %q virtual machine", vm.Name)) {
			return nil
		}

	}

	if _, err := asyncRequest(&egoscale.DestroyVirtualMachine{ID: vm.ID}, fmt.Sprintf("Destroying %q ", vm.Name)); err != nil {
		return err
	}

	folder := path.Join(gConfigFolder, "instances", vm.ID)

	if _, err := os.Stat(folder); !os.IsNotExist(err) {
		if err := os.RemoveAll(folder); err != nil {
			return err
		}
	}

	fmt.Println(vm.ID)

	return nil
}

func init() {
	vmDeleteCmd.Flags().BoolP("force", "f", false, "Attempt to remove vitual machine without prompting for confirmation")
	vmCmd.AddCommand(vmDeleteCmd)
}
