package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// kubeDeleteCmd represents the delete command
var kubeDeleteCmd = &cobra.Command{
	Use:   "delete <cluster name>",
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
			if !askQuestion(fmt.Sprintf("sure you want to delete %q cluster instance", vm.Name)) {
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
	kubeDeleteCmd.Flags().BoolP("force", "f", false, "Attempt to remove cluster instance without prompting for confirmation")
	kubeCmd.AddCommand(kubeDeleteCmd)
}
