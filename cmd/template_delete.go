package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

func init() {
	templateCmd.AddCommand(templateDeleteCmd)
}

// templateDeleteCmd represents the delete command
var templateDeleteCmd = &cobra.Command{
	Use:   "delete <template id>+",
	Short: "Delete a template",
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
			id, err := egoscale.ParseUUID(arg)
			if err != nil {
				return err
			}
			cmd := egoscale.DeleteTemplate{
				ID: id,
			}
			if err != nil {
				return err
			}

			if !force {
				if !askQuestion(fmt.Sprintf("Are you sure you want to delete template %q?", arg)) {
					continue
				}
			}

			tasks = append(tasks, task{
				cmd,
				fmt.Sprintf("deleting template %q", cmd.ID.String()),
			})
		}

		resps := asyncTasks(tasks)
		errs := filterErrors(resps)
		if len(errs) > 0 {
			return errs[0]
		}
		return nil
	},
}

func init() {
	templateDeleteCmd.Flags().BoolP("force", "f", false, "Attempt to remove a template without prompting for confirmation")
	templateCmd.AddCommand(templateDeleteCmd)
}
