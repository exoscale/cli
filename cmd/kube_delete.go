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

		if err := vm.Delete(gContext, cs); err != nil {
			fmt.Println("failed")
			return err
		}

		fmt.Println("done")

		deleteKeyPair(*vm.ID)
		deleteKubeData(clusterName)

		return nil
	},
}

func init() {
	kubeDeleteCmd.Flags().BoolP("force", "f", false, "Attempt to remove cluster instance without prompting for confirmation")
	kubeCmd.AddCommand(kubeDeleteCmd)
}
