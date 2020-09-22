package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var affinitygroupDeleteCmd = &cobra.Command{
	Use:     "delete <name | id>+",
	Short:   "Delete affinity group",
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
			cmd, err := prepareDeleteAffinityGroup(arg)
			if err != nil {
				return err
			}

			if !force {
				if !askQuestion(fmt.Sprintf("sure you want to delete %q affinity group", arg)) {
					continue
				}
			}

			tasks = append(tasks, task{
				cmd,
				fmt.Sprintf("deleting %q affinity group", cmd.Name),
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

func prepareDeleteAffinityGroup(name string) (*egoscale.DeleteAffinityGroup, error) {
	aff, err := getAffinityGroupByNameOrID(name)
	if err != nil {
		return nil, err
	}

	return &egoscale.DeleteAffinityGroup{ID: aff.ID, Name: aff.Name}, nil
}

func init() {
	affinitygroupDeleteCmd.Flags().BoolP("force", "f", false, "Attempt to remove affinity group without prompting for confirmation")
	affinitygroupCmd.AddCommand(affinitygroupDeleteCmd)
}
