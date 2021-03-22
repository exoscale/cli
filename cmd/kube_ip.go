package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var kubeIPCmd = &cobra.Command{
	Use:   "ip CLUSTER-NAME",
	Short: "Print a standalone Kubernetes cluster's IP address",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}
		clusterName := args[0]

		vm, err := getKubeVM(clusterName)
		if err != nil {
			return err
		}

		fmt.Println(vm.IP().String())

		return nil
	},
}

func init() {
	kubeCmd.AddCommand(kubeIPCmd)
}
