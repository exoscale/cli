package cmd

import (
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// snapshotCmd represents the snapshot command
var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Snapshots allow you to save the volume of your machine in its current state",
}

func getSnapshotWithNameOrID(name string) (*egoscale.Snapshot, error) {
	snapshot := &egoscale.Snapshot{}

	id, err := egoscale.ParseUUID(name)
	if err != nil {
		snapshot.Name = name
	} else {
		snapshot.ID = id
	}

	if err := cs.GetWithContext(gContext, snapshot); err != nil {
		return nil, err
	}

	return snapshot, nil
}

func init() {
	RootCmd.AddCommand(snapshotCmd)
}
