package cmd

import (
	"fmt"
	"os"
	"strings"

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
			table.SetHeader([]string{"VM", "State", "Created On", "Size", "ID"})
			res, err := cs.ListWithContext(gContext, egoscale.Snapshot{})
			if err != nil {
				return err
			}

			var vmNameTmp string
			for _, s := range res {
				snapshot := s.(*egoscale.Snapshot)
				vmName := snapshotVMName(*snapshot)
				if vmName == vmNameTmp {
					vmName = ""
				}
				table.Append([]string{vmName, snapshot.State, snapshot.Created, fmt.Sprintf("%v", humanize.IBytes(uint64(snapshot.Size))), snapshot.ID.String()})
				if vmName != "" {
					vmNameTmp = vmName
				}
			}

			table.Render()

			return nil
		}

		table.SetHeader([]string{"VM", "State", "Created On", "Size", "ID"})

		for _, arg := range args {
			vm, err := getVirtualMachineByNameOrID(arg)
			if err != nil {
				return err
			}

			volume := &egoscale.Volume{
				VirtualMachineID: vm.ID,
				Type:             "ROOT",
			}

			resp, err := cs.GetWithContext(gContext, volume)
			if err != nil {
				return err
			}

			snapshots, err := cs.ListWithContext(gContext, egoscale.Snapshot{
				VolumeID: resp.(*egoscale.Volume).ID,
			})

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

// snapshotVMName returns the instance name based on the snapshot name
func snapshotVMName(snapshot egoscale.Snapshot) string {
	names := strings.SplitN(snapshot.Name, "_"+snapshot.VolumeName+"_", 2)
	return names[0]
}
