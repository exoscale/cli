package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var privnetDeleteCmd = &cobra.Command{
	Use:     "delete NAME|ID",
	Short:   "Delete a Private Network",
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
				if !askQuestion(fmt.Sprintf("Are you sure you want to delete Private Network %q?", arg)) {
					continue
				}
			}

			tasks = append(tasks, task{
				cmd,
				fmt.Sprintf("Deleting Private Network %q", cmd.ID.String()),
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
	privnetDeleteCmd.Flags().BoolP("force", "f", false, cmdFlagForceHelp)
	privnetCmd.AddCommand(privnetDeleteCmd)
}
