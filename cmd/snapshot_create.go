package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

func init() {
	snapshotCmd.AddCommand(&cobra.Command{
		Use:   "create INSTANCE-NAME|ID",
		Short: "Create a snapshot of a Compute instance volume",
		Long: fmt.Sprintf(`This command creates a snapshot of a Compute instance volume.

Supported output template annotations: %s`,
			strings.Join(output.OutputterTemplateAnnotations(&snapshotShowOutput{}), ", ")),
		Aliases: gCreateAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Usage()
			}

			return output(createSnapshot(args[0]))
		},
	})
}

func createSnapshot(vmID string) (outputter, error) {
	vm, err := getVirtualMachineByNameOrID(vmID)
	if err != nil {
		return nil, err
	}

	resp, err := cs.GetWithContext(gContext, &egoscale.Volume{
		VirtualMachineID: vm.ID,
		Type:             "ROOT",
	})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Compute instance volume: %v", err)
	}

	createSnapshotReq := &egoscale.CreateSnapshot{VolumeID: resp.(*egoscale.Volume).ID}
	res, err := asyncRequest(createSnapshotReq, fmt.Sprintf("Creating snapshot of %q", vm.Name))
	if err != nil {
		return nil, err
	}

	if !gQuiet {
		return showSnapshot(res.(*egoscale.Snapshot))
	}

	return nil, nil
}
