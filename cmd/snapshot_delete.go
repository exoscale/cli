package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var snapshotDeleteCmd = &cobra.Command{
	Use:     "delete <name | id>+",
	Short:   "Delete a snapshot",
	Aliases: gDeleteAlias,
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
			q := fmt.Sprintf("Are you sure you want to delete the snapshot: %q", arg)
			if !force && !askQuestion(q) {
				continue
			}
			volume, err := getSnapshotWithNameOrID(arg)
			if err != nil {
				return err
			}
			t := task{egoscale.DeleteSnapshot{ID: volume.ID}, fmt.Sprintf("Deleting snapshot %s", volume.ID.String())}
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
	snapshotDeleteCmd.Flags().BoolP("force", "f", false, "Attempt to remove snapshot without prompting for confirmation")
}
