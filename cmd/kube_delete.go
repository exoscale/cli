package cmd

import (
	"fmt"
	"os"

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

		clusterInstance, err := loadKubeData(clusterName, "instance")
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("%q: no such cluster", clusterName)
			}
			return err
		}

		vm, err := getVirtualMachineByNameOrID(string(clusterInstance))
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

		fmt.Printf("Destroying cluster instance... ")

		if err := cs.DeleteWithContext(gContext, vm); err != nil {
			fmt.Println("instance deletion failed")
			return err
		}

		fmt.Println("done")

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
