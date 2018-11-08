package cmd

import (
	"fmt"
	"log"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var snapshotCreateCmd = &cobra.Command{
	Use:     "create <vm name | vm id>",
	Short:   "Create a snapshot of an instance volume",
	Aliases: gCreateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		vm, err := getVMWithNameOrID(args[0])
		if err != nil {
			return err
		}

		volume := &egoscale.Volume{
			VirtualMachineID: vm.ID,
			Type:             "ROOT",
		}

		if err := cs.GetWithContext(gContext, volume); err != nil {
			return err
		}

		res, err := asyncRequest(&egoscale.CreateSnapshot{VolumeID: volume.ID}, fmt.Sprintf("Create Snapshot of %q", vm.Name))
		if err != nil {
			return err
		}

		result := res.(*egoscale.Snapshot)

		log.Printf("Snapshot %q of %q successfully created", result.Name, vm.Name)

		return nil
	},
}

func init() {
	snapshotCmd.AddCommand(snapshotCreateCmd)
}
