package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// revertCmd represents the revert command
var snapshotRevertCmd = &cobra.Command{
	Use:   "revert <name | id>+",
	Short: "Revert a snapshot to an instance volume",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		tasks := make([]task, 0, len(args))
		for _, arg := range args {
			snapshot, err := getSnapshotWithNameOrID(arg)
			if err != nil {
				return err
			}
			task := task{egoscale.RevertSnapshot{ID: snapshot.ID}, fmt.Sprintf("Revert snapshot %q", snapshot.Name)}
			tasks = append(tasks, task)
		}

		results := asyncTasks(tasks)
		errs := filterErrors(results)
		if len(errs) > 0 {
			return errs[0]
		}

		return nil
	},
}

func init() {
	snapshotCmd.AddCommand(snapshotRevertCmd)
}
