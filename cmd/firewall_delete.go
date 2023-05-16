package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/egoscale"
)

var firewallDeleteCmd = &cobra.Command{
	Use:     "delete NAME|ID",
	Short:   "Delete a Security Group",
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
			sg, err := getSecurityGroupByNameOrID(arg)
			if err != nil {
				return err
			}

			q := fmt.Sprintf("Are you sure you want to delete the Security Group %q?", sg.Name)
			if !force && !askQuestion(q) {
				continue
			}

			cmd := &egoscale.DeleteSecurityGroup{ID: sg.ID}
			tasks = append(tasks, task{
				cmd,
				fmt.Sprintf("Deleting Security Group %q", sg.Name),
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
	firewallDeleteCmd.Flags().BoolP("force", "f", false, cmdFlagForceHelp)
	firewallCmd.AddCommand(firewallDeleteCmd)
}
