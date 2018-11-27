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

		query := &egoscale.Volume{
			VirtualMachineID: vm.ID,
			Type:             "ROOT",
		}

		resp, err := cs.GetWithContext(gContext, query)
		if err != nil {
			return err
		}

		createSnapshot := &egoscale.CreateSnapshot{
			VolumeID: resp.(*egoscale.Volume).ID,
		}

		message := fmt.Sprintf("Creating snapshot of %q", vm.Name)

		res, err := asyncRequest(createSnapshot, message)
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
