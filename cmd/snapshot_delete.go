package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var snapshotDeleteCmd = &cobra.Command{
	Use:   "delete <name | id>+",
	Short: "A brief description of your command",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		tasks := make([]task, 0, len(args))

		for _, arg := range args {
			volume, err := getSnapshotWithNameOrID(arg)
			if err != nil {
				return err
			}
			t := task{egoscale.DeleteSnapshot{ID: volume.ID}, fmt.Sprintf("Delete Snapshot %q", volume.Name)}
			tasks = append(tasks, t)
		}

		taskResponses := asyncTasks(tasks)
		errs := filterErrors(taskResponses)
		if len(errs) > 0 {
			return errs[0]
		}

		return nil
	},
}

func init() {
	snapshotCmd.AddCommand(snapshotDeleteCmd)
}
