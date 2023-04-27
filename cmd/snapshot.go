package cmd

import (
	"fmt"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Snapshots allow you to save the volume of a machine in its current state",
}

func getSnapshotByNameOrID(v string) (*egoscale.Snapshot, error) {
	snapshot := &egoscale.Snapshot{}

	id, err := egoscale.ParseUUID(v)
	if err != nil {
		snapshot.Name = v
	} else {
		snapshot.ID = id
	}

	resp, err := globalstate.EgoscaleClient.GetWithContext(gContext, snapshot)
	switch err {
	case nil:
		return resp.(*egoscale.Snapshot), nil

	case egoscale.ErrNotFound:
		return nil, fmt.Errorf("unknown Snapshot %q", v)

	case egoscale.ErrTooManyFound:
		return nil, fmt.Errorf("multiple Snapshots match %q", v)

	default:
		return nil, err
	}
}

func init() {
	vmCmd.AddCommand(snapshotCmd)
}
