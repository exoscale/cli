package cmd

import (
	"fmt"
	"os"

	humanize "github.com/dustin/go-humanize"
	"github.com/exoscale/egoscale"

	"github.com/exoscale/cli/table"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var snapshotListCmd = &cobra.Command{
	Use:     "list [vm name | vm id]+",
	Short:   "List snapshot",
	Aliases: gListAlias,
	RunE: func(cmd *cobra.Command, args []string) error {

		table := table.NewTable(os.Stdout)

		if len(args) == 0 {
			table.SetHeader([]string{"State", "Create On", "Size", "ID"})
			res, err := cs.ListWithContext(gContext, egoscale.Snapshot{})
			if err != nil {
				return err
			}

			for _, s := range res {
				snapshot := s.(*egoscale.Snapshot)
				table.Append([]string{snapshot.State, snapshot.Created, fmt.Sprintf("%v", humanize.IBytes(uint64(snapshot.Size))), snapshot.ID.String()})
			}

			table.Render()

			return nil
		}

		table.SetHeader([]string{"VM", "State", "Create On", "Size", "ID"})

		for _, arg := range args {
			vm, err := getVMWithNameOrID(arg)
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

			snapshots, err := cs.ListWithContext(gContext, egoscale.Snapshot{VolumeID: volume.ID})
			if err != nil {
				return err
			}

			for _, s := range snapshots {
				snapshot := s.(*egoscale.Snapshot)

				table.Append([]string{vm.Name, snapshot.State, snapshot.Created, fmt.Sprintf("%v", humanize.IBytes(uint64(snapshot.Size))), snapshot.ID.String()})
				vm.Name = ""
			}

		}

		table.Render()

		return nil
	},
}

func init() {
	snapshotCmd.AddCommand(snapshotListCmd)
}
