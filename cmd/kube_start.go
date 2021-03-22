package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var kubeStartCmd = &cobra.Command{
	Use:   "start CLUSTER-NAME",
	Short: "Start a stopped standalone Kubernetes cluster instance",
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
			if !askQuestion(fmt.Sprintf("Are you sure you want to stop Kubernetes cluster instance %q?", vm.Name)) {
				return nil
			}
		}

		if err := startVirtualMachine(vm.Name, ""); err != nil {
			fmt.Println("failed")
			return err
		}

		return nil
	},
}

func init() {
	kubeStartCmd.Flags().BoolP("force", "f", false, cmdFlagForceHelp)
	kubeCmd.AddCommand(kubeStartCmd)
}
