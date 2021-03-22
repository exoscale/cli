package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var kubeDeleteCmd = &cobra.Command{
	Use:   "delete NAME",
	Short: "Delete a standalone Kubernetes cluster instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}
		clusterName := args[0]

		vm, err := getKubeVM(clusterName)
		if err != nil {
			return err
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		if !force {
			if !askQuestion(fmt.Sprintf("Are you sure you want to delete Kubernetes cluster instance %q?", vm.Name)) {
				return nil
			}
		}

		resps := asyncTasks([]task{{
			&egoscale.DestroyVirtualMachine{ID: vm.ID},
			fmt.Sprintf("Destroying cluster instance %q", clusterName),
		}})
		errs := filterErrors(resps)
		if len(errs) > 0 {
			return errs[0]
		}

		if err := deleteKeyPair(*vm.ID); err != nil {
			return err
		}

		return deleteKubeData(clusterName)
	},
}

func init() {
	kubeDeleteCmd.Flags().BoolP("force", "f", false, cmdFlagForceHelp)
	kubeCmd.AddCommand(kubeDeleteCmd)
}
