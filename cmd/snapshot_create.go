package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/egoscale"
)

func init() {
	snapshotCmd.AddCommand(&cobra.Command{
		Use:   "create INSTANCE-NAME|ID",
		Short: "Create a snapshot of a Compute instance volume",
		Long: fmt.Sprintf(`This command creates a snapshot of a Compute instance volume.

Supported output template annotations: %s`,
			strings.Join(output.TemplateAnnotations(&snapshotShowOutput{}), ", ")),
		Aliases: gCreateAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Usage()
			}

			return printOutput(createSnapshot(args[0]))
		},
	})
}

func createSnapshot(vmID string) (output.Outputter, error) {
	vm, err := getVirtualMachineByNameOrID(vmID)
	if err != nil {
		return nil, err
	}

	resp, err := globalstate.EgoscaleClient.GetWithContext(gContext, &egoscale.Volume{
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

	if !globalstate.Quiet {
		return showSnapshot(res.(*egoscale.Snapshot))
	}

	return nil, nil
}
