package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var privnetDeleteCmd = &cobra.Command{
	Use:     "delete <name | id>+",
	Short:   "Delete private network",
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
			cmd, err := deletePrivnet(arg)
			if err != nil {
				return err
			}

			if !force {
				if !askQuestion(fmt.Sprintf("sure you want to delete %q private network", arg)) {
					continue
				}
			}

			tasks = append(tasks, task{
				cmd,
				fmt.Sprintf("delete %q privnet", cmd.ID.String()),
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

func deletePrivnet(name string) (*egoscale.DeleteNetwork, error) {
	addrReq := &egoscale.DeleteNetwork{}
	var err error
	network, err := getNetwork(name, nil)
	if err != nil {
		return nil, err
	}
	addrReq.ID = network.ID
	return addrReq, nil
}

func init() {
	privnetDeleteCmd.Flags().BoolP("force", "f", false, "Attempt to remove private network without prompting for confirmation")
	privnetCmd.AddCommand(privnetDeleteCmd)
}
