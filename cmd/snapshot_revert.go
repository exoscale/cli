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

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		tasks := make([]task, 0, len(args))
		for _, arg := range args {
			snapshot, err := getSnapshotWithNameOrID(arg)
			if err != nil {
				return err
			}

			vmName := snapshotVMName(*snapshot)
			q := fmt.Sprintf("Are you sure you want to revert %q using the snapshot: %q", vmName, snapshot.Name)
			if !force && !askQuestion(q) {
				continue
			}

			tasks = append(tasks, task{
				egoscale.RevertSnapshot{ID: snapshot.ID},
				fmt.Sprintf("Reverting snapshot %q", snapshot.Name),
			})
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
	snapshotRevertCmd.Flags().BoolP("force", "f", false, "Attempt to revert snapshot without prompting for confirmation")
}
