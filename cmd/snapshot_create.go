package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

func init() {
	snapshotCmd.AddCommand(&cobra.Command{
		Use:   "create <vm name | vm id>",
		Short: "Create a snapshot of a Compute instance volume",
		Long: fmt.Sprintf(`This command creates a snapshot of a Compute instance volume.

Supported output template annotations: %s`,
			strings.Join(outputterTemplateAnnotations(&snapshotShowOutput{}), ", ")),
		Aliases: gCreateAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Usage()
			}

			return createSnapshot(args[0])
		},
	})
}

func createSnapshot(vmID string) error {
	vm, err := getVirtualMachineByNameOrID(vmID)
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

	res, err := asyncRequest(createSnapshot, fmt.Sprintf("Creating snapshot of %q", vm.Name))
	if err != nil {
		return err
	}

	if !gQuiet {
		return output(showSnapshot(res.(*egoscale.Snapshot)))
	}

	return nil
}
