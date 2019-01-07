package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// kubeEnvCmd represents the env command
var kubeIpCmd = &cobra.Command{
	Use:   "ip <cluster name>",
	Short: "Print a standalone Kubernetes cluster's IP address",
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

		fmt.Println(vm.IP().String())

		return nil
	},
}

func init() {
	kubeCmd.AddCommand(kubeIpCmd)
}
