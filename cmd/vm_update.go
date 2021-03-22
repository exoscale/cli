package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var vmUpdateCmd = &cobra.Command{
	Use:               "update NAME|ID",
	Short:             "Update a Compute instance properties",
	ValidArgsFunction: completeVMNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			vmEdit egoscale.UpdateVirtualMachine
			edited bool
		)

		if len(args) != 1 {
			return cmd.Usage()
		}

		vm, err := getVirtualMachineByNameOrID(args[0])
		if err != nil {
			return err
		}
		vmEdit.ID = vm.ID

		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		if cmd.Flags().Changed("name") {
			vmEdit.Name = name
			vmEdit.DisplayName = name
			edited = true
		}

		userDataPath, err := cmd.Flags().GetString("cloud-init-file")
		if err != nil {
			return err
		}
		if userDataPath != "" {
			vmEdit.UserData, err = getUserDataFromFile(userDataPath)
			if err != nil {
				return err
			}
			edited = true
		}

		if edited {
			_, err = cs.RequestWithContext(gContext, &vmEdit)
			if err != nil {
				return fmt.Errorf("unable to update Compute instance: %s", err)
			}

			if !gQuiet {
				fmt.Println("Compute instance updated successfully")
			}

			return nil
		}

		return cmd.Usage()
	},
}

func init() {
	vmUpdateCmd.Flags().String("name", "", "display name")
	vmUpdateCmd.Flags().String("cloud-init-file", "", "path to a cloud-init user data file")
	vmCmd.AddCommand(vmUpdateCmd)
}
