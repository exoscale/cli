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
		userDataCompress, err := cmd.Flags().GetBool("cloud-init-compress")
		if err != nil {
			return err
		}
		if userDataPath != "" {
			vmEdit.UserData, err = getUserDataFromFile(userDataPath, userDataCompress)
			if err != nil {
				return err
			}
			edited = true
		}

		if edited {
			if _, err = cs.RequestWithContext(gContext, &vmEdit); err != nil {
				return fmt.Errorf("unable to update Compute instance: %s", err)
			}
		}

		if cmd.Flags().Changed("reverse-dns") {
			reverseDNS, err := cmd.Flags().GetString("reverse-dns")
			if err != nil {
				return err
			}

			if _, err = cs.RequestWithContext(gContext, &egoscale.UpdateReverseDNSForVirtualMachine{
				ID:         vm.ID,
				DomainName: reverseDNS,
			}); err != nil {
				return fmt.Errorf("unable to update Compute instance reverse DNS: %s", err)
			}
		}

		if !gQuiet {
			return printOutput(showVM(vm.ID.String()))
		}

		return nil
	},
}

func init() {
	vmUpdateCmd.Flags().String("name", "", "instance display name")
	vmUpdateCmd.Flags().String("cloud-init-file", "", "path to a cloud-init user data file")
	vmUpdateCmd.Flags().BoolP("cloud-init-compress", "", false, "compress instance cloud-init user data")
	vmUpdateCmd.Flags().String("reverse-dns", "", "instance reverse DNS record")
	vmCmd.AddCommand(vmUpdateCmd)
}
