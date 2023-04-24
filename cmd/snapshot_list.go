package cmd

import (
	"fmt"
	"strings"

	humanize "github.com/dustin/go-humanize"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/egoscale"

	"github.com/spf13/cobra"
)

type snapshotListItemOutput struct {
	ID       string `json:"id"`
	Date     string `json:"date"`
	Instance string `json:"instance"`
	State    string `json:"state"`
	Size     string `json:"size"`
}

type snapshotListOutput []snapshotListItemOutput

func (o *snapshotListOutput) toJSON()  { output.JSON(o) }
func (o *snapshotListOutput) toText()  { output.Text(o) }
func (o *snapshotListOutput) toTable() { output.Table(o) }

func init() {
	snapshotCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List snapshots",
		Long: fmt.Sprintf(`This command lists existing Compute instance disk snapshots.

Supported output template annotations: %s`,
			strings.Join(output.OutputterTemplateAnnotations(&snapshotListOutput{}), ", ")),
		Aliases: gListAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			return output(listSnapshots(args))
		},
	})
}

func listSnapshots(instances []string) (outputter, error) {
	out := snapshotListOutput{}

	if len(instances) == 0 {
		snapshots, err := cs.ListWithContext(gContext, egoscale.Snapshot{})
		if err != nil {
			return nil, err
		}

		for _, s := range snapshots {
			snapshot := s.(*egoscale.Snapshot)
			instance := snapshotVMName(*snapshot)

			out = append(out, snapshotListItemOutput{
				ID:       snapshot.ID.String(),
				Instance: instance,
				Date:     snapshot.Created,
				State:    snapshot.State,
				Size:     humanize.IBytes(uint64(snapshot.Size)),
			})
		}

		return &out, nil
	}

	for _, i := range instances {
		instance, err := getVirtualMachineByNameOrID(i)
		if err != nil {
			return nil, err
		}

		volume, err := cs.GetWithContext(gContext, &egoscale.Volume{
			VirtualMachineID: instance.ID,
			Type:             "ROOT",
		})
		if err != nil {
			return nil, err
		}

		snapshots, err := cs.ListWithContext(gContext, egoscale.Snapshot{VolumeID: volume.(*egoscale.Volume).ID})
		if err != nil {
			return nil, err
		}

		for _, s := range snapshots {
			snapshot := s.(*egoscale.Snapshot)

			out = append(out, snapshotListItemOutput{
				ID:       snapshot.ID.String(),
				Instance: instance.Name,
				Date:     snapshot.Created,
				State:    snapshot.State,
				Size:     humanize.IBytes(uint64(snapshot.Size)),
			})
		}
	}

	return &out, nil
}

// snapshotVMName returns the instance name based on the snapshot name.
func snapshotVMName(snapshot egoscale.Snapshot) string {
	names := strings.SplitN(snapshot.Name, "_"+snapshot.VolumeName+"_", 2)
	return names[0]
}
